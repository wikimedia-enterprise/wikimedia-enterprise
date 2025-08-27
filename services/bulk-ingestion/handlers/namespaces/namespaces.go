// Package namespaces implements the gRPC handler for namespaces. The handler produces a kafka message for each namespace per project.
// The API response from this handler includes the total number of namespaces processed and the total number of errors during processing.
package namespaces

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"wikimedia-enterprise/services/bulk-ingestion/submodules/config"
	"wikimedia-enterprise/services/bulk-ingestion/submodules/schema"
	"wikimedia-enterprise/services/bulk-ingestion/submodules/wmf"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"go.uber.org/dig"
)

// Parameters dependencies for the handler.
type Parameters struct {
	dig.In
	Stream schema.UnmarshalProducer
	API    wmf.API
	S3     s3iface.S3API
	Env    *env.Environment
	Cfg    config.NamespacesMetadataGetter
}

// Handler gets all the projects by calling mediawiki actions API. Then, it gets all the namespaces for each project
// and produces a kafka message for each namespace.
func Handler(ctx context.Context, p *Parameters, req *pb.NamespacesRequest) (*pb.NamespacesResponse, error) {
	res := new(pb.NamespacesResponse)
	nss, err := p.API.GetNamespaces(ctx, p.Env.DefaultProject)

	if err != nil {
		return nil, err
	}

	dcs := map[int]string{}
	for ns, mtd := range p.Cfg.GetNamespacesMetadata() {
		dcs[ns] = mtd.Description
	}

	mgs := []*schema.Message{}
	nns := new(strings.Builder)

	for _, nsp := range nss {
		switch nsp.ID {
		case 0, 6, 10, 14:
			snp := &schema.Namespace{
				Identifier:  nsp.ID,
				Name:        nsp.Name,
				Description: dcs[nsp.ID],
			}

			if len(snp.Name) == 0 && nsp.ID == 0 {
				snp.Name = "Article"
			}

			dta, err := json.Marshal(snp)

			if err != nil {
				return nil, err
			}

			lin := &s3.PutObjectInput{
				Bucket: aws.String(p.Env.AWSBucket),
				Key:    aws.String(fmt.Sprintf("namespaces/%d.json", snp.Identifier)),
				Body:   bytes.NewReader(dta),
			}

			if _, err := p.S3.PutObjectWithContext(ctx, lin); err != nil {
				return nil, err
			}

			_, _ = fmt.Fprintf(nns, "%s\n", string(dta))
			snp.Event = schema.NewEvent(schema.EventTypeCreate)
			mgs = append(mgs, &schema.Message{
				Config: schema.ConfigNamespace,
				Topic:  p.Env.TopicNamespaces,
				Key: &schema.Key{
					Identifier: fmt.Sprintf("namespaces/%d", nsp.ID),
					Type:       schema.KeyTypeNamespace,
				},
				Value: snp,
			})

			res.Total++
		}
	}

	lin := &s3.PutObjectInput{
		Bucket: aws.String(p.Env.AWSBucket),
		Key:    aws.String("aggregations/namespaces/namespaces.ndjson"),
		Body:   strings.NewReader(nns.String()),
	}

	if _, err := p.S3.PutObjectWithContext(ctx, lin); err != nil {
		return nil, err
	}

	if err := p.Stream.Produce(ctx, mgs...); err != nil {
		return nil, err
	}

	_ = p.Stream.Flush(1000)

	return res, nil
}
