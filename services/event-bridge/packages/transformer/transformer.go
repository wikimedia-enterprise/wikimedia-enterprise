package transformer

import (
	"context"
	"fmt"
	"net/url"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/builder"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

type TransformPageChangeToArticle interface {
	PageChangeToArticle(ctx context.Context, dic langid.Dictionarer, evt *eventstream.PageChange) (*schema.Article, error)
}

type Transformations interface {
	TransformPageChangeToArticle
}

type Transforms struct {
	Transformations
}

func New() Transformations {
	return &Transforms{}
}

func (t *Transforms) PageChangeToArticle(ctx context.Context, dic langid.Dictionarer, evt *eventstream.PageChange) (*schema.Article, error) {

	edr := builder.NewEditorBuilder().
		Identifier(evt.Data.Performer.UserID).
		Name(evt.Data.Performer.UserText).
		EditCount(evt.Data.Performer.UserEditCount).
		Groups(evt.Data.Performer.UserGroups).
		IsBot(evt.Data.Performer.UserIsBot).
		DateStarted(&evt.Data.Performer.UserRegistrationDt).Build()

	ver := builder.NewVersionBuilder().
		Identifier(evt.Data.Revision.RevID).
		Comment(evt.Data.Revision.Comment).
		IsMinorEdit(evt.Data.Revision.IsMinorEdit).
		Size(evt.Data.Revision.RevSize).
		Editor(edr).
		Build()

	lng, err := dic.GetLanguage(ctx, evt.Data.Database)

	if err != nil {
		if evt.Data.Database == "commonswiki" {
			lng = "en"
		} else {
			log.Error(
				"dictionary get language error",
				log.Any("database", evt.Data.Database),
				log.Any("event_id", evt.Data.Page.PageID),
				log.Any("error", err),
			)

			return nil, err
		}

	}

	url, err := url.QueryUnescape(evt.Data.Meta.URI)

	if err != nil {
		log.Error(
			"url query unescape error",
			log.Any("event_id", evt.Data.Page.PageID),
			log.Any("error", err),
			log.Any("url", evt.Data.Meta.URI),
		)

		return nil, err
	}

	art := builder.NewArticleBuilder(fmt.Sprintf("https://%s", evt.Data.Meta.Domain)).
		Identifier(evt.Data.Page.PageID).
		Name(evt.Data.Page.PageTitle).
		DateCreated(&evt.Data.Revision.RevDt).
		DateModified(&evt.Data.Revision.RevDt).
		IsPartOf(&schema.Project{
			Identifier: evt.Data.Database,
			URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
		}).
		Namespace(&schema.Namespace{Identifier: evt.Data.Page.PageNamespace}).
		InLanguage(&schema.Language{
			Identifier: lng,
		}).
		Visibility(&schema.Visibility{Text: evt.Data.Revision.IsContentVisible,
			Editor:  evt.Data.Revision.IsEditorVisible,
			Comment: evt.Data.Revision.IsCommentVisible}).
		URL(url).
		Version(ver).Build()

	return art, nil
}
