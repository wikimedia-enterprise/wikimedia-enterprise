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
	"wikimedia-enterprise/services/snapshots/libraries/uploader"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	nbf "github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/klauspost/pgzip"
	"go.uber.org/dig"
)

// ErrEndMessage represents an error when we reached end of messages.
var ErrEndMessage = errors.New("end message")

type buffer struct {
	name string
	data []byte
}

func (b *buffer) len() int {
	return len(b.data)
}

type counter struct {
	itr int
	mtx sync.Mutex
}

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
	S3       s3iface.S3API
	Cfg      config.API
	Pool     libkafka.ConsumerGetter
	Stream   schema.Unmarshaler
	Uploader uploader.Uploader
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

	ngs := 4
	idr := fmt.Sprintf("%s_namespace_%d", req.GetProject(), req.GetNamespace())
	str := time.Now().UTC()
	ers := make(chan error, ngs)
	bfs := make(chan *buffer, 1)
	pbf := nbf.New(int64(h.Env.PipeBufferSize))
	prr, pwr := nio.Pipe(pbf)
	hrr := &HashReader{PipeReader: prr}

	if err != nil {
		return nil, err
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
			log.Printf("writing a buffer for id: `%s` with the name: `%s` and length: `%d`\n", idr, buf.name, buf.len())

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

			buf.data = nil // manually freeing the space from buffer

			if err := trw.Flush(); err != nil {
				ers <- err
				return
			}
		}

		_ = trw.Close()
		_ = gzw.Close()
		_ = pwr.Close()
		ers <- nil
		log.Printf("gzip go routine has finished for id: `%s`\n", idr)
	}()

	bir := idr

	// TODO: we need to find a better way to handle this, this is only a temporary solution
	if req.GetSince() > 0 && req.GetPrefix() != h.Env.Prefix {
		bir = fmt.Sprintf(
			"%s/%s",
			time.Unix(0, req.GetSince()*int64(time.Millisecond)).Format("2006-01-02"),
			idr,
		)
	}

	go func() {
		log.Printf("starting upload go routine for id: `%s`\n", idr)
		uin := &s3manager.UploadInput{
			Body:   hrr,
			Bucket: aws.String(h.Env.AWSBucket),
			Key:    aws.String(fmt.Sprintf("%s/%s.tar.gz", req.GetPrefix(), bir)),
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
		newBuffer := func() *buffer {
			return &buffer{
				name: fmt.Sprintf("%s_%d.ndjson", idr, ctr.get()),
				data: make([]byte, 0),
			}
		}
		buf := newBuffer()

	MSGS:
		for msg := range mgs {
			res.Total++

			art := new(schema.Article)

			if err := h.Stream.Unmarshal(ctx, msg.Value, art); err != nil {
				log.Printf("avro unmarshal error for id: `%s` with offset: `%d` with error: `%v`\n", idr, msg.TopicPartition.Offset, err)
				res.Errors++
				continue
			}

			// in case we encounter an article to exclude, no need to push it into snapshot
			if art.Event != nil && len(req.ExcludeEvents) > 0 {
				for _, ee := range req.ExcludeEvents {
					if art.Event.Type == ee {
						continue MSGS
					}
				}
			}

			dta, err := json.Marshal(art)

			if err != nil {
				log.Printf("json marshal error for id: `%s` with offset: `%d` with error: `%v`\n", idr, msg.TopicPartition.Offset, err)
				res.Errors++
				continue
			}

			if buf.len()+len(dta) > h.Env.BufferSize {
				bfs <- buf
				buf = newBuffer()
			}

			buf.data = append(buf.data, append(dta, []byte("\n")...)...)
		}

		if buf.len() > 0 {
			bfs <- buf
		}

		close(bfs)
		ers <- nil
		log.Printf("finishing buffers go routine for id: `%s`\n", idr)
	}()

	go func() {
		log.Printf("starting consumers go routine for id: `%s`\n", idr)
		cid := fmt.Sprintf("%s_%s", req.GetPrefix(), idr)
		pts := h.Cfg.GetPartitions(req.GetProject(), int(req.GetNamespace()))
		erp := make(chan error, len(pts))

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

		for i := 0; i < len(pts); i++ {
			if err := <-erp; err != nil {
				erm = err
			}
		}

		close(mgs)
		ers <- erm
		log.Printf("finishing consumers go routine for id: `%s`\n", idr)
	}()

	for i := 0; i < ngs; i++ {
		if err := <-ers; err != nil {
			return nil, err
		}
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
