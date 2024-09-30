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
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	libkafka "wikimedia-enterprise/services/snapshots/libraries/kafka"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"

	nbf "github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/klauspost/pgzip"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/confluentinc/confluent-kafka-go/kafka"

	"go.uber.org/dig"
)

// ErrEndMessage represents an error when we reached end of messages.
var ErrEndMessage = errors.New("end message")

const numExportWorkers = 4

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

// counter is a struct that contains the counter.
type counter struct {
	itr int
	mtx sync.Mutex
}

// get will return the next value of the counter.
func (c *counter) get() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.itr++

	return c.itr
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
	Counter *counter
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

	idr := fmt.Sprintf("%s_namespace_%d", req.GetProject(), req.GetNamespace())
	str := time.Now().UTC()
	ers := make(chan error, numExportWorkers)
	bfs := make(chan *buffer, 1)
	pbf := nbf.New(int64(h.Env.PipeBufferSize))
	prr, pwr := nio.Pipe(pbf)
	hrr := &HashReader{PipeReader: prr}

	// Chunking initialize
	ncg := h.Env.ChunkWorkers
	ecg := make(chan error, ncg)
	cuf := make(chan *buffer, 1)
	var opt ChunkOptions
	chunks := NewChunks()

	// Create many go routines that can receive chunks from the  cuf channel and process them
	if req.EnableChunking {
		opt = ChunkOptions{
			Ctx:     ctx,
			Request: req,
			Key:     idr,
			Counter: &counter{itr: -1},
		}

		for i := 0; i < ncg; i++ {
			go func() {
				for buf := range cuf {
					log.Printf("started chunking for id: `%s`\n", buf.name)

					// Write to chunk file (Tar -> Gzip -> Hash -> upload -> S3)
					cnm, err := h.handleChunk(buf.data, &opt)
					if err != nil {
						ecg <- err
						log.Printf("failed chunking for id: `%s`\n", buf.name)
						return
					}

					chunks.mu.Lock()
					chunks.data = append(chunks.data, *cnm)
					chunks.mu.Unlock()

					log.Printf("completed chunking for id: `%s`\n", buf.name)
				}

				ecg <- nil
			}()
		}
	}

	go func() {
		log.Printf("starting gzip go routine for id: `%s`\n", idr)

		gzw := pgzip.NewWriter(pwr)

		if err := gzw.SetConcurrency(1<<20, runtime.NumCPU()*2); err != nil {
			ers <- err
			return
		}

		trw := tar.NewWriter(gzw)

		for buf := range bfs {
			log.Printf("processing buffer for id: `%s` with the name: `%s` and length: `%d`\n", idr, buf.name, buf.len())

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
				ers <- err
				return
			}

			if _, err := trw.Write(buf.data); err != nil {
				ers <- err
				return
			}

			if err := trw.Flush(); err != nil {
				ers <- err
				return
			}

			buf.data = nil // manually freeing the space from buffer
		}

		close(cuf)
		_ = trw.Close()
		_ = gzw.Close()
		_ = pwr.Close()
		//close(ecg)

		ers <- nil
		log.Printf("gzip go routine has finished for id: `%s`\n", idr)
	}()

	bir := getSinceName(req, h.Env.Prefix, idr)

	go func() {
		log.Printf("starting upload go routine for id: `%s`\n", idr)
		uin := &s3manager.UploadInput{
			Body:    hrr,
			Bucket:  aws.String(h.Env.AWSBucket),
			Key:     aws.String(fmt.Sprintf("%s/%s.tar.gz", req.GetPrefix(), bir)),
			Tagging: aws.String(fmt.Sprintf("type=%s", req.Prefix)),
		}
		opt := func(upl *s3manager.Uploader) {
			upl.PartSize = int64(h.Env.UploadPartSize)
			upl.Concurrency = h.Env.UploadConcurrency
		}

		_, err := h.Uploader.Upload(uin, opt)
		ers <- err
		log.Printf("upload go routine has finished for id: `%s`\n", idr)
	}()

	res := new(pb.ExportResponse)
	mgs := make(chan *kafka.Message, h.Env.MessagesChannelCap)
	ctr := &counter{itr: -1}

	go func() {
		log.Printf("starting buffers go routine for id: `%s`\n", idr)

		buf := newBuffer(idr, ctr.get()) // create a new empty buffer

		for msg := range mgs {
			res.Total++

			var dta []byte
			var err error

			unmarshal := func(msgValue []byte, target interface{}) error {
				if err := h.Stream.Unmarshal(ctx, msgValue, target); err != nil {
					log.Printf("avro unmarshal error for id: `%s` with offset: `%d` with error: `%v`\n", idr, msg.TopicPartition.Offset, err)
					res.Errors++
					return err
				}
				return nil
			}

			art := new(schema.Article)
			if err = unmarshal(msg.Value, art); err != nil {
				continue
			}

			if excludeEvent(art.Event, req.ExcludeEvents) {
				continue
			}

			dta, err = json.Marshal(art)

			// Skip if the data is empty, example `{}`
			if len(dta) <= 2 {
				continue
			}

			if err != nil {
				log.Printf("json marshal error for id: `%s` with offset: `%d` with error: `%v`\n", idr, msg.TopicPartition.Offset, err)
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
		ers <- nil
		log.Printf("finishing buffers go routine for id: `%s`\n", idr)
	}()

	go func() {
		log.Printf("starting consumers go routine for id: `%s`\n", idr)
		cid := fmt.Sprintf("%s_%s", req.GetPrefix(), idr)
		pts := h.Cfg.GetPartitions(req.GetProject(), int(req.GetNamespace()))
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
				cbk := func(msg *kafka.Message) error {
					if msg.Timestamp.After(str) {
						log.Printf("finishing with the ErrEndMessage for id: `%s` with partition: `%v`\n", idr, pid)
						return libkafka.ErrReadAllEndMessage
					}

					mgs <- msg
					return nil
				}

				erp <- csr.ReadAll(ctx, int(req.GetSince()), tpc, []int{pid}, cbk)
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
		ers <- erm
		log.Printf("finishing consumers go routine for id: `%s`\n", idr)
	}()

	// wait for all ers (main export) and ecg (chunk export) channels to finish, these are errors from the top-level go-routines
	for i := 0; i < numExportWorkers; i++ {
		if err := <-ers; err != nil {
			return nil, err
		}
	}

	close(ers)

	if req.EnableChunking {
		for i := 0; i < h.Env.ChunkWorkers; i++ {
			if err := <-ecg; err != nil {
				return nil, err
			}
		}
		close(ecg)
	}

	hin := &s3.HeadObjectInput{
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s.tar.gz", req.GetPrefix(), bir)),
	}

	hop, err := h.S3.HeadObjectWithContext(ctx, hin)

	if err != nil {
		return nil, err
	}

	sze, err := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(*hop.ContentLength)/1000000), 64)

	if err != nil {
		return nil, err
	}

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
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s.json", req.GetPrefix(), bir)),
		Body:   bytes.NewReader(dta),
	}

	if _, err := h.S3.PutObjectWithContext(ctx, pin); err != nil {
		return nil, err
	}

	return res, nil
}

// handleChunk tars a bytes buffer and uploads as chunks/{snapshot_identifier}/{chunk_identifier}.tar.gz e.g., chunks/enwiki_namespace_0/chunk_2.tar.gz
// Also uploads metadata as chunks/{snapshot_identifier}/{chunk_identifier}.json e.g., chunks/enwiki_namespace_0/chunk_2.json
// It appends to a chunks slice by adding the chunks processed, if successful. Returns an error, if not.
func (h *Handler) handleChunk(buf []byte, opts *ChunkOptions) (*string, error) {
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
		Body:   bytes.NewReader(compressedBuffer.Bytes()),
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s/%s.tar.gz", "chunks", opts.Key, cid)),
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
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s/%s.json", "chunks", opts.Key, cid)),
		Body:   bytes.NewReader(dta),
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
func getSinceName(req *pb.ExportRequest, prefix, name string) string {
	if req.GetSince() > 0 && req.GetPrefix() != prefix {
		return fmt.Sprintf(
			"%s/%s",
			time.Unix(0, req.GetSince()*int64(time.Millisecond)).Format("2006-01-02"),
			name,
		)
	}

	return name
}
