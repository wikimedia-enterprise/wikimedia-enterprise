package aggregatecommons

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/klauspost/pgzip"
	"go.uber.org/dig"

	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"
)

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

	keys, err := h.listObjects(ctx, prefix)
	if err != nil {
		return nil, err
	}

	res := &pb.AggregateCommonsResponse{}
	bufferChan := make(chan []byte, 100)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		h.processObjects(ctx, keys, bufferChan, res)
	}()

	go func() {
		defer wg.Done()
		h.processBuffers(bufferChan, errChan, req)
	}()

	wg.Wait()

	close(errChan)

	for i := 0; i < 1; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (h *Handler) listObjects(ctx context.Context, prefix string) ([]*string, error) {
	var keys []*string
	err := h.S3.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(h.Env.AWSBucket),
		Prefix: aws.String(prefix),
	}, func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, obj := range page.Contents {
			if strings.HasSuffix(*obj.Key, ".json") {
				keys = append(keys, obj.Key)
			}
		}
		return true
	})

	return keys, err
}

func (h *Handler) processObjects(ctx context.Context, keys []*string, bufferChan chan<- []byte, res *pb.AggregateCommonsResponse) {
	defer close(bufferChan)

	for _, key := range keys {
		res.Total++
		output, err := h.S3.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(h.Env.AWSBucket),
			Key:    key,
		})

		if err != nil {
			log.Error(err.Error(), log.Any("key", *key))
			res.Errors++
			continue
		}

		data, err := io.ReadAll(output.Body)

		if oer := output.Body.Close(); oer != nil {
			log.Error(oer.Error(), log.Any("key", *key))
			res.Errors++
			continue
		}

		if err != nil {
			log.Error(err.Error(), log.Any("key", *key))
			res.Errors++
			continue
		}

		bufferChan <- data
	}
}

func (h *Handler) processBuffers(bufferChan <-chan []byte, errChan chan<- error, req *pb.AggregateCommonsRequest) {
	pr, pw := io.Pipe()
	gzw := pgzip.NewWriter(pw)
	tw := tar.NewWriter(gzw)

	go func() {
		defer pw.Close()
		defer gzw.Close()
		defer tw.Close()

		lineCount := 0
		fileCount := 0
		var currentBuffer []byte

		for data := range bufferChan {
			currentBuffer = append(currentBuffer, data...)
			currentBuffer = append(currentBuffer, '\n')
			lineCount++

			if lineCount >= h.Env.LineLimit {
				h.Write(fileCount, currentBuffer, errChan, tw)
				lineCount = 0
				fileCount++
				currentBuffer = currentBuffer[:0]
			}
		}

		if len(currentBuffer) > 0 {
			h.Write(fileCount, currentBuffer, errChan, tw)
		}

		if err := tw.Flush(); err != nil {
			errChan <- err
			return
		}
	}()

	errChan <- h.uploadToS3(pr, req)
}

func (h *Handler) uploadToS3(reader io.Reader, req *pb.AggregateCommonsRequest) error {
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

	_, err := h.Uploader.Upload(input)
	return err
}

func (h *Handler) Write(fct int, cbr []byte, errChan chan<- error, tw *tar.Writer) {
	filename := fmt.Sprintf("commonswiki_namespace_6_%d.ndjson", fct)
	hdr := &tar.Header{
		Name:    filename,
		Mode:    0600,
		Size:    int64(len(cbr)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		errChan <- err
		return
	}

	if _, err := tw.Write(cbr); err != nil {
		errChan <- err
		return
	}
}
