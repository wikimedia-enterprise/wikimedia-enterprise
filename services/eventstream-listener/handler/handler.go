package handler

import (
	"context"
	"fmt"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/operations"
	"wikimedia-enterprise/services/eventstream-listener/submodules/log"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"
	"wikimedia-enterprise/services/eventstream-listener/submodules/tracing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
	"go.uber.org/dig"
)

// Parameters dependency injection parameters for the handler.
type Parameters struct {
	dig.In
	Redis         redis.Cmdable
	Producer      schema.Producer
	KafkaProducer *kafka.Producer
	Env           *env.Environment
	Filter        *filter.Filter
	Tracer        tracing.Tracer
	Operations    *operations.Operations
}

const (
	LastEventTimeKey = "stream:article-change:since"
)

// PageChange handler for the page change stream.
func PageChange(ctx context.Context, p *Parameters) func(evt *eventstream.PageChange) error {
	return func(evt *eventstream.PageChange) error {
		// Check for support of either current or prior namespace. This applies to all events. More event-specific check follows.
		if !p.Filter.IsSupported(evt.Data.Database, evt.Data.Page.PageNamespace, evt.Data.PriorState.Page.PageNamespace) {
			return nil
		}

		eid := uuid.New().String()
		end, trx := p.Tracer.StartTrace(ctx, "eventstream-listener", map[string]string{"event.identifier": eid})
		var err error

		defer func() {
			if err != nil {
				end(err, fmt.Sprintf("error processing %s event with id %s", evt.Data.PageChangeKind, eid))
			} else {
				end(nil, fmt.Sprintf("%s event with id %s processed", evt.Data.PageChangeKind, eid))
			}
		}()

		trx = context.WithValue(trx, operations.EventIdentifierKey, eid)

		msgs, err := p.Operations.Execute(trx, evt)

		if err != nil {
			return err
		}

		err = p.Producer.Produce(trx, msgs...)

		if err != nil {
			log.Error(
				"produce error",
				log.Any("event_id", eid),
				log.Any("event_type", evt.Data.PageChangeKind),
				log.Any("article", evt.Data.Page.PageTitle),
				log.Any("project", evt.Data.Database),
				log.Any("error", err),
			)

			return err
		}

		return p.Redis.Set(ctx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
