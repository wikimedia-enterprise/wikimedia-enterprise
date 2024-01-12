// Package aggregate creates new handler for aggregation with dependency injection.
package aggregate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"go.uber.org/dig"

	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/snapshots/config/env"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
)

// Handler represents dependency injection and logic for the aggregate handler.
type Handler struct {
	dig.In
	Env *env.Environment
	S3  s3iface.S3API
}

// Aggregate goes through the snapshots and creates an aggregation for API call listings.
func (h *Handler) Aggregate(ctx context.Context, req *pb.AggregateRequest) (*pb.AggregateResponse, error) {
	prx := req.GetPrefix()

	if len(prx) == 0 {
		prx = h.Env.Prefix
	}

	key := prx

	if req.GetSince() > 0 {
		key = fmt.Sprintf("%s/%s",
			prx,
			time.Unix(0, req.GetSince()*int64(time.Millisecond)).Format("2006-01-02"),
		)
	}

	kys := []*string{}
	lin := &s3.ListObjectsV2Input{
		Bucket: aws.String(h.Env.AWSBucket),
		Prefix: aws.String(key),
	}
	gpf := func(out *s3.ListObjectsV2Output, _ bool) bool {
		for _, cnt := range out.Contents {
			if strings.HasSuffix(*cnt.Key, ".json") && !strings.Contains(*cnt.Key, h.Env.FreeTierGroup) {
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
			log.Println(err.Error())
			continue
		}

		snp := new(schema.Snapshot)

		if err := json.NewDecoder(out.Body).Decode(snp); err != nil {
			_ = out.Body.Close()
			log.Println(err.Error())
			res.Errors++
			continue
		}

		_ = out.Body.Close()

		if snp.Size != nil && snp.Size.Value > 0 {
			dta, err := json.Marshal(snp)

			if err != nil {
				log.Println(err.Error())
				res.Errors++
				continue
			}

			hrs += fmt.Sprintf("%s\n", string(dta))
		}
	}

	pin := &s3.PutObjectInput{
		Bucket: aws.String(h.Env.AWSBucket),
		Key:    aws.String(fmt.Sprintf("aggregations/%s/%s.ndjson", key, prx)),
		Body:   strings.NewReader(strings.Trim(hrs, "\n")),
	}

	if _, err := h.S3.PutObjectWithContext(ctx, pin); err != nil {
		return nil, err
	}

	return res, nil
}
