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

// Parameters dependencies for the handler.
type Parameters struct {
	dig.In
	Env     *env.Environment
	Stream  schema.UnmarshalProducer
	Tracer  tracing.Tracer
	Metrics *pr.Metrics
}

// NewArticleVisibility produces an article visibility change message in response to a revision visibility change message.
func NewArticleVisibility(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)

		end, trx := p.Tracer.StartTrace(hcx, "article-visibility", map[string]string{})
		var err error
		defer func() {
			if err != nil {
				end(err, "error processing article-visibility event")
			} else {
				end(nil, "article-visbility event processed")
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

		pmg := &schema.Message{
			Config:  schema.ConfigArticle,
			Topic:   p.Env.TopicArticles,
			Value:   art,
			Key:     key,
			Headers: hds,
		}

		return p.Stream.Produce(trx, pmg)
	}
}
