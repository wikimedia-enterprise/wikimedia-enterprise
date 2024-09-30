package operations

import (
	"context"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/packages/common"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

func newBaseOperation() Operation {
	return &baseOperation{}
}

type baseOperation struct {
}

func (co *baseOperation) Execute(ctx context.Context, prm *OperationParameters, evt *eventstream.PageChange, art *schema.Article) ([]string, *schema.Article, error) {

	var tp string

	switch evt.Data.PageChangeKind {
	case Create:
		tp = schema.EventTypeCreate
	case UnDelete:
		tp = schema.EventTypeCreate
	case Update:
		tp = schema.EventTypeUpdate
	case Delete:
		tp = schema.EventTypeDelete
	case VisibilityChange:
		tp = schema.EventTypeVisibilityChange
	}

	art.Event = schema.NewEvent(tp)
	art.Event.Identifier = ctx.Value(common.EventIdentifierContextKey).(string)

	return prm.Env.OutputTopics.GetNamesByEvent(evt.Data.Database, tp), art, nil
}
