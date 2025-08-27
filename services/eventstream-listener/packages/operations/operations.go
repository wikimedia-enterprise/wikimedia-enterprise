package operations

import (
	"context"
	"fmt"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/transformer"
	"wikimedia-enterprise/services/eventstream-listener/submodules/log"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

const (
	// Event types
	Create           = "create"
	Undelete         = "undelete"
	Update           = "edit"
	Delete           = "delete"
	Move             = "move"
	VisibilityChange = "visibility_change"

	// Event id key
	EventIdentifierKey EventIdentifierType = "event.identifier"
)

// EventIdentifierType for masking the type to propagate in context.
type EventIdentifierType string

// Operation interface provides signature for event processing helpers.
type Operation interface {
	Execute(ctx context.Context, evt *eventstream.PageChange) ([]*schema.Message, error)
}

// New returns an Operations instance.
func New(trs *transformer.Transforms, fr *filter.Filter, env *env.Environment) *Operations {
	ops := new(Operations)

	ops.OperationsMap = map[string]Operation{
		Update:           NewUpdateOperation(trs, fr, env),
		Create:           NewCreateOperation(trs, fr, env),
		Undelete:         NewCreateOperation(trs, fr, env),
		Delete:           NewDeleteOperation(trs, fr, env),
		VisibilityChange: NewVisibilityOperation(trs, fr, env),
		Move:             NewMoveOperation(trs, fr, env),
	}

	return ops
}

// Operations type to implement event processing method.
type Operations struct {
	OperationsMap map[string]Operation
}

// Execute implements high level event processing.
func (o *Operations) Execute(ctx context.Context, evt *eventstream.PageChange) ([]*schema.Message, error) {
	if _, exists := o.OperationsMap[evt.Data.PageChangeKind]; !exists {
		log.Error(
			"event not supported",
			log.Any("event_type", evt.Data.PageChangeKind),
			log.Any("article", evt.Data.Page.PageTitle),
			log.Any("page_id", evt.Data.Page.PageID),
			log.Any("namespace", evt.Data.Page.PageNamespace),
			log.Any("revision", evt.Data.Revision.RevID),
			log.Any("project", evt.Data.Database),
		)

		return []*schema.Message{}, fmt.Errorf("unsupported event type %s", evt.Data.PageChangeKind)
	}

	return o.OperationsMap[evt.Data.PageChangeKind].Execute(ctx, evt)
}
