// Package projects is the gRPC handler for `Projects` endpoint.
package projects

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"go.uber.org/dig"
)

// Parameters dependencies for the handler.
type Parameters struct {
	dig.In
	Stream schema.UnmarshalProducer
	Env    *env.Environment
	API    wmf.API
	S3     s3iface.S3API
}

// Handler gets all the projects by calling mediawiki actions API, and produces a kafka message for each project.
func Handler(ctx context.Context, p *Parameters, req *pb.ProjectsRequest) (*pb.ProjectsResponse, error) {
	lgs, err := p.API.GetLanguages(ctx, p.Env.DefaultProject)

	if err != nil {
		return nil, err
	}

	res := new(pb.ProjectsResponse)
	sls := []*schema.Language{}
	sps := []*schema.Project{}
	nps := new(strings.Builder)
	nls := new(strings.Builder)

	for _, lng := range lgs {
		for _, prj := range lng.Projects {
			spr := &schema.Project{
				Identifier: prj.DBName,
				Name:       prj.SiteName,
				Code:       prj.Code,
				URL:        prj.URL,
				InLanguage: &schema.Language{
					Identifier: lng.Code,
				},
			}

			dta, err := json.Marshal(spr)

			if err != nil {
				return nil, err
			}

			pin := &s3.PutObjectInput{
				Bucket: aws.String(p.Env.AWSBucket),
				Key:    aws.String(fmt.Sprintf("projects/%s.json", spr.Identifier)),
				Body:   bytes.NewReader(dta),
			}

			if _, err := p.S3.PutObjectWithContext(ctx, pin); err != nil {
				return nil, err
			}

			_, _ = fmt.Fprintf(nps, "%s\n", string(dta))
			sps = append(sps, spr)
		}

		sln := &schema.Language{
			Identifier:    lng.Code,
			Name:          lng.LocalName,
			AlternateName: lng.Name,
			Direction:     lng.Dir,
		}

		dta, err := json.Marshal(sln)

		if err != nil {
			return nil, err
		}

		lin := &s3.PutObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(fmt.Sprintf("languages/%s.json", sln.Identifier)),
			Body:   bytes.NewReader(dta),
		}

		if _, err := p.S3.PutObjectWithContext(ctx, lin); err != nil {
			return nil, err
		}

		_, _ = fmt.Fprintf(nls, "%s\n", string(dta))
		sls = append(sls, sln)
	}

	lin := &s3.PutObjectInput{
		Bucket: aws.String(p.Env.AWSBucket),
		Key:    aws.String("aggregations/languages/languages.ndjson"),
		Body:   strings.NewReader(nls.String()),
	}

	if _, err := p.S3.PutObjectWithContext(ctx, lin); err != nil {
		return nil, err
	}

	cds := `{"identifier":"wiki","name":"Wikipedia","description":"The free encyclopedia."}
{"identifier":"wikibooks","name":"Wikibooks","description":"E-book textbooks and annotated texts."}
{"identifier":"wikinews","name":"Wikinews","description":"The free news source."}
{"identifier":"wikiquote","name":"Wikiquote","description":"Quotes across your favorite books, movies, authors, and more."}
{"identifier":"wikisource","name":"Wikisource","description":"The free digital library."}
{"identifier":"wikivoyage","name":"Wikivoyage","description":"The ultimate worldwide travel guide."}
{"identifier":"wiktionary","name":"Wiktionary","description":"A dictionary for over 170 languages."}
{"identifier":"wikiversity","name":"Wikiversity","description":"Free learning resources."}`

	scn := bufio.NewScanner(strings.NewReader(cds))

	for scn.Scan() {
		cde := map[string]string{}
		cjn := scn.Text()

		if err := json.Unmarshal([]byte(cjn), &cde); err != nil {
			return nil, err
		}

		cin := &s3.PutObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(fmt.Sprintf("codes/%s.json", cde["identifier"])),
			Body:   strings.NewReader(cjn),
		}

		if _, err := p.S3.PutObjectWithContext(ctx, cin); err != nil {
			return nil, err
		}
	}

	cin := &s3.PutObjectInput{
		Bucket: aws.String(p.Env.AWSBucket),
		Key:    aws.String("aggregations/codes/codes.ndjson"),
		Body:   strings.NewReader(cds),
	}

	if _, err := p.S3.PutObjectWithContext(ctx, cin); err != nil {
		return nil, err
	}

	pin := &s3.PutObjectInput{
		Bucket: aws.String(p.Env.AWSBucket),
		Key:    aws.String("aggregations/projects/projects.ndjson"),
		Body:   strings.NewReader(nps.String()),
	}

	if _, err := p.S3.PutObjectWithContext(ctx, pin); err != nil {
		return nil, err
	}

	mgs := []*schema.Message{}

	for _, sln := range sls {
		sln.Event = schema.NewEvent(schema.EventTypeCreate)
		mgs = append(mgs, &schema.Message{
			Config: schema.ConfigLanguage,
			Topic:  p.Env.TopicLanguages,
			Value:  sln,
			Key: &schema.Key{
				Identifier: fmt.Sprintf("languages/%s", sln.Identifier),
				Type:       schema.KeyTypeLanguage,
			},
		})

		res.Total++
	}

	for _, spr := range sps {
		spr.Event = schema.NewEvent(schema.EventTypeCreate)
		mgs = append(mgs, &schema.Message{
			Config: schema.ConfigProject,
			Topic:  p.Env.TopicProjects,
			Value:  spr,
			Key: &schema.Key{
				Identifier: fmt.Sprintf("projects/%s", spr.Identifier),
				Type:       schema.KeyTypeProject,
			},
		})

		res.Total++
	}

	if err := p.Stream.Produce(ctx, mgs...); err != nil {
		return nil, err
	}

	_ = p.Stream.Flush(1000)

	return res, nil
}
