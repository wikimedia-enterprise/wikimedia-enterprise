// Package handler contains article delete handler for the structured data service.
package handler

import (
	"context"
	"time"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/submodules/log"
	pr "wikimedia-enterprise/services/structured-data/submodules/prometheus"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/subscriber"
	"wikimedia-enterprise/services/structured-data/submodules/tracing"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"go.uber.org/dig"
)

// Parameters is a dependency injection params for the handler.
type Parameters struct {
	dig.In
	Stream  schema.UnmarshalProducer
	Env     *env.Environment
	Tracer  tracing.Tracer
	Metrics *pr.Metrics
}

// NewArticleDelete propagates delete event.
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

		dtn := time.Now().UTC()
		art.Event.SetDatePublished(&dtn)

		if art.Event.IsInternal != nil && *art.Event.IsInternal {
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

		return p.Stream.Produce(trx, mgs...)
	}
}
