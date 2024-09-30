package handler

import (
	"context"
	"fmt"
	"net/url"
	"time"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/tracing"
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
	Tracer     tracing.Tracer
}

// LastEventTimeKey is the redis key where the latest event will be stored.
const LastEventTimeKey = "stream:revision-create:since"

// RevisionCreate handler for the revision create stream.
func RevisionCreate(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.RevisionCreate) error {
	return func(evt *eventstream.RevisionCreate) error {
		if evt.Data.PageIsRedirect || !fr.Projects.IsSupported(evt.Data.Database) || !fr.Namespaces.IsSupported(evt.Data.PageNamespace) {
			return nil
		}
		sev := schema.NewEvent(schema.EventTypeUpdate)

		end, trx := p.Tracer.StartTrace(ctx, "revision-create", map[string]string{"event": "revision-create"})
		var err error
		defer func() {
			if err != nil {
				end(err, fmt.Sprintf("error processing revision-create event with id %s", sev.Identifier))
			} else {
				end(nil, fmt.Sprintf("revision-create event with id %s processed", sev.Identifier))
			}
		}()

		article := new(schema.Article)
		article.Event = sev
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

		langid, err := p.Dictionary.GetLanguage(trx, evt.Data.Database)
		if err != nil {
			log.Error(
				"dictionary get language error",
				log.Any("database", evt.Data.Database),
				log.Any("event_id", article.Event.Identifier),
				log.Any("error", err),
			)

			return err
		}
		article.InLanguage = &schema.Language{
			Identifier: langid,
		}

		url, err := url.QueryUnescape(evt.Data.Meta.URI)

		if err != nil {
			log.Error(
				"error unescaping url",
				log.Any("url", evt.Data.Meta.URI),
				log.Any("event_id", article.Event.Identifier),
				log.Any("error", err),
			)

			return err
		}

		article.URL = url

		hds := tracing.NewHeadersCarrier().InjectContext(trx)

		err = p.Producer.Produce(trx, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticleUpdate,
			Value:  article,
			Key: &schema.Key{
				Identifier: fmt.Sprintf("/%s/%s", article.IsPartOf.Identifier, article.Name),
				Type:       schema.KeyTypeArticle,
			},
			Headers: hds,
		})

		if err != nil {
			log.Error(
				"producer produce error",
				log.Any("error", err),
				log.Any("revision", article.Version.Identifier),
				log.Any("event_id", sev.Identifier),
			)

			return err
		}

		return p.Redis.Set(trx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
