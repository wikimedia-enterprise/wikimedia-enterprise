// Package handler creates a handler that will keep articles in s3 updated.
package handler

import (
	"context"
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

// Parameters dependency injection for structured handler.
type Parameters struct {
	dig.In
	Stream  schema.UnmarshalProducer
	Storage storage.Updater
	Tracer  tracing.Tracer
	Metrics *pr.Metrics
	Env     *env.Environment
}

// New create new structured handler.
func New(p *Parameters) subscriber.Handler {
	limiter := rate.NewLimiter(rate.Limit(p.Env.MaxMsgsPerSecond), 1)

	return func(ctx context.Context, msg *kafka.Message) error {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}

		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)

		str := new(schema.AvroStructured)

		p.Metrics.Inc(pr.OdmTtlEvents)
		if err := p.Stream.Unmarshal(hcx, msg.Value, str); err != nil {
			return err
		}

		val, err := str.ToJsonStruct()

		if err != nil {
			return err
		}

		return p.Storage.Update(hcx, msg.Key, val.Event.Type, val)
	}
}
