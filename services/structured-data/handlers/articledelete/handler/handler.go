// Package handler contains article delete handler for the structured data service.
package handler

import (
	"context"
	"net/url"
	"strings"
	"time"
	"wikimedia-enterprise/general/log"
	pr "wikimedia-enterprise/general/prometheus"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"go.uber.org/dig"
)

// Parameters is a dependency injection params for the handler.
type Parameters struct {
	dig.In
	Stream  schema.UnmarshalProducer
	Env     *env.Environment
	API     wmf.API
	Tracer  tracing.Tracer
	Metrics *pr.Metrics
}

// NewArticleDelete verifies that the page was deleted.
// If the mediawiki API call fails, it produces to the error or dead letter topic based on the fail count.
func NewArticleDelete(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)

		end, trx := p.Tracer.StartTrace(hcx, "article-delete", map[string]string{})
		var err error
		defer func() {
			if err != nil {
				end(err, "error processing article-delete event")
			} else {
				end(nil, "article-delete event processed")
			}
		}()

		key := new(schema.Key)

		if err := p.Stream.Unmarshal(trx, msg.Key, key); err != nil {
			return err
		}

		art := new(schema.Article)

		if err := p.Stream.Unmarshal(trx, msg.Value, art); err != nil {
			return err
		}

		dtb := art.IsPartOf.Identifier
		tcs, err := p.Env.Topics.
			GetNames(dtb, art.Namespace.Identifier)

		if err != nil {
			return err
		}

		// props to delete to make request simpler and faster
		pdl := []string{
			"rvprop",
			"rvslots",
			"inprop",
			"ppprop",
			"rdlimit",
			"wbeulimit",
		}

		pge, err := p.API.GetPage(trx, art.IsPartOf.Identifier, art.Name, func(v *url.Values) {
			for _, prp := range pdl {
				v.Del(prp)
			}

			// limiting down to only basic information about the page
			// and redirects to catch page-move events that leave behind redirects
			v.Set("prop", "info|redirects")
		})

		dtn := time.Now().UTC()
		art.Event.SetDatePublished(&dtn)

		if dur := dtn.Sub(*art.Event.DateCreated); dur.Milliseconds() > p.Env.LatencyThresholdMS {
			log.Warn("latency threshold exceeded",
				log.Any("name", art.Name),
				log.Any("url", art.URL),
				log.Any("revision", art.Version.Identifier),
				log.Any("language", art.InLanguage.Identifier),
				log.Any("namespace", art.Namespace.Identifier),
				log.Any("event_id", art.Event.Identifier),
				log.Any("duration", dur.Milliseconds()),
			)
		}

		hds := tracing.NewHeadersCarrier().InjectContext(trx)

		mgs := []*schema.Message{
			{
				Config:  schema.ConfigArticle,
				Topic:   p.Env.TopicArticles,
				Value:   art,
				Key:     key,
				Headers: hds,
			},
		}

		for _, tpc := range tcs {
			mgs = append(mgs, &schema.Message{
				Config:  schema.ConfigArticle,
				Topic:   tpc,
				Value:   art,
				Key:     key,
				Headers: hds,
			})
		}

		if err == wmf.ErrPageNotFound {
			return p.Stream.Produce(trx, mgs...)
		}

		if pge != nil && pge.Missing {
			return p.Stream.Produce(trx, mgs...)
		}

		// normalize title for comparison
		ntl := strings.ToLower(strings.ReplaceAll(art.Name, "_", " "))

		// if page-move occurred for pge.Title and redirect is present, titles will be different and
		// a pge.Title should be in the redirects array
		if pge != nil && strings.ToLower(pge.Title) != ntl {
			lgf := []log.Field{
				log.Any("name", art.Name),
				log.Any("url", art.URL),
				log.Any("revision", art.Version.Identifier),
				log.Any("language", art.InLanguage.Identifier),
				log.Any("namespace", art.Namespace.Identifier),
				log.Any("event_id", art.Event.Identifier),
			}

			for _, red := range pge.Redirects {
				if strings.ToLower(red.Title) == ntl {
					log.Info("redirect left behind after page-move", lgf...)
					return p.Stream.Produce(trx, mgs...)
				}
			}

			log.Warn("page title does not match redirect title, but no redirect found", lgf...)
		}

		return err
	}
}
