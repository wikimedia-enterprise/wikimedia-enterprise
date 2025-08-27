// Package aggregate provides a handler to collect and store article data in redis.
package aggregate

import (
	"context"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"
	"wikimedia-enterprise/services/content-integrity/submodules/schema"
	"wikimedia-enterprise/services/content-integrity/submodules/subscriber"

	"wikimedia-enterprise/services/content-integrity/submodules/log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
)

// Parameters represents dependency injection and
// logic for the aggregate handler.
type Parameters struct {
	dig.In
	Stream    schema.UnmarshalProducer
	Collector collector.ArticleAPI
}

// NewAggregate method creates new aggregation handler.
func NewAggregate(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		art := new(schema.Article)

		if err := p.Stream.Unmarshal(ctx, msg.Value, art); err != nil {
			log.Error(err)
			return err
		}

		switch art.Event.Type {
		case schema.EventTypeCreate:
			if art.Namespace.Identifier == 0 {
				if _, err := p.Collector.SetDateCreated(ctx, art.IsPartOf.Identifier, art.Identifier, art.DateCreated); err != nil {
					log.Error(err)
					return err
				}

				ner := p.Collector.SetName(ctx, art.IsPartOf.Identifier, art.Identifier, art.Name)
				if ner != nil {
					log.Error(ner)
					return ner
				}
			}
		case schema.EventTypeUpdate:
			vrn := &collector.Version{
				Identifier:  art.Version.Identifier,
				DateCreated: art.DateModified,
				Editor:      art.Version.Editor.Name,
			}

			if _, err := p.Collector.PrependVersion(ctx, art.IsPartOf.Identifier, art.Identifier, vrn); err != nil {
				log.Error(err)
				return err
			}

		case schema.EventTypeMove:
			if art.Namespace.Identifier == 0 {
				if _, err := p.Collector.SetDateNamespaceMoved(ctx, art.IsPartOf.Identifier, art.Identifier, art.DateModified); err != nil {
					log.Error(err)
					return err
				}

				ner := p.Collector.SetName(ctx, art.IsPartOf.Identifier, art.Identifier, art.Name)
				if ner != nil {
					log.Error(ner)
					return ner
				}
			}
		}

		return nil
	}
}
