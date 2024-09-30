package handler

import (
	"context"
	"fmt"
	"slices"
	"time"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/common"
	"wikimedia-enterprise/services/event-bridge/packages/filter"
	"wikimedia-enterprise/services/event-bridge/packages/operations"

	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
	"go.uber.org/dig"
)

// Parameters dependency injection parameters for the handler.
type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Env        *env.Environment
	Dictionary langid.Dictionarer
	Tracer     tracing.Tracer
	Operations *operations.Operations
}

// LastEventTimeKey is the redis key where the latest event will be stored.
const LastEventTimeKey = "stream:article-change:since"

// PageCreate handler for the page create stream.
func PageChange(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.PageChange) error {
	return func(evt *eventstream.PageChange) error {
		if evt.Data.Page.PageIsRedirect || !slices.Contains(common.SupportedProjects, evt.Data.Database) || !slices.Contains(common.SupportedNamespaces, evt.Data.Page.PageNamespace) {
			return nil
		}

		eid := uuid.New().String()

		end, trx := p.Tracer.StartTrace(ctx, "page-change", map[string]string{"event.identifier": eid})
		var err error
		defer func() {
			if err != nil {
				end(err, fmt.Sprintf("error processing page-create event with id %s", eid))
			} else {
				end(nil, fmt.Sprintf("page-create event with id %s processed", eid))
			}
		}()

		trx = context.WithValue(trx, common.EventIdentifierContextKey, eid)

		opm := &operations.OperationParameters{
			Producer:   p.Producer,
			Env:        p.Env,
			Dictionary: p.Dictionary,
			Tracer:     p.Tracer,
		}

		tpcs, art, err := p.Operations.Execute(trx, opm, evt)

		if err != nil {
			return err
		}

		hds := tracing.NewHeadersCarrier().InjectContext(ctx)

		for _, tpc := range tpcs {

			log.Info("sending message to kafka", log.Any("tpc", tpc), log.Any("identifier", fmt.Sprintf("/%s/%s", art.IsPartOf.Identifier, art.Name)))

			err = p.Producer.Produce(trx, &schema.Message{
				Config: schema.ConfigArticle,
				Topic:  tpc,
				Value:  art,
				Key: &schema.Key{
					Identifier: fmt.Sprintf("/%s/%s", art.IsPartOf.Identifier, art.Name),
					Type:       schema.KeyTypeArticle,
				},
				Headers: hds,
			})

			if err != nil {
				log.Error(
					"producer produce error",
					log.Any("event_id", art.Event.Identifier),
					log.Any("revision", art.Version.Identifier),
					log.Any("error", err),
				)

				return err
			}
		}

		return p.Redis.Set(ctx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}

}
