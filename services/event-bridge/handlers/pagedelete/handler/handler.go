package handler

import (
	"context"
	"fmt"
	"net/url"
	"time"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/filter"

	redis "github.com/redis/go-redis/v9"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
	"go.uber.org/dig"
)

// LastEventTimeKey last event time location in redis.
const LastEventTimeKey = "stream:page-delete:since"

// Parameters required for handler.
type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Dictionary langid.Dictionarer
	Env        *env.Environment
}

// Pagedelete parses a pagedelete event, creates a kafka message based on article schema, writes to TopicEventBridgeArticleDelete topic, and stores the processed event timestamp.
func PageDelete(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.PageDelete) error {
	return func(evt *eventstream.PageDelete) error {
		if evt.Data.PageIsRedirect || !fr.Namespaces.IsSupported(evt.Data.PageNamespace) || !fr.Projects.IsSupported(evt.Data.Database) {
			return nil
		}

		article := new(schema.Article)
		article.Event = schema.NewEvent(schema.EventTypeDelete)
		article.Name = evt.Data.PageTitle
		article.Identifier = evt.Data.PageID
		article.Namespace = &schema.Namespace{
			Identifier: evt.Data.PageNamespace,
		}
		article.IsPartOf = &schema.Project{
			Identifier: evt.Data.Database,
			URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
		}
		article.Version = &schema.Version{
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
		}

		url, err := url.QueryUnescape(evt.Data.Meta.URI)

		if err != nil {
			return err
		}

		article.URL = url

		langid, err := p.Dictionary.GetLanguage(ctx, evt.Data.Database)
		if err != nil {
			return err
		}
		article.InLanguage = &schema.Language{
			Identifier: langid,
		}

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

		return p.Redis.Set(ctx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
