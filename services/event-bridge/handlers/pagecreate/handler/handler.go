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

// Parameters dependency injection parameters for the handler.
type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Env        *env.Environment
	Dictionary langid.Dictionarer
}

// LastEventTimeKey is the redis key where the latest event will be stored.
const LastEventTimeKey = "stream:article-create:since"

// PageCreate handler for the page create stream.
func PageCreate(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.PageCreate) error {
	return func(evt *eventstream.PageCreate) error {
		if evt.Data.PageIsRedirect || !fr.Projects.IsSupported(evt.Data.Database) || !fr.Namespaces.IsSupported(evt.Data.PageNamespace) {
			return nil
		}

		art := new(schema.Article)
		art.Event = schema.NewEvent(schema.EventTypeCreate)
		art.Identifier = evt.Data.PageID
		art.Name = evt.Data.PageTitle
		art.DateCreated = &evt.Data.RevTimestamp
		art.DateModified = &evt.Data.RevTimestamp

		art.Version = &schema.Version{
			Identifier:  evt.Data.RevID,
			Comment:     evt.Data.Comment,
			IsMinorEdit: evt.Data.RevMinorEdit,
			Size: &schema.Size{
				UnitText: "B",
				Value:    float64(evt.Data.RevLen),
			},
			Editor: &schema.Editor{
				Identifier:  evt.Data.Performer.UserID,
				Name:        evt.Data.Performer.UserText,
				EditCount:   evt.Data.Performer.UserEditCount,
				Groups:      evt.Data.Performer.UserGroups,
				IsBot:       evt.Data.Performer.UserIsBot,
				DateStarted: &evt.Data.Performer.UserRegistrationDt,
			},
		}

		art.IsPartOf = &schema.Project{
			Identifier: evt.Data.Database,
			URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
		}

		art.Namespace = &schema.Namespace{
			Identifier: evt.Data.PageNamespace,
		}

		lng, err := p.Dictionary.GetLanguage(ctx, evt.Data.Database)

		if err != nil {
			return err
		}

		art.InLanguage = &schema.Language{
			Identifier: lng,
		}

		url, err := url.QueryUnescape(evt.Data.Meta.URI)

		if err != nil {
			return err
		}

		art.URL = url

		err = p.Producer.Produce(ctx, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticleCreate,
			Value:  art,
			Key: &schema.Key{
				Identifier: fmt.Sprintf("/%s/%s", art.IsPartOf.Identifier, art.Name),
				Type:       schema.KeyTypeArticle,
			},
		})

		if err != nil {
			return err
		}

		return p.Redis.Set(ctx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
