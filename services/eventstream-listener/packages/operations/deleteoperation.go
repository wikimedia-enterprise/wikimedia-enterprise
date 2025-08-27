package operations

import (
	"context"
	"fmt"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/transformer"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"
	"wikimedia-enterprise/services/eventstream-listener/submodules/tracing"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

// NewDeleteOperation returns DeleteOperation instance.
func NewDeleteOperation(trs *transformer.Transforms, fr *filter.Filter, env *env.Environment) Operation {
	return &DeleteOperation{
		Transfer: trs,
		Filter:   fr,
		Env:      env,
	}
}

// DeleteOperation type to implement delete event processing.
type DeleteOperation struct {
	Env      *env.Environment
	Filter   *filter.Filter
	Transfer *transformer.Transforms
}

// Execute implements delete event processing.
func (u *DeleteOperation) Execute(ctx context.Context, evt *eventstream.PageChange) ([]*schema.Message, error) {
	msgs := []*schema.Message{}

	// Narrow down namespace and handle redirect
	if !u.Filter.IsSupported(evt.Data.Database, evt.Data.Page.PageNamespace) || evt.Data.Page.PageIsRedirect {
		return msgs, nil
	}

	evtType := schema.EventTypeDelete
	art, err := u.Transfer.EventToArticle(ctx, evt)

	if err != nil {
		return msgs, err
	}

	event := schema.NewEvent(evtType)
	event.SetIdentifier(ctx.Value(EventIdentifierKey).(string))
	art.Event = event

	topics := u.Env.OutputTopics.GetNamesByEventType(evt.Data.Database, evtType)

	for _, topic := range topics {
		msgs = append(msgs, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  topic,
			Value:  art,
			Key: &schema.Key{
				Identifier: fmt.Sprintf("/%s/%s", art.IsPartOf.Identifier, art.Name),
				Type:       schema.KeyTypeArticle,
			},
			Headers: tracing.NewHeadersCarrier().InjectContext(ctx),
		})
	}

	return msgs, nil
}
