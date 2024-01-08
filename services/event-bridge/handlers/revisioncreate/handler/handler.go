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

type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Env        *env.Environment
	Dictionary langid.Dictionarer
}

// LastEventTimeKey is the redis key where the latest event will be stored.
const LastEventTimeKey = "stream:revision-create:since"

// RevisionCreate handler for the revision create stream.
func RevisionCreate(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.RevisionCreate) error {
	return func(evt *eventstream.RevisionCreate) error {
		if evt.Data.PageIsRedirect || !fr.Projects.IsSupported(evt.Data.Database) || !fr.Namespaces.IsSupported(evt.Data.PageNamespace) {
			return nil
		}

		article := new(schema.Article)
		article.Event = schema.NewEvent(schema.EventTypeUpdate)
		article.Identifier = evt.Data.PageID
		article.Name = evt.Data.PageTitle
		article.DateModified = &evt.Data.RevTimestamp
		article.PreviousVersion = &schema.PreviousVersion{
			Identifier: evt.Data.RevParentID,
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
		article.IsPartOf = &schema.Project{
			Identifier: evt.Data.Database,
			URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
		}
		article.Namespace = &schema.Namespace{
			Identifier: evt.Data.PageNamespace,
		}

		langid, err := p.Dictionary.GetLanguage(ctx, evt.Data.Database)
		if err != nil {
			return err
		}
		article.InLanguage = &schema.Language{
			Identifier: langid,
		}

		url, err := url.QueryUnescape(evt.Data.Meta.URI)

		if err != nil {
			return err
		}

		article.URL = url

		err = p.Producer.Produce(ctx, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticleUpdate,
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
