package main

import (
	"context"
	"log"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/handlers/articles/handler"
	"wikimedia-enterprise/services/on-demand/packages/container"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	app := func(env *env.Environment, sbs subscriber.Subscriber, prs handler.Parameters) error {
		ctx := context.Background()
		hdl := handler.New(&prs)
		scf := &subscriber.Config{
			NumberOfWorkers: env.NumberOfWorkers,
			Topics:          env.TopicArticles,
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
		}

		go func() {
			for evt := range scf.Events {
				if evt.Error != nil {
					log.Println(evt.Error)
				}
			}
		}()

		return sbs.Subscribe(
			ctx,
			hdl,
			scf,
		)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Panic(err)
	}
}
