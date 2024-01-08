package main

import (
	"context"
	"log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articlebulk/handler"
	"wikimedia-enterprise/services/structured-data/packages/container"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	app := func(env *env.Environment, rtr schema.Retryer, sbs subscriber.Subscriber, pms handler.Parameters) error {
		ctx := context.Background()
		hdl := handler.NewArticleBulk(&pms)
		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          []string{env.TopicArticleBulk},
			NumberOfWorkers: env.NumberOfWorkers,
		}

		go func() {
			for evt := range scf.Events {
				if evt.Message == nil {
					log.Println(evt.Error)
				}

				if evt.Message != nil {
					rms := &schema.RetryMessage{
						Config:          schema.ConfigArticleNames,
						TopicError:      env.TopicArticleBulkError,
						TopicDeadLetter: env.TopicArticleBulkDeadLetter,
						MaxFailCount:    env.MaxFailCount,
						Error:           evt.Error,
						Message:         evt.Message,
					}

					if err := rtr.Retry(ctx, rms); err != nil {
						log.Println(err)
					}
				}
			}
		}()

		return sbs.
			Subscribe(ctx, hdl, scf)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Panic(err)
	}
}
