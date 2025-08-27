// Package export creates new handler for export with dependency injection.
package export

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/md5" // #nosec G501
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	libkafka "wikimedia-enterprise/services/snapshots/libraries/kafka"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"
	"wikimedia-enterprise/services/snapshots/submodules/config"
	"wikimedia-enterprise/services/snapshots/submodules/log"
	"wikimedia-enterprise/services/snapshots/submodules/schema"

	nbf "github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/klauspost/pgzip"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/confluentinc/confluent-kafka-go/kafka"

	"go.uber.org/dig"
)

var (
	// "ns": namespace. "namespace" already exists as a label.
	// "type": one of "snapshots", "structured-snapshots" or "batches".
	commonLabels             = []string{"project", "ns", "type"}
	snapshotSizeBytesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snapshot_size_bytes",
		Help: "The size of generated snapshots",
	}, append(commonLabels, "has_errors", "partitions"))
	snapshotArticlesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snapshot_articles",
		Help: "The number of articles in generated snapshots",
	}, append(commonLabels, "has_errors", "partitions"))
	snapshotErrorsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snapshot_errors",
		Help: "The number of errors encountered in generated snapshots",
	}, commonLabels)
	snapshotTimeSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "snapshot_time_seconds_v2",
		Help:    "The time it took to generate a given snapshot",
		Buckets: prometheus.ExponentialBuckets(1, 2, 14), // 14 buckets, 1s to ~8192s (~2.2h)
	}, append(commonLabels, "has_errors", "partitions"))
	snapshotMessagesProcessedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snapshot_messages_processed",
		Help: "The number of messages processed when generating snapshots, updated in real time",
	}, commonLabels)
)

// ErrEndMessage represents an error when we reached end of messages.
var ErrEndMessage = errors.New("end message")

const (
	numExportWorkers = 4
	maxBatchSize     = 1000
)

type buffer struct {
	name string
	data []byte
}

// newBuffer will create a new buffer.
func newBuffer(name string, count int) *buffer {
	return &buffer{
		name: fmt.Sprintf("%s_%d.ndjson", name, count),
		data: make([]byte, 0),
	}
}

// copyBuffer will copy the buffer.
func copyBuffer(name string, data []byte) *buffer {
	dup := make([]byte, len(data))
	copy(dup, data)

	return &buffer{
		name: name,
		data: dup,
	}
}

// len will return the length of the data.
func (b *buffer) len() int {
	return len(b.data)
}

// Counter is a struct that contains the counter.
type Counter struct {
	Itr int
	mtx sync.Mutex
}

// get will return the next value of the counter.
func (c *Counter) get() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.Itr++

	return c.Itr
}

// HashReader is a reader that will read from pipe and write to hash.
type HashReader struct {
	*nio.PipeReader
	hsh hash.Hash
}

// Read will read from pipe and write to hash.
func (r *HashReader) Read(pld []byte) (int, error) {
	if r.hsh == nil {
		r.hsh = md5.New() // #nosec G401
	}

	nth, err := r.PipeReader.Read(pld)

	if err != nil {
		return nth, err
	}

	_, err = r.hsh.Write(pld[:nth])
	return nth, err
}

// Sum will return hash sum.
func (r *HashReader) Sum() string {
	return fmt.Sprintf("%x", r.hsh.Sum(nil))
}

// Handler export handler execution struct with dependency injection.
type Handler struct {
	dig.In
	Env      *env.Environment
	S3       s3tracerproxy.S3TracerProxy
	Cfg      config.API
	Pool     libkafka.ConsumerGetter
	Stream   schema.Unmarshaler
	Uploader uploader.Uploader
}

// ChunkOptions is a struct that contains all the options for chunking.
type ChunkOptions struct {
	Ctx     context.Context
	Request *pb.ExportRequest
	Key     string
	Counter *Counter
}

// Chunks is a struct that contains all the chunks.
type Chunks struct {
	data []string
	mu   sync.Mutex
}

// NewChunks will create new chunks.
func NewChunks() *Chunks {
	return &Chunks{
		data: make([]string, 0),
		mu:   sync.Mutex{},
	}
}

// Export will create new snapshot by namespace and project identifier.
func (h *Handler) Export(ctx context.Context, req *pb.ExportRequest) (*pb.ExportResponse, error) {
	runtime.GC()
	defer runtime.GC()

	if len(req.GetPrefix()) == 0 {
		req.Prefix = h.Env.Prefix
	}

	tpc, err := h.Env.Topics.
		GetName(req.GetProject(), int(req.GetNamespace()))

	if err != nil {
		return nil, err
	}

	labelValues := []string{req.GetProject(), fmt.Sprintf("%d", req.GetNamespace()), req.GetPrefix()}

	idr := fmt.Sprintf("%s_namespace_%d", req.GetProject(), req.GetNamespace())
	idrField := log.Any("idr", idr)
	str := time.Now().UTC()
	exportErrors := make(chan error, numExportWorkers)
	bfs := make(chan *buffer, 1)
	pbf := nbf.New(int64(h.Env.PipeBufferSize))
	prr, pwr := nio.Pipe(pbf)
	hrr := &HashReader{PipeReader: prr}

	// Chunking initialize
	numChunkWorkers := h.Env.ChunkWorkers
	chunkErrors := make(chan error, numChunkWorkers)
	cuf := make(chan *buffer, 1)
	var opt ChunkOptions
	chunks := NewChunks()

	// Create many go routines that can receive chunks from the  cuf channel and process them
	if req.EnableChunking {
		opt = ChunkOptions{
			Ctx:     ctx,
			Request: req,
			Key:     idr,
			Counter: &Counter{Itr: -1},
		}

		h.CleanupChunks(ctx, idr)

		log.Info("finished cleaning up existing chunks", idrField)

		for i := 0; i < numChunkWorkers; i++ {
			go func() {
				for buf := range cuf {
					chunkField := log.Any("chunk", buf.name)
					log.Info("started chunking", chunkField, idrField)

					// Write to chunk file (Tar -> Gzip -> Hash -> upload -> S3)
					cnm, err := h.HandleChunk(buf.data, &opt)
					if err != nil {
						chunkErrors <- err
						log.Error("failed chunking", chunkField, idrField)
						return
					}

					chunks.mu.Lock()
					chunks.data = append(chunks.data, *cnm)
					chunks.mu.Unlock()

					log.Info("completed chunking", chunkField, idrField)
				}

				chunkErrors <- nil
			}()
		}
	}

	go func() {
		log.Info("starting gzip go routine", idrField)

		gzw := pgzip.NewWriter(pwr)

		if err := gzw.SetConcurrency(1<<20, runtime.NumCPU()*2); err != nil {
			exportErrors <- err
			return
		}

		trw := tar.NewWriter(gzw)

		for buf := range bfs {
			log.Info("processing buffer", idrField, log.Any("buffer", buf.name), log.Any("length", buf.len()))

			if buf.len() == 0 {
				continue
			}

			if req.EnableChunking {
				cuf <- copyBuffer(buf.name, buf.data) // copy the buffer
			}

			hdr := &tar.Header{
				Name:    buf.name,
				Size:    int64(buf.len()),
				Mode:    0766,
				ModTime: time.Now().UTC(),
			}

			if err := trw.WriteHeader(hdr); err != nil {
				exportErrors <- err
				return
			}

			if _, err := trw.Write(buf.data); err != nil {
				exportErrors <- err
				return
			}

			if err := trw.Flush(); err != nil {
				exportErrors <- err
				return
			}

			buf.data = nil // manually freeing the space from buffer
		}

		close(cuf)
		_ = trw.Close()
		_ = gzw.Close()
		_ = pwr.Close()

		exportErrors <- nil
		log.Info("gzip go routine has finished", idrField)
	}()

	bir := getSinceName(req, idr)
	filename := fmt.Sprintf("%s/%s.tar.gz", req.GetPrefix(), bir)
	metadataFilename := fmt.Sprintf("%s/%s.json", req.GetPrefix(), bir)

	go func() {
		log.Info("starting upload go routine", idrField)
		uin := &s3manager.UploadInput{
			Body:    hrr,
			Bucket:  aws.String(h.Env.AWSBucket),
			Key:     aws.String(filename),
			Tagging: aws.String(fmt.Sprintf("type=%s", req.Prefix)),
		}
		opt := func(upl *s3manager.Uploader) {
			upl.PartSize = int64(h.Env.UploadPartSize)
			upl.Concurrency = h.Env.UploadConcurrency
		}

		_, err := h.Uploader.Upload(uin, opt)
		exportErrors <- err
		log.Info("upload go routine has finished", idrField)
	}()

	res := new(pb.ExportResponse)
	mgs := make(chan *kafka.Message, h.Env.MessagesChannelCap)
	ctr := &Counter{Itr: -1}

	go func() {
		log.Info("starting buffers go routine", idrField)

		buf := newBuffer(idr, ctr.get()) // create a new empty buffer

		for msg := range mgs {
			res.Total++
			snapshotMessagesProcessedCounter.WithLabelValues(labelValues...).Inc()

			var dta []byte
			var err error

			unmarshal := func(msgValue []byte, target interface{}) error {
				if err := h.Stream.Unmarshal(ctx, msgValue, target); err != nil {
					log.Error("avro unmarshal error", idrField, log.Any("partition", msg.TopicPartition.Partition), log.Any("offset", msg.TopicPartition.Offset), log.Any("error", err))
					res.Errors++
					return err
				}
				return nil
			}

			switch req.GetType() {
			case schema.KeyTypeStructured:
				std := new(schema.AvroStructured)
				if err = unmarshal(msg.Value, std); err != nil {
					continue
				}

				if excludeEvent(std.Event, req.ExcludeEvents) {
					continue
				}

				var jst *schema.Structured
				jst, err = std.ToJsonStruct()
				if err != nil {
					log.Error("json marshal error", idrField, log.Any("partition", msg.TopicPartition.Partition), log.Any("offset", msg.TopicPartition.Offset), log.Any("error", err))
					res.Errors++
					continue
				}

				dta, err = json.Marshal(jst)
			default:
				art := new(schema.Article)
				if err = unmarshal(msg.Value, art); err != nil {
					continue
				}

				if excludeEvent(art.Event, req.ExcludeEvents) {
					continue
				}

				dta, err = json.Marshal(art)
			}

			// Skip if the data is empty, example `{}`
			if len(dta) <= 2 {
				continue
			}

			if err != nil {
				log.Error("json marshal error", idrField, log.Any("partition", msg.TopicPartition.Partition), log.Any("offset", msg.TopicPartition.Offset), log.Any("error", err))
				res.Errors++
				continue
			}

			if buf.len()+len(dta) > h.Env.BufferSize {
				bfs <- buf                      //signal that we have more buffer to compress
				buf = newBuffer(idr, ctr.get()) // create a new empty buffer
			}

			buf.data = append(buf.data, append(dta, []byte("\n")...)...)
		}

		bfs <- buf //signal that we have more buffer to compress

		close(bfs)
		exportErrors <- nil
		log.Info("finishing buffers go routine", idrField)
	}()

	pts := h.Cfg.GetPartitions(req.GetProject(), int(req.GetNamespace()))
	go func() {
		log.Info("starting consumers go routine", idrField)
		cid := fmt.Sprintf("%s_%s", req.GetPrefix(), idr)
		erp := make(chan error, len(pts))

		// Create many go routines that consume messages from the kafka topic and push them to the mgs channel
		for i := 0; i < len(pts); i++ {
			go func(pid int) {
				csr, err := h.Pool.GetConsumer(cid)

				if err != nil {
					erp <- err
					return
				}

				defer csr.Close()
				var lastIncluded kafka.Offset
				cbk := func(msg *kafka.Message) error {
					if msg.Timestamp.After(str) {
						log.Info("finishing with the ErrEndMessage", idrField, log.Any("partition", pid), log.Any("last_offset", lastIncluded))
						return libkafka.ErrReadAllEndMessage
					}

					lastIncluded = msg.TopicPartition.Offset
					mgs <- msg
					return nil
				}

				err = csr.ReadAll(ctx, int(req.GetSince()), tpc, pid, idr, cbk)
				log.Info("finishing consumer go routine", idrField, log.Any("partition", pid), log.Any("last_offset", lastIncluded))
				erp <- err
			}(pts[i])
		}

		var erm error

		// wait for all consumers to finish
		for i := 0; i < len(pts); i++ {
			if err := <-erp; err != nil {
				erm = err
			}
		}

		close(mgs)
		close(erp)
		exportErrors <- erm
		log.Info("finishing consumers go routine", idrField)
	}()

	// wait for all exportErrors (main export) and chunkErrors (chunk export) channels to finish, these are errors from the top-level go-routines
	for i := 0; i < numExportWorkers; i++ {
		if err := <-exportErrors; err != nil {
			return nil, err
		}
	}

	close(exportErrors)

	if req.EnableChunking {
		for i := 0; i < h.Env.ChunkWorkers; i++ {
			if err := <-chunkErrors; err != nil {
				return nil, err
			}
		}
		close(chunkErrors)
	}

	hin := &s3.HeadObjectInput{
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(filename),
	}

	hop, err := h.S3.HeadObjectWithContext(ctx, hin)

	if err != nil {
		return nil, err
	}

	sizeString := fmt.Sprintf("%.3f", float64(*hop.ContentLength)/1000000)
	sze, err := strconv.ParseFloat(sizeString, 64)

	if err != nil {
		return nil, err
	}

	log.Info("wrote snapshot", idrField, log.Any("file", filename), log.Any("size_mb", sizeString), log.Any("articles", res.Total), log.Any("errors", res.Errors))
	hasErr := strconv.FormatBool(res.Errors > 0)
	numPts := fmt.Sprintf("%d", len(pts))
	snapshotSizeBytesCounter.WithLabelValues(append(labelValues, hasErr, numPts)...).Add(float64(*hop.ContentLength))
	snapshotArticlesCounter.WithLabelValues(append(labelValues, hasErr, numPts)...).Add(float64(res.Total))
	snapshotErrorsCounter.WithLabelValues(labelValues...).Add(float64(res.Errors))
	snapshotTimeSeconds.WithLabelValues(append(labelValues, hasErr, numPts)...).Observe(float64(time.Since(str).Seconds()))

	dateModified := time.Now().UTC()
	snp := &schema.Snapshot{
		Identifier:   idr,
		Version:      hrr.Sum(),
		DateModified: &dateModified,
		Namespace: &schema.Namespace{
			Identifier: int(req.GetNamespace()),
		},
		IsPartOf: &schema.Project{
			Identifier: req.GetProject(),
		},
		InLanguage: &schema.Language{
			Identifier: req.GetLanguage(),
		},
		Size: &schema.Size{
			Value:    sze,
			UnitText: "MB",
		},
		Chunks: chunks.data,
	}

	dta, err := json.Marshal(snp)

	if err != nil {
		return nil, err
	}

	pin := &s3.PutObjectInput{
		Bucket:  aws.String(h.Env.AWSBucket),
		Key:     aws.String(metadataFilename),
		Body:    bytes.NewReader(dta),
		Tagging: aws.String(fmt.Sprintf("type=%s", req.Prefix)),
	}

	if _, err := h.S3.PutObjectWithContext(ctx, pin); err != nil {
		return nil, err
	}

	return res, nil
}

// handleChunk tars a bytes buffer and uploads as chunks/{snapshot_identifier}/{chunk_identifier}.tar.gz e.g., chunks/enwiki_namespace_0/chunk_2.tar.gz
// Also uploads metadata as chunks/{snapshot_identifier}/{chunk_identifier}.json e.g., chunks/enwiki_namespace_0/chunk_2.json
// It appends to a chunks slice by adding the chunks processed, if successful. Returns an error, if not.
func (h *Handler) HandleChunk(buf []byte, opts *ChunkOptions) (*string, error) {
	cid := fmt.Sprintf("chunk_%d", opts.Counter.get())

	var compressedBuffer bytes.Buffer

	gzipWriter := pgzip.NewWriter(&compressedBuffer)
	if err := gzipWriter.SetConcurrency(1<<20, runtime.NumCPU()*2); err != nil {
		return nil, fmt.Errorf("error setting gzip concurrency for chunk %s of %s: %w", cid, opts.Key, err)
	}

	tarWriter := tar.NewWriter(gzipWriter)

	header := &tar.Header{
		Name:    fmt.Sprintf("%s.ndjson", cid),
		Size:    int64(len(buf)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("error writing tar header for chunk %s of %s: %w", cid, opts.Key, err)
	}

	if _, err := tarWriter.Write(buf); err != nil {
		return nil, fmt.Errorf("error writing data to tar for chunk %s of %s: %w", cid, opts.Key, err)
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("error closing tar writer for chunk %s of %s: %w", cid, opts.Key, err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip writer for chunk %s of %s: %w", cid, opts.Key, err)
	}

	hash := md5.New() // #nosec G401
	if _, err := hash.Write(compressedBuffer.Bytes()); err != nil {
		return nil, fmt.Errorf("error creating MD5 hash for chunk %s of %s: %w", cid, opts.Key, err)
	}
	md5sum := fmt.Sprintf("%x", hash.Sum(nil))

	pin := &s3.PutObjectInput{
		Body:    bytes.NewReader(compressedBuffer.Bytes()),
		Bucket:  aws.String(h.Env.AWSBucket),
		Key:     aws.String(fmt.Sprintf("%s/%s/%s.tar.gz", "chunks", opts.Key, cid)),
		Tagging: aws.String(fmt.Sprintf("type=%s_chunks", opts.Request.Prefix)),
	}

	if _, err := h.S3.PutObjectWithContext(opts.Ctx, pin); err != nil {
		return nil, fmt.Errorf("error uploading chunk %s of %s: %w", cid, opts.Key, err)
	}

	dateModified := time.Now().UTC()
	sze, err := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(compressedBuffer.Len())/1000000), 64)

	if err != nil {
		return nil, fmt.Errorf("error calculating size for chunk %s of %s: %w", cid, opts.Key, err)
	}

	cnm := fmt.Sprintf("%s_%s", opts.Key, cid)
	cnk := &schema.Snapshot{
		Identifier:   cnm,
		Version:      md5sum,
		DateModified: &dateModified,
		Namespace: &schema.Namespace{
			Identifier: int(opts.Request.GetNamespace()),
		},
		IsPartOf: &schema.Project{
			Identifier: opts.Request.GetProject(),
		},
		InLanguage: &schema.Language{
			Identifier: opts.Request.GetLanguage(),
		},
		Size: &schema.Size{
			Value:    sze,
			UnitText: "MB",
		},
	}

	dta, err := json.Marshal(cnk)
	if err != nil {
		return nil, fmt.Errorf("error marshaling chunk metadata for chunk %s of %s: %w", cid, opts.Key, err)
	}

	metadataPin := &s3.PutObjectInput{
		Bucket:  aws.String(h.Env.AWSBucket),
		Key:     aws.String(fmt.Sprintf("%s/%s/%s.json", "chunks", opts.Key, cid)),
		Body:    bytes.NewReader(dta),
		Tagging: aws.String(fmt.Sprintf("type=%s_chunks", opts.Request.Prefix)),
	}

	if _, err := h.S3.PutObjectWithContext(opts.Ctx, metadataPin); err != nil {
		return nil, fmt.Errorf("error uploading chunk metadata for chunk %s of %s: %w", cid, opts.Key, err)
	}

	return &cnm, nil
}

// excludeEvent checks if the structured content event should be excluded.
func excludeEvent(event *schema.Event, excludeEvents []string) bool {
	if event != nil && len(excludeEvents) > 0 {
		for _, ee := range excludeEvents {
			if event.Type == ee {
				return true
			}
		}
	}

	return false
}

// getSinceName returns the batches key, if Since is set in request parameter
func getSinceName(req *pb.ExportRequest, name string) string {
	if req.GetSince() > 0 && req.GetPrefix() == "batches" {
		if req.GetEnableNonCumulativeBatches() {
			// Include the hour in the path.
			return fmt.Sprintf(
				"%s/%s",
				time.Unix(0, req.GetSince()*int64(time.Millisecond)).UTC().Format("2006-01-02/15"),
				name,
			)
		}

		return fmt.Sprintf(
			"%s/%s",
			time.Unix(0, req.GetSince()*int64(time.Millisecond)).UTC().Format("2006-01-02"),
			name,
		)
	}

	return name
}

// CleanupChunks deletes chunks from S3, ignoring keys that contain "group_1"
func (h *Handler) CleanupChunks(ctx context.Context, key string) {
	prefix := fmt.Sprintf("chunks/%s/", key)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(h.Env.AWSBucket),
		Prefix: aws.String(prefix),
	}

	var objectsToDelete []*s3.ObjectIdentifier

	err := h.S3.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			if strings.Contains(*obj.Key, fmt.Sprintf("_%s.", h.Env.FreeTierGroup)) {
				continue
			}

			if strings.HasSuffix(*obj.Key, ".tar.gz") || strings.HasSuffix(*obj.Key, ".json") {
				objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
					Key: obj.Key,
				})
			}
		}
		return !lastPage
	})

	if err != nil {
		log.Error("error listing objects for cleanup", log.Any("error", err))
		return
	}

	if len(objectsToDelete) == 0 {
		return
	}

	// Delete in batches of 1000 which is the maximum AWS allows for batch delete
	for i := 0; i < len(objectsToDelete); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(h.Env.AWSBucket),
			Delete: &s3.Delete{
				Objects: batch,
				Quiet:   aws.Bool(true),
			},
		}

		if _, err := h.S3.DeleteObjectsWithContext(ctx, deleteInput); err != nil {
			log.Error(
				"error deleting chunk batch",
				log.Any("error", err),
				log.Any("chunk", key),
			)
			continue
		}
	}
}
