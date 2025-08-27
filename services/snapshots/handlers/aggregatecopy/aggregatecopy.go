// Package aggregatecopy creates new handler for AggregateCopy with dependency injection.
package aggregatecopy

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/dig"

	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	s3tracerproxy "wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
)

// Handler represents dependency injection and logic for the aggregate copy handler.
type Handler struct {
	dig.In
	Env *env.Environment
	S3  s3tracerproxy.S3TracerProxy
}

// AggregateCopy copies project snapshots aggregated metadata.
func (h *Handler) AggregateCopy(ctx context.Context, req *pb.AggregateCopyRequest) (*pb.AggregateCopyResponse, error) {
	res := new(pb.AggregateCopyResponse)
	res.Total = int32(0)

	sources := []struct {
		src string
		dst string
	}{}

	// If no projects and namespaces then copy the root metadata file, example `aggregations/snapshots/snapshots.ndjson`
	if len(req.Projects) == 0 && len(req.Namespaces) == 0 {
		sources = append(sources, struct {
			src string
			dst string
		}{
			src: fmt.Sprintf("aggregations/%s/%s.ndjson", h.Env.Prefix, h.Env.Prefix),
			dst: fmt.Sprintf("aggregations/%s/%s_%s.ndjson", h.Env.Prefix, h.Env.Prefix, h.Env.FreeTierGroup),
		})
	}

	for _, prj := range req.Projects {
		for _, ns := range req.Namespaces {
			sources = append(sources, struct {
				src string
				dst string
			}{
				src: fmt.Sprintf("aggregations/chunks/%s_namespace_%d/chunks.ndjson", prj, ns),
				dst: fmt.Sprintf("aggregations/chunks/%s_namespace_%d/chunks_%s.ndjson", prj, ns, h.Env.FreeTierGroup),
			})
		}
	}

	for _, pair := range sources {
		_, err := h.S3.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(h.Env.AWSBucket),
			CopySource: aws.String(fmt.Sprintf("%s/%s", h.Env.AWSBucket, pair.src)),
			Key:        aws.String(pair.dst),
		})

		if err != nil {
			log.Println(err)
			log.Println(pair.src)
			log.Println(pair.dst)
			res.Errors++
		}

		res.Total++
	}

	return res, nil
}
