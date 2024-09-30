package operations

import (
	"context"
	"fmt"
	"slices"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/services/event-bridge/packages/common"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

func newMoveOperation() Operation {
	return &moveOperation{}
}

type moveOperation struct {
}

func (co *moveOperation) Execute(ctx context.Context, prm *OperationParameters, evt *eventstream.PageChange, art *schema.Article) ([]string, *schema.Article, error) {

	if slices.Contains(common.SupportedNamespaces, evt.Data.PriorState.Page.PageNamespace) {
		art.Event = schema.NewEvent(schema.EventTypeDelete)
		art.Event.Identifier = ctx.Value(common.EventIdentifierContextKey).(string)

		hds := tracing.NewHeadersCarrier().InjectContext(ctx)

		tpcs := prm.Env.OutputTopics.GetNamesByEvent(evt.Data.Database, Delete)

		for _, tpc := range tpcs {
			err := prm.Producer.Produce(ctx, &schema.Message{
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
					log.Any("error", err),
					log.Any("event_id", ctx.Value(common.EventIdentifierContextKey).(string)),
					log.Any("revision", art.Version.Identifier),
				)

				return nil, nil, err
			}
		}
	}

	art.Event = schema.NewEvent(schema.EventTypeMove)
	art.Event.Identifier = ctx.Value(common.EventIdentifierContextKey).(string)

	return prm.Env.OutputTopics.GetNamesByEvent(evt.Data.Database, schema.EventTypeMove), art, nil

}
