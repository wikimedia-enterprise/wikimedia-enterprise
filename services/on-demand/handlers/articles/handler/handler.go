// Package handler creates a handler that will keep articles in s3 updated.
package handler

import (
	"context"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/on-demand/libraries/storage"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
)

// Parameters dependency injection for articles handler.
type Parameters struct {
	dig.In
	Stream  schema.UnmarshalProducer
	Storage storage.Updater
}

// New create new articles handler.
func New(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		art := new(schema.Article)

		if err := p.Stream.Unmarshal(ctx, msg.Value, art); err != nil {
			return err
		}

		return p.Storage.Update(ctx, msg.Key, art.Event.Type, art)
	}
}
