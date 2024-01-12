package main

import (
	"context"
	"log"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articlevisibility/handler"
	"wikimedia-enterprise/services/structured-data/packages/container"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	app := func(env *env.Environment, sbs subscriber.Subscriber, pms handler.Parameters) error {
		ctx := context.Background()
		hdl := handler.NewArticleVisibility(&pms)
		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          []string{env.TopicArticleVisibilityChange},
			NumberOfWorkers: env.NumberOfWorkers,
		}

		go func() {
			for evt := range scf.Events {
				log.Println(evt.Error)
			}
		}()

		return sbs.
			Subscribe(ctx, hdl, scf)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Panic(err)
	}
}
