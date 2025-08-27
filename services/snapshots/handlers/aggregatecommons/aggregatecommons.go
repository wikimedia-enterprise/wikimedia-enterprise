package aggregatecommons

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/djherbis/nio/v3"
	"github.com/klauspost/pgzip"
	"go.uber.org/dig"

	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"
	"wikimedia-enterprise/services/snapshots/submodules/log"

	nbf "github.com/djherbis/buffer"
)

// errSize error channel size
const errSize = 2

type Handler struct {
	dig.In
	Env      *env.Environment
	S3       s3tracerproxy.S3TracerProxy
	Uploader uploader.Uploader
}

func (h *Handler) AggregateCommons(ctx context.Context, req *pb.AggregateCommonsRequest) (*pb.AggregateCommonsResponse, error) {
	var prefix string
	if req.GetTimePeriod() != "" {
		prefix = fmt.Sprintf("commons/batches/%s", req.GetTimePeriod())
	} else {
		prefix = "commons/pages"
	}

	res := &pb.AggregateCommonsResponse{}

	keysChan := make(chan *string, h.Env.CommonsAggregationKeyChannelSize)
	bufferChan := make(chan []byte, h.Env.PipeBufferSize)
	errChan := make(chan error, errSize)

	pbf := nbf.New(int64(h.Env.PipeBufferSize))
	prr, pwr := nio.Pipe(pbf)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		h.processObjects(ctx, keysChan, bufferChan, h.Env.CommonsAggregationWorkercount, res)
	}()

	go func() {
		defer wg.Done()
		h.processBuffers(bufferChan, errChan, pwr)
	}()

	go func() {
		defer wg.Done()
		err := h.upload(prr, req)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		if err := h.listObjects(ctx, prefix, keysChan); err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for i := 0; i < errSize; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (h *Handler) listObjects(ctx context.Context, prefix string, keys chan<- *string) error {
	defer close(keys)
	total := 0

	err := h.S3.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(h.Env.AWSBucketCommons),
		Prefix: aws.String(prefix),
	}, func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, obj := range page.Contents {
			if strings.HasSuffix(*obj.Key, ".json") {
				keys <- obj.Key
				total++
			}
		}
		return true
	})

	log.Info("all keys have been listed", log.Any("total_keys", total))

	return err
}

func (h *Handler) processObjects(ctx context.Context, keys <-chan *string, bufferChan chan<- []byte, workers int, res *pb.AggregateCommonsResponse) {
	defer close(bufferChan)

	var mu sync.Mutex
	total := int32(0)
	errors := int32(0)

	startTime := time.Now()
	log.Info("starting to process objects")

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			h.processObject(ctx, keys, bufferChan, &mu, &total, &errors)
		}()
	}

	done := make(chan struct{})

	// Log periodically
	go func() {
		ticker := time.NewTicker(time.Duration(h.Env.LogInterval) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				currentTotal := total
				mu.Unlock()
				log.Info("processing objects in progress", log.Any("total_processed", currentTotal))
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
	close(done)

	res.Total = total
	res.Errors = errors

	log.Info("finished processing all objects", log.Any("total_processed", res.Total),
		log.Any("errors", res.Errors),
		log.Any("duration", time.Since(startTime)))
}

func (h *Handler) processObject(ctx context.Context, keys <-chan *string, bufferChan chan<- []byte, mu *sync.Mutex, total, errors *int32) {
	for key := range keys {
		mu.Lock()
		*total++
		mu.Unlock()

		output, err := h.S3.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(h.Env.AWSBucketCommons),
			Key:    key,
		})

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
				log.Info("object for S3 key not found", log.Any("key", *key))
			} else {
				log.Error(err.Error(), log.Any("key", *key))
			}
			mu.Lock()
			*errors++
			mu.Unlock()
			return
		}

		data, err := io.ReadAll(output.Body)
		oer := output.Body.Close()
		if oer != nil {
			log.Error(oer.Error(), log.Any("key", *key))
			mu.Lock()
			*errors++
			mu.Unlock()
			return
		}

		if err != nil {
			log.Error(err.Error(), log.Any("key", *key))
			mu.Lock()
			*errors++
			mu.Unlock()
			return
		}

		bufferChan <- data
	}
}

func (h *Handler) processBuffers(bufferChan <-chan []byte, errChan chan<- error, pwr *nio.PipeWriter) {
	gzipWriter := pgzip.NewWriter(pwr)
	if err := gzipWriter.SetConcurrency(1<<20, runtime.NumCPU()*2); err != nil {
		errChan <- fmt.Errorf("error setting gzip concurrency: %w", err)
		return
	}

	tarWriter := tar.NewWriter(gzipWriter)

	defer func() {
		if err := tarWriter.Close(); err != nil {
			errChan <- fmt.Errorf("error closing tar writer: %w", err)
		}
		if err := gzipWriter.Close(); err != nil {
			errChan <- fmt.Errorf("error closing gzip writer: %w", err)
		}
		if err := pwr.Close(); err != nil {
			errChan <- fmt.Errorf("error closing pipe writer: %w", err)
		}
	}()

	lineCount := 0
	fileCount := 0
	var currentBuffer bytes.Buffer

	for data := range bufferChan {
		currentBuffer.Write(data)
		currentBuffer.WriteByte('\n')
		lineCount++

		if lineCount >= h.Env.LineLimit {
			if err := h.Write(fileCount, currentBuffer.Bytes(), errChan, tarWriter); err != nil {
				errChan <- err
				return
			}
			lineCount = 0
			fileCount++
			currentBuffer.Reset()
		}
	}

	if currentBuffer.Len() > 0 {
		if err := h.Write(fileCount, currentBuffer.Bytes(), errChan, tarWriter); err != nil {
			errChan <- err
			return
		}
	}

	if err := tarWriter.Flush(); err != nil {
		errChan <- fmt.Errorf("error flushing tar writer: %w", err)
		return
	}

	if err := gzipWriter.Flush(); err != nil {
		errChan <- fmt.Errorf("error flushing gzip writer: %w", err)
		return
	}
}

func (h *Handler) upload(reader *nio.PipeReader, req *pb.AggregateCommonsRequest) error {
	var key string
	var tag string
	if req.GetTimePeriod() != "" {
		key = fmt.Sprintf("batches/%s/commonswiki_namespace_6.tar.gz", req.GetTimePeriod())
		tag = "type=commons-batches"
	} else {
		key = "snapshots/commonswiki_namespace_6.tar.gz"
		tag = "type=commons-snapshots"
	}

	input := &s3manager.UploadInput{
		Bucket:  aws.String(h.Env.AWSBucket),
		Key:     aws.String(key),
		Body:    reader,
		Tagging: aws.String(tag),
	}
	opt := func(upl *s3manager.Uploader) {
		upl.Concurrency = h.Env.UploadConcurrency
		upl.PartSize = int64(h.Env.UploadPartSize)
	}

	log.Info("starting S3 upload", log.Any("key", key), log.Any("tag", tag))
	startTime := time.Now()

	_, err := h.Uploader.Upload(input, opt)
	if err != nil {
		log.Error("s3 upload failed", log.Any("error", err))
	} else {
		log.Info("s3 upload completed",
			log.Any("duration", time.Since(startTime)),
			log.Any("key", key))
	}

	return err
}

func (h *Handler) Write(fct int, cbr []byte, errChan chan<- error, tw *tar.Writer) error {
	filename := fmt.Sprintf("commonswiki_namespace_6_%d.ndjson", fct)
	hdr := &tar.Header{
		Name:    filename,
		Mode:    0600,
		Size:    int64(len(cbr)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		errChan <- fmt.Errorf("error writing tar header: %w", err)
		return err
	}

	if _, err := tw.Write(cbr); err != nil {
		errChan <- fmt.Errorf("error writing data to tar: %w", err)
		return err
	}

	err := tw.Flush()
	if err != nil {
		errChan <- fmt.Errorf("error flushing tar writer: %w", err)
	}

	return nil
}
