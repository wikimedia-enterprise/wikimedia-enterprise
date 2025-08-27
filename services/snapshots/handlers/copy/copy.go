// Package copy creates new handler for copy with dependency injection.
package copy

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/submodules/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/dig"
)

const (
	// copyWorkers default number of workers for concurrency.
	copyWorkers = 10

	// maxUploadSizeBytes for copying in parts. Should be 5MB <= maxUploadSizeBytes <= 4 GB.
	maxUploadSizeBytes = 4294967296

	// maxDeleteBatch for deleting chunks
	maxDeleteBatch = 1000
)

var (
	// chunkJSONRe regex for chunk JSON files.
	chunkJSONRe = regexp.MustCompile(`^chunk_(\d+)\.json$`)

	// chunkTarGzRe regex for chunk tar.gz files.
	chunkTarGzRe = regexp.MustCompile(`^chunk_(\d+)\.tar\.gz$`)
)

// Handler copy handler execution struct with dependency injection.
type Handler struct {
	dig.In
	Env *env.Environment
	S3  s3tracerproxy.S3TracerProxy
}

// Copy copies project snapshots and metadata.
func (h *Handler) Copy(ctx context.Context, req *pb.CopyRequest) (*pb.CopyResponse, error) {
	if req.Workers == 0 {
		req.Workers = copyWorkers
	}

	res := new(pb.CopyResponse)
	pts := make(map[string]string) // source-destination path pairs to copy

	for _, prj := range req.Projects {
		pts[fmt.Sprintf("snapshots/%s_namespace_%d.json", prj, req.Namespace)] = fmt.Sprintf("snapshots/%s_namespace_%d_%s.json", prj, req.Namespace, h.Env.FreeTierGroup)
		pts[fmt.Sprintf("snapshots/%s_namespace_%d.tar.gz", prj, req.Namespace)] = fmt.Sprintf("snapshots/%s_namespace_%d_%s.tar.gz", prj, req.Namespace, h.Env.FreeTierGroup)
		prefix := fmt.Sprintf("chunks/%s_namespace_%d/", prj, req.Namespace)
		chunks, err := h.CopyChunkFiles(ctx, prefix, prj, req.Namespace)
		if err != nil {
			log.Error(
				"failed to get chunk copy paths or delete old group_1 chunks",
				log.Any("err", err),
				log.Any("project", prj),
				log.Any("namespace", req.Namespace),
				log.Any("chunk count", len(chunks)),
			)

			continue
		}

		// Ignore project that have no chunks
		if len(chunks) == 0 {

			continue
		}

		for src, dst := range chunks {
			pts[src] = dst
		}
	}

	sze := len(pts)
	res.Total = int32(sze) // #nosec G115

	srs := make(chan string, sze)
	ers := make(chan error, sze)

	for src := range pts {
		srs <- src
	}

	close(srs)

	for i := 0; i < int(req.Workers); i++ {
		go func() {
			for src := range srs {
				hr, err := h.S3.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
					Bucket: aws.String(h.Env.AWSBucket),
					Key:    aws.String(src),
				})

				if err != nil {
					ers <- err
					continue
				}

				if *hr.ContentLength <= maxUploadSizeBytes {
					_, err = h.S3.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
						Bucket:     aws.String(h.Env.AWSBucket),
						CopySource: aws.String(fmt.Sprintf("%s/%s", h.Env.AWSBucket, src)),
						Key:        aws.String(pts[src]),
					})

					ers <- err
					continue
				}

				err = h.MultipartCopy(ctx, src, pts[src], *hr.ContentLength)
				ers <- err
			}
		}()
	}

	for i := 0; i < sze; i++ {
		if err := <-ers; err != nil {
			fmt.Println(err)
			res.Errors++
		}
	}

	close(ers)

	return res, nil
}

// CopyChunkFiles deletes old free tier chunks all in the S3 bucket before creating new ones by copying.
func (h *Handler) CopyChunkFiles(ctx context.Context, prefix string, prj string, namespace int32) (map[string]string, error) {
	chunkCopyPaths := make(map[string]string)
	filesToDelete := make([]*s3.ObjectIdentifier, 0)
	freeTier := fmt.Sprintf("_%s.", h.Env.FreeTierGroup)

	cleanUpChunkFiles := func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			key := strings.TrimPrefix(*obj.Key, prefix)
			if strings.Contains(key, freeTier) {
				tempKeyForRegex := strings.Replace(key, freeTier, "", 1)
				if chunkJSONRe.MatchString(tempKeyForRegex) || chunkTarGzRe.MatchString(tempKeyForRegex) {
					filesToDelete = append(filesToDelete, &s3.ObjectIdentifier{
						Key: obj.Key,
					})
				}
			} else {
				jsn := chunkJSONRe.FindStringSubmatch(key)
				trn := chunkTarGzRe.FindStringSubmatch(key)

				if len(jsn) > 0 {
					chunkCopyPaths[*obj.Key] = fmt.Sprintf("%schunk_%s_%s.json", prefix, jsn[1], h.Env.FreeTierGroup)
				} else if len(trn) > 0 {
					chunkCopyPaths[*obj.Key] = fmt.Sprintf("%schunk_%s_%s.tar.gz", prefix, trn[1], h.Env.FreeTierGroup)
				}
			}
		}
		return true
	}

	err := h.S3.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(h.Env.AWSBucket),
		Prefix: aws.String(prefix),
	}, cleanUpChunkFiles)

	if err != nil {
		return nil, fmt.Errorf("failed to list objects with prefix '%s': %w", prefix, err)
	}

	if len(filesToDelete) > 0 {
		for i := 0; i < len(filesToDelete); i += maxDeleteBatch {
			end := i + maxDeleteBatch
			if end > len(filesToDelete) {
				end = len(filesToDelete)
			}

			batch := filesToDelete[i:end]
			_, err = h.S3.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(h.Env.AWSBucket),
				Delete: &s3.Delete{
					Objects: batch,
				},
			})

			if err != nil {
				log.Error(
					"failed to delete existing chunk files",
					log.Any("err", err),
					log.Any("project", prj),
					log.Any("namespace", namespace),
					log.Any("prefix", prefix),
				)
			} else {
				log.Info(
					"deleted existing chunk files",
					log.Any("project", prj),
					log.Any("namespace", namespace),
					log.Any("deleted_count", len(batch)),
					log.Any("prefix", prefix),
				)
			}
		}
	}

	if len(chunkCopyPaths) > 0 {
		log.Info(
			"chunks for copying",
			log.Any("project", prj),
			log.Any("namespace", namespace),
			log.Any("chunks_count", len(chunkCopyPaths)),
			log.Any("prefix", prefix),
		)
	} else {
		log.Info(
			"no chunks for copying",
			log.Any("project", prj),
			log.Any("namespace", namespace),
			log.Any("prefix", prefix),
		)
	}

	return chunkCopyPaths, nil
}

// multipartCopy copies a file in parts from src to dst.
func (h *Handler) MultipartCopy(ctx context.Context, src, dst string, size int64) error {
	cmr, err := h.S3.CreateMultipartUploadWithContext(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(dst),
	})
	if err != nil {
		return fmt.Errorf("failed to initiate multipart upload for %s: %w", dst, err)
	}

	cmu := &s3.CompletedMultipartUpload{}
	maxPart := int(math.Ceil(float64(size) / float64(maxUploadSizeBytes)))

	for prt := 0; prt < maxPart; prt++ {
		from := prt * maxUploadSizeBytes
		to := (prt * maxUploadSizeBytes) + maxUploadSizeBytes

		if prt != 0 {
			from += 1
		}
		if to > int(size) {
			to = int(size) - 1
		}

		upr, err := h.S3.UploadPartCopyWithContext(ctx, &s3.UploadPartCopyInput{
			Bucket:          aws.String(h.Env.AWSBucket),
			CopySource:      aws.String(fmt.Sprintf("%s/%s", h.Env.AWSBucket, src)),
			CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", from, to)),
			Key:             aws.String(dst),
			PartNumber:      aws.Int64(int64(prt) + 1),
			UploadId:        aws.String(*cmr.UploadId),
		})

		if err != nil {
			return fmt.Errorf("failed to upload part %d for %s: %w", prt, dst, err)
		}

		cmu.Parts = append(cmu.Parts, &s3.CompletedPart{
			ETag:       upr.CopyPartResult.ETag,
			PartNumber: aws.Int64(int64(prt) + 1),
		})
	}

	_, err = h.S3.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(h.Env.AWSBucket),
		Key:             aws.String(dst),
		UploadId:        aws.String(*cmr.UploadId),
		MultipartUpload: cmu,
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload for %s: %w", dst, err)
	}

	return nil
}
