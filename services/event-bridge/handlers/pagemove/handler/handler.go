package handler

import (
	"context"
	"fmt"
	"net/url"
	"time"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
	"go.uber.org/dig"

	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/filter"

	redis "github.com/redis/go-redis/v9"
)

// LastEventTimeKey is the redis key where the latest event will be stored.
const LastEventTimeKey = "stream:page-move:since"

// Parameters required for handler.
type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Dictionary langid.Dictionarer
	Env        *env.Environment
}

// PageMove handler for the page move stream. This handler treats a pagemove event
// as if it was a pagedelete event.
func PageMove(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.PageMove) error {
	return func(evt *eventstream.PageMove) error {
		if evt.Data.PageIsRedirect || !fr.Projects.IsSupported(evt.Data.Database) {
			return nil
		}

		langid, err := p.Dictionary.GetLanguage(ctx, evt.Data.Database)
		if err != nil {
			return err
		}

		dte := time.Now()

		article := &schema.Article{
			Event:        schema.NewEvent(schema.EventTypeMove),
			Identifier:   evt.Data.PageID,
			Name:         evt.Data.PriorState.PageTitle,
			DateCreated:  &dte,
			DateModified: &dte,
			Version: &schema.Version{
				Identifier: evt.Data.RevID,
				Comment:    evt.Data.Comment,
				Editor: &schema.Editor{
					Identifier:  evt.Data.Performer.UserID,
					Name:        evt.Data.Performer.UserText,
					EditCount:   evt.Data.Performer.UserEditCount,
					Groups:      evt.Data.Performer.UserGroups,
					IsBot:       evt.Data.Performer.UserIsBot,
					DateStarted: &evt.Data.Performer.UserRegistrationDt,
				},
			},
			IsPartOf: &schema.Project{
				Identifier: evt.Data.Database,
				URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
			},
			Namespace: &schema.Namespace{
				Identifier: evt.Data.PriorState.PageNamespace,
			},
			InLanguage: &schema.Language{
				Identifier: langid,
			},
		}

		url, err := url.QueryUnescape(evt.Data.Meta.URI)

		if err != nil {
			return err
		}

		article.URL = url

		err = p.Producer.Produce(ctx, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticleMove,
			Value:  article,
			Key: &schema.Key{
				Identifier: fmt.Sprintf("/%s/%s", article.IsPartOf.Identifier, article.Name),
				Type:       schema.KeyTypeArticle,
			},
		})

		if err != nil {
			return err
		}

		if fr.Namespaces.IsSupported(evt.Data.PriorState.PageNamespace) {
			article.Event = schema.NewEvent(schema.EventTypeDelete)
			err = p.Producer.Produce(ctx, &schema.Message{
				Config: schema.ConfigArticle,
				Topic:  p.Env.TopicArticleDelete,
				Value:  article,
				Key: &schema.Key{
					Identifier: fmt.Sprintf("/%s/%s", article.IsPartOf.Identifier, article.Name),
					Type:       schema.KeyTypeArticle,
				},
			})

			if err != nil {
				return err
			}
		}

		return p.Redis.Set(ctx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
