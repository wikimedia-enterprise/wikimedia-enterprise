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
	res.Total = int32(1)

	_, err := h.S3.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(h.Env.AWSBucket),
		CopySource: aws.String(fmt.Sprintf("%s/aggregations/%s/%s.ndjson", h.Env.AWSBucket, h.Env.Prefix, h.Env.Prefix)),
		Key:        aws.String(fmt.Sprintf("aggregations/%s/%s_%s.ndjson", h.Env.Prefix, h.Env.Prefix, h.Env.FreeTierGroup)),
	})

	if err != nil {
		log.Println(err)
		res.Errors++
	}

	return res, nil
}
