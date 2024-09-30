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

// LastEventTimeKey last event time location in redis.
const LastEventTimeKey = "stream:page-delete:since"

// Parameters required for handler.
type Parameters struct {
	dig.In
	Redis      redis.Cmdable
	Producer   schema.Producer
	Dictionary langid.Dictionarer
	Env        *env.Environment
	Tracer     tracing.Tracer
}

// Pagedelete parses a pagedelete event, creates a kafka message based on article schema, writes to TopicEventBridgeArticleDelete topic, and stores the processed event timestamp.
func PageDelete(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.PageDelete) error {
	return func(evt *eventstream.PageDelete) error {
		if evt.Data.PageIsRedirect || !fr.Namespaces.IsSupported(evt.Data.PageNamespace) || !fr.Projects.IsSupported(evt.Data.Database) {
			return nil
		}

		article := new(schema.Article)
		sev := schema.NewEvent(schema.EventTypeDelete)
		article.Event = sev

		end, trx := p.Tracer.StartTrace(ctx, "page-delete", map[string]string{"event.id": sev.Identifier})
		var err error
		defer func() {
			if err != nil {
				end(err, fmt.Sprintf("error processing page-delete event with id %s", sev.Identifier))
			} else {
				end(nil, fmt.Sprintf("page-delete event with id %s processed", sev.Identifier))
			}
		}()

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
			log.Error(
				"error unescaping url",
				log.Any("url", evt.Data.Meta.URI),
				log.Any("event_id", article.Event.Identifier),
				log.Any("error", err),
			)

			return err
		}

		article.URL = url

		langid, err := p.Dictionary.GetLanguage(ctx, evt.Data.Database)
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

		hds := tracing.NewHeadersCarrier().InjectContext(trx)

		err = p.Producer.Produce(trx, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticleDelete,
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
				log.Any("revision", article.Version.Identifier),
				log.Any("error", err),
			)

			return err
		}

		return p.Redis.Set(trx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
