package handler

import (
	"context"
	"fmt"
	"net/url"
	"time"

	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
	"go.uber.org/dig"

	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/tracing"
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
	Tracer     tracing.Tracer
}

// PageMove handler for the page move stream. This handler treats a pagemove event
// as if it was a pagedelete event.
func PageMove(ctx context.Context, p *Parameters, fr *filter.Filter) func(evt *eventstream.PageMove) error {
	return func(evt *eventstream.PageMove) error {
		if evt.Data.PageIsRedirect || !fr.Projects.IsSupported(evt.Data.Database) {
			return nil
		}

		sev := schema.NewEvent(schema.EventTypeMove)

		end, trx := p.Tracer.StartTrace(ctx, "page-move", map[string]string{"event.identifier": sev.Identifier})
		var err error
		defer func() {
			if err != nil {
				end(err, fmt.Sprintf("error processing page-move event with id %s", sev.Identifier))
			} else {
				end(nil, fmt.Sprintf("page-move event with id %s processed", sev.Identifier))
			}
		}()

		langid, err := p.Dictionary.GetLanguage(trx, evt.Data.Database)
		if err != nil {
			log.Error(
				"dictionary get language error",
				log.Any("database", evt.Data.Database),
				log.Any("event_id", sev.Identifier),
				log.Any("error", err),
			)

			return err
		}

		dte := time.Now()

		article := &schema.Article{
			Event:        sev,
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
			log.Error(
				"url query unescape error",
				log.Any("error", err),
				log.Any("event_id", sev.Identifier),
				log.Any("url", evt.Data.Meta.URI),
			)

			return err
		}

		article.URL = url

		hds := tracing.NewHeadersCarrier().InjectContext(trx)

		err = p.Producer.Produce(trx, &schema.Message{
			Config: schema.ConfigArticle,
			Topic:  p.Env.TopicArticleMove,
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
				log.Any("event_id", sev.Identifier),
				log.Any("revision", article.Version.Identifier),
			)

			return err
		}

		if fr.Namespaces.IsSupported(evt.Data.PriorState.PageNamespace) {
			article.Event = schema.NewEvent(schema.EventTypeDelete)
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
					log.Any("error", err),
					log.Any("event_id", sev.Identifier),
					log.Any("revision", article.Version.Identifier),
				)

				return err
			}
		}

		return p.Redis.Set(trx, LastEventTimeKey, evt.Data.Meta.Dt, time.Hour*24).Err()
	}
}
