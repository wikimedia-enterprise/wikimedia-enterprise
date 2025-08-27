// Package handler creates a handler that will keep articles in s3 updated.
package handler

import (
	"context"
	"fmt"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/libraries/storage"
	pr "wikimedia-enterprise/services/on-demand/submodules/prometheus"
	"wikimedia-enterprise/services/on-demand/submodules/schema"
	"wikimedia-enterprise/services/on-demand/submodules/subscriber"
	"wikimedia-enterprise/services/on-demand/submodules/tracing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
	"golang.org/x/time/rate"
)

// Parameters dependency injection for articles handler.
type Parameters struct {
	dig.In
	Stream  schema.UnmarshalProducer
	Storage storage.Updater
	Tracer  tracing.Tracer
	Metrics *pr.Metrics
	Env     *env.Environment
}

// New create new articles handler.
func New(p *Parameters) subscriber.Handler {
	limiter := rate.NewLimiter(rate.Limit(p.Env.MaxMsgsPerSecond), 1)

	return func(ctx context.Context, msg *kafka.Message) error {
		if err := limiter.Wait(ctx); err != nil {
			return fmt.Errorf("error throttling: %w", err)
		}

		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)

		art := new(schema.Article)

		p.Metrics.Inc(pr.OdmTtlEvents)
		if err := p.Stream.Unmarshal(hcx, msg.Value, art); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}

		err := p.Storage.Update(hcx, msg.Key, art.Event.Type, art)
		if err != nil {
			return fmt.Errorf("error uploading message: %w", err)
		}

		return nil
	}
}
