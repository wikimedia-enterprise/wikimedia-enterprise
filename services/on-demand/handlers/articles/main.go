package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	pr "wikimedia-enterprise/general/prometheus"
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

	sgs := make(chan os.Signal, 1)
	signal.Notify(sgs, os.Interrupt, syscall.SIGTERM)

	app := func(env *env.Environment, sbs *subscriber.Subscriber, prs handler.Parameters) error {
		ctx := context.Background()
		hdl := handler.New(&prs)
		scf := &subscriber.Config{
			NumberOfWorkers: env.NumberOfWorkers,
			Topics:          env.TopicArticles,
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
		}

		prs.Metrics.AddOnDemandMetrics()

		go func() {
			for evt := range scf.Events {
				prs.Metrics.Inc(pr.OdmTtlErrs)
				if evt.Error != nil {
					log.Println(evt.Error)
				}
			}
		}()

		go func() {
			<-sgs
			err := prs.Tracer.Shutdown(ctx)

			if err != nil {
				log.Println(err)
			}
		}()

		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    env.PrometheusPort,
				Metrics: prs.Metrics,
			}); err != nil {
				log.Println(err)
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
