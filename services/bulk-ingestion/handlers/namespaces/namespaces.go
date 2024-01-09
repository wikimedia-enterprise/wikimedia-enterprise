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

	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"

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
}

// Handler gets all the projects by calling mediawiki actions API. Then, it gets all the namespaces for each project
// and produces a kafka message for each namespace.
func Handler(ctx context.Context, p *Parameters, req *pb.NamespacesRequest) (*pb.NamespacesResponse, error) {
	res := new(pb.NamespacesResponse)
	nss, err := p.API.GetNamespaces(ctx, p.Env.DefaultProject)

	if err != nil {
		return nil, err
	}

	dcs := map[int]string{
		0:  "The main namespace, article namespace, or mainspace is the namespace of Wikipedia that contains the encyclopedia proper—that is, where 'live' Wikipedia articles reside.",
		6:  "The File namespace is a namespace consisting of administration pages in which all of Wikipedia's media content resides. On Wikipedia, all media filenames begin with the prefix File:, including data files for images, video clips, or audio clips, including document length clips; or MIDI files (a small file of computer music instructions).",
		10: "The Template namespace on Wikipedia is used to store templates, which contain Wiki markup intended for inclusion on multiple pages, usually via transclusion. Although the Template namespace is used for storing most templates, it is possible to transclude and substitute from other namespaces, and so some template pages are placed in other namespaces, such as the User namespace.",
		14: "Categories are intended to group together pages on similar subjects. They are implemented by a MediaWiki feature that adds any page with a text like [[Category:XYZ]] in its wiki markup to the automated listing that is the category with name XYZ. Categories help readers to find, and navigate around, a subject area, to see pages sorted by title, and to thus find article relationships.",
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
