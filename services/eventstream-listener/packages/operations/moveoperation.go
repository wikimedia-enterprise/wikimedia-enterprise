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

// NewMoveOperation returns an instance of MoveOperation.
func NewMoveOperation(trs *transformer.Transforms, fr *filter.Filter, env *env.Environment) Operation {
	return &MoveOperation{
		Transfer: trs,
		Filter:   fr,
		Env:      env,
	}
}

// MoveOperation type to implement move event processing.
type MoveOperation struct {
	Env      *env.Environment
	Filter   *filter.Filter
	Transfer *transformer.Transforms
}

// Execute implements move event processing.
func (u *MoveOperation) Execute(ctx context.Context, evt *eventstream.PageChange) ([]*schema.Message, error) {
	msgs := []*schema.Message{}

	art, err := u.Transfer.EventToArticle(ctx, evt)

	if err != nil {
		return msgs, err
	}

	// Narrow down to determine whether to publish move event
	if u.Filter.IsSupported(evt.Data.Database, evt.Data.Page.PageNamespace) && !evt.Data.Page.PageIsRedirect {

		evtType := schema.EventTypeMove
		event := schema.NewEvent(evtType)
		event.SetIdentifier(ctx.Value(EventIdentifierKey).(string))
		art.Event = event

		topics := u.Env.OutputTopics.GetNamesByEventType(evt.Data.Database, evtType)
		// Move messages
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
	}

	// We always want to publish delete event if prior namespace was supported.
	if u.Filter.IsSupported(evt.Data.Database, evt.Data.PriorState.Page.PageNamespace) {
		evtType := schema.EventTypeDelete
		event := schema.NewEvent(evtType)
		event.SetIdentifier(ctx.Value(EventIdentifierKey).(string))

		priorArticle := &schema.Article{
			Event: event,
			Name:  evt.Data.PriorState.Page.PageTitle,
			Namespace: &schema.Namespace{
				Identifier: evt.Data.PriorState.Page.PageNamespace,
			},
			Identifier:   art.Identifier,
			DateModified: art.DateModified,
			IsPartOf:     art.IsPartOf,
			InLanguage:   art.InLanguage,
			URL:          art.URL,
			Version:      art.Version,
		}

		topics := u.Env.OutputTopics.GetNamesByEventType(evt.Data.Database, evtType)
		// Delete messages
		for _, topic := range topics {
			msgs = append(msgs, &schema.Message{
				Config: schema.ConfigArticle,
				Topic:  topic,
				Value:  priorArticle,
				Key: &schema.Key{
					Identifier: fmt.Sprintf("/%s/%s", priorArticle.IsPartOf.Identifier, priorArticle.Name),
					Type:       schema.KeyTypeArticle,
				},
				Headers: tracing.NewHeadersCarrier().InjectContext(ctx),
			})
		}
	}

	return msgs, nil
}
