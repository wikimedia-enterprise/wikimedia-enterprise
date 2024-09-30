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
	"go.uber.org/dig"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

// Parameters needed by handler.
type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Dictionary langid.Dictionarer
	Env        *env.Environment
	Tracer     tracing.Tracer
}

// LastEventTimeKey the redis key where the latest event time will be stored
const LastEventTimeKey = "stream:revision-visibility:since"

// RevisionVisibility handler (suppression of revision events) for the revision visibility stream.
func RevisionVisibility(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.RevisionVisibilityChange) error {
	return func(evt *eventstream.RevisionVisibilityChange) error {
		if evt.Data.PageIsRedirect || !fr.Namespaces.IsSupported(evt.Data.PageNamespace) || !fr.Projects.IsSupported(evt.Data.Database) {
			return nil
		}

		sev := schema.NewEvent(schema.EventTypeVisibilityChange)
		end, trx := p.Tracer.StartTrace(ctx, "revision-visibility", map[string]string{"event": "revision-create"})
		var err error
		defer func() {
			if err != nil {
				end(err, fmt.Sprintf("error processing revision-visibilty event with id %s", sev.Identifier))
			} else {
				end(nil, fmt.Sprintf("revision-visibilty event with id %s processed", sev.Identifier))
			}
		}()

		// build data according to our defined schema
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
		article.Visibility = &schema.Visibility{
			Text:    evt.Data.Visibility.Text,
			Editor:  evt.Data.Visibility.User,
			Comment: evt.Data.Visibility.Comment,
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
			Topic:  p.Env.TopicArticleVisibilityChange,
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
			)

			return err
		}

		return p.Redis.Set(trx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
