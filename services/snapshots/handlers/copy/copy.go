// Package copy creates new handler for copy with dependency injection.
package copy

import (
	"context"
	"fmt"
	"math"

	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/dig"
)

// copyWorkers default number of workers for concurrency.
const copyWorkers = 10

// maxUploadSizeBytes for copying in parts. Should be 5MB <= maxUploadSizeBytes <= 4 GB.
const maxUploadSizeBytes = 4294967296

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
	}

	sze := len(pts)
	res.Total = int32(sze)

	srs := make(chan string, sze)
	ers := make(chan error, sze)

	for src := range pts {
		srs <- src
	}

	close(srs)

	for i := 0; i < int(req.Workers); i++ {
		go func() {
		PATHS:
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

				cmr, err := h.S3.CreateMultipartUploadWithContext(ctx, &s3.CreateMultipartUploadInput{
					Bucket: aws.String(h.Env.AWSBucket),
					Key:    aws.String(pts[src]),
				})

				if err != nil {
					ers <- err
					continue
				}

				cmu := &s3.CompletedMultipartUpload{}
				maxPart := int(math.Ceil(float64(*hr.ContentLength) / float64(maxUploadSizeBytes)))

				for prt := 0; prt < maxPart; prt++ {
					from := prt * maxUploadSizeBytes
					to := (prt * maxUploadSizeBytes) + maxUploadSizeBytes

					if prt != 0 {
						from += 1
					}

					if to > int(*hr.ContentLength) {
						to = int(*hr.ContentLength) - 1
					}

					upr, err := h.S3.UploadPartCopyWithContext(ctx, &s3.UploadPartCopyInput{
						Bucket:          aws.String(h.Env.AWSBucket),
						CopySource:      aws.String(fmt.Sprintf("%s/%s", h.Env.AWSBucket, src)),
						CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", from, to)),
						Key:             aws.String(pts[src]),
						PartNumber:      aws.Int64(int64(prt) + 1),
						UploadId:        aws.String(*cmr.UploadId),
					})

					if err != nil {
						ers <- err
						continue PATHS
					}

					cmu.Parts = append(cmu.Parts, &s3.CompletedPart{
						ETag:       upr.CopyPartResult.ETag,
						PartNumber: aws.Int64(int64(prt) + 1),
					})
				}

				_, err = h.S3.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
					Bucket:          aws.String(h.Env.AWSBucket),
					Key:             aws.String(pts[src]),
					UploadId:        aws.String(*cmr.UploadId),
					MultipartUpload: cmu,
				})

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
