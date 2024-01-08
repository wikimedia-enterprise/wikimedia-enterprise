package handler

import (
	"context"
	"time"

	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/structured-data/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
)

// Parameters dependencies for the handler.
type Parameters struct {
	dig.In
	Env    *env.Environment
	Stream schema.UnmarshalProducer
}

// NewArticleVisibility produces an article visibility change message in response to a revision visibility change message.
func NewArticleVisibility(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		key := new(schema.Key)

		if err := p.Stream.Unmarshal(ctx, msg.Key, key); err != nil {
			return err
		}

		art := new(schema.Article)

		if err := p.Stream.Unmarshal(ctx, msg.Value, art); err != nil {
			return err
		}

		dtn := time.Now().UTC()
		art.Event.SetDatePublished(&dtn)

		if dur := dtn.Sub(*art.Event.DateCreated); dur.Milliseconds() > p.Env.LatencyThresholdMS {
			log.Warn("latency threshold exceeded",
				log.Any("name", art.Name),
				log.Any("project", art.Namespace),
				log.Any("URL", art.URL),
				log.Any("identifier", art.Identifier),
				log.Any("version", art.Version.Identifier),
				log.Any("duration", dur.Milliseconds()),
			)
		}

		pmg := &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticles,
			Value:  art,
			Key:    key,
		}

		return p.Stream.Produce(ctx, pmg)
	}
}
