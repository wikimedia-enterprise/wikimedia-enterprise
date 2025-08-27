package transformer

import (
	"context"
	"fmt"
	"net/url"
	"wikimedia-enterprise/services/eventstream-listener/packages/builder"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/submodules/log"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

// TransformEventToArticle interface to convert event to article.
type TransformEventToArticle interface {
	EventToArticle(ctx context.Context, evt *eventstream.PageChange) (*schema.Article, error)
}

// New returns a Transforms instance.
func New(fr *filter.Filter) *Transforms {
	return &Transforms{
		Filter: fr,
	}
}

// Transforms implements article creation from event.
type Transforms struct {
	Filter *filter.Filter
}

// EventToArticle returns an article using event's fields.
func (t *Transforms) EventToArticle(ctx context.Context, evt *eventstream.PageChange) (*schema.Article, error) {
	lng, err := t.Filter.ResolveLang(evt.Data.Database)

	if err != nil {
		return nil, err
	}

	edr := builder.NewEditorBuilder().
		Identifier(evt.Data.Performer.UserID).
		Name(evt.Data.Performer.UserText).
		EditCount(evt.Data.Performer.UserEditCount).
		Groups(evt.Data.Performer.UserGroups).
		IsBot(evt.Data.Performer.UserIsBot).
		DateStarted(&evt.Data.Performer.UserRegistrationDt).
		Build()

	ver := builder.NewVersionBuilder().
		Identifier(evt.Data.Revision.RevID).
		Comment(evt.Data.Revision.Comment).
		IsMinorEdit(evt.Data.Revision.IsMinorEdit).
		Size(evt.Data.Revision.RevSize).
		Editor(edr).
		Build()

	url, err := url.QueryUnescape(evt.Data.Meta.URI)

	if err != nil {
		log.Error(
			"url query unescape error",
			log.Any("revision_id", evt.Data.Revision.RevID),
			log.Any("title", evt.Data.Page.PageTitle),
			log.Any("url", evt.Data.Meta.URI),
			log.Any("error", err),
		)

		return nil, err
	}

	art := builder.NewArticleBuilder().
		Identifier(evt.Data.Page.PageID).
		Name(evt.Data.Page.PageTitle).
		DateModified(&evt.Data.Revision.RevDt).
		IsPartOf(&schema.Project{
			Identifier: evt.Data.Database,
			URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
		}).
		Namespace(&schema.Namespace{Identifier: evt.Data.Page.PageNamespace}).
		InLanguage(&schema.Language{
			Identifier: lng,
		}).
		PreviousVersion(evt.Data.PriorState.Revision.RevID).
		URL(url).
		Version(ver).
		Build()

	return art, nil
}
