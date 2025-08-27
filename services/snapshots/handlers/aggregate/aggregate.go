// Package aggregate creates new handler for aggregation with dependency injection.
package aggregate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/dig"

	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/submodules/log"
	"wikimedia-enterprise/services/snapshots/submodules/schema"
)

// Handler represents dependency injection and logic for the aggregate handler.
type Handler struct {
	dig.In
	Env *env.Environment
	S3  s3tracerproxy.S3TracerProxy
}

// Aggregate goes through the snapshots and creates an aggregation for API call listings.
func (h *Handler) Aggregate(ctx context.Context, req *pb.AggregateRequest) (*pb.AggregateResponse, error) {
	prx := req.GetPrefix()

	if len(prx) == 0 {
		prx = h.Env.Prefix
	}

	key := prx

	// if snapshot_identifier is present, append it to the key
	if req.GetSnapshot() != "" {
		key = fmt.Sprintf("%s/%s", key, req.GetSnapshot())
	}
	outputPath := fmt.Sprintf("aggregations/%s/%s.ndjson", key, prx)

	if req.GetSince() > 0 {
		if req.GetPrefix() == "batches" {
			key = fmt.Sprintf("%s/%s",
				prx,
				time.Unix(0, req.GetSince()*int64(time.Millisecond)).Format("2006-01-02"),
			)
			outputPath = fmt.Sprintf("aggregations/%s/%s.ndjson", key, prx)
		} else if strings.Index(req.GetPrefix(), "batches") == 0 {
			// Example: batches/2025-07-02/02
			key = req.GetPrefix()
			outputPath = fmt.Sprintf("aggregations/%s/batches.ndjson", key)
		}
	}

	kys := []*string{}
	lin := &s3.ListObjectsV2Input{
		Bucket: aws.String(h.Env.AWSBucket),
		Prefix: aws.String(key),
	}
	gpf := func(out *s3.ListObjectsV2Output, _ bool) bool {
		for _, cnt := range out.Contents {
			withoutPrefix, found := strings.CutPrefix(*cnt.Key, key)
			withoutPrefix = strings.TrimLeft(withoutPrefix, "/")
			// Ignore files in sub-paths.
			isInSubPath := !found || strings.Contains(withoutPrefix, "/")
			if strings.HasSuffix(*cnt.Key, ".json") && !strings.Contains(*cnt.Key, h.Env.FreeTierGroup) && !isInSubPath {
				kys = append(kys, cnt.Key)
			}
		}

		return true
	}

	if err := h.S3.ListObjectsV2PagesWithContext(ctx, lin, gpf); err != nil {
		return nil, err
	}

	res := new(pb.AggregateResponse)
	hrs := ""

	for _, key := range kys {
		res.Total++

		out, err := h.S3.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(h.Env.AWSBucket),
			Key:    key,
		})

		if err != nil {
			log.Error(
				err.Error(),
				log.Any("key", key),
			)
			continue
		}

		snp := new(schema.Snapshot)

		if err := json.NewDecoder(out.Body).Decode(snp); err != nil {
			_ = out.Body.Close()
			log.Error(
				err.Error(),
				log.Any("key", key),
			)
			res.Errors++
			continue
		}

		_ = out.Body.Close()

		if snp.Size != nil && snp.Size.Value > 0 {
			dta, err := json.Marshal(snp)

			if err != nil {
				log.Error(
					err.Error(),
					log.Any("key", key),
				)

				res.Errors++
				continue
			}

			hrs += fmt.Sprintf("%s\n", string(dta))
		}
	}

	if len(hrs) > 0 {
		pin := &s3.PutObjectInput{
			Bucket: aws.String(h.Env.AWSBucket),
			Key:    aws.String(outputPath),
			Body:   strings.NewReader(strings.Trim(hrs, "\n")),
		}

		if _, err := h.S3.PutObjectWithContext(ctx, pin); err != nil {
			return nil, err
		}
	} else {
		log.Info(
			"no metadata found to aggregate",
			log.Any("key", key),
		)
	}

	return res, nil
}
