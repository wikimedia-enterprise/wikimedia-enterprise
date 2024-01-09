// Package articles implements the gRPC handler for articles. The handler produces a kafka message with article names (max 50 names per message)
// for a given namespace per project.
// The API response from this handler includes the total number of messages processed and the total number of errors during processing.
package articles

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"

	"go.uber.org/dig"
)

// Parameters dependencies for the handler.
type Parameters struct {
	dig.In
	Stream schema.UnmarshalProducer
	Client wmf.API
	Env    *env.Environment
}

func chunk[T any](its []T, sze int) (chn [][]T) {
	for sze < len(its) {
		its, chn = its[sze:], append(chn, its[0:sze:sze])
	}

	return append(chn, its)
}

// Handler gets all the article titles by calling mediawiki all pages API.
// Then, it splits pages to chunks of max 50 article names
// and produces a kafka message for each chunk.
func Handler(ctx context.Context, p *Parameters, req *pb.ArticlesRequest) (*pb.ArticlesResponse, error) {
	res := new(pb.ArticlesResponse)

	prj, err := p.Client.GetProject(ctx, req.GetProject())

	if err != nil {
		return nil, err
	}

	cbk := func(pgs []*wmf.Page) {
		for _, cpg := range chunk(pgs, 50) {
			nms := []string{}

			for _, pge := range cpg {
				nms = append(nms, pge.Title)
			}

			msg := &schema.Message{
				Config: schema.ConfigArticleNames,
				Topic:  p.Env.TopicArticles,
				Key: &schema.Key{
					Identifier: fmt.Sprintf("article-names/%s/%d/%s", req.Project, req.Namespace, strings.Join(nms, "|")),
					Type:       schema.KeyTypeArticleNames,
				},
				Value: &schema.ArticleNames{
					Event: schema.NewEvent(schema.EventTypeCreate),
					Names: nms,
					IsPartOf: &schema.Project{
						Identifier: prj.DBName,
						URL:        prj.URL,
					},
				},
			}

			if err := p.Stream.Produce(ctx, msg); err != nil {
				res.Errors++
				log.Println(err)
				continue
			}

			res.Total++
		}
	}

	opt := func(v *url.Values) {
		v.Set("apnamespace", strconv.FormatInt(int64(req.Namespace), 10))
	}

	if err := p.Client.GetAllPages(ctx, req.Project, cbk, opt); err != nil {
		return nil, err
	}

	return res, nil
}
