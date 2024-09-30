package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articledelete/handler"
	"wikimedia-enterprise/services/structured-data/packages/container"

	pr "wikimedia-enterprise/general/prometheus"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	sgs := make(chan os.Signal, 1)
	signal.Notify(sgs, os.Interrupt, syscall.SIGTERM)

	app := func(env *env.Environment, rtr schema.Retryer, sbs *subscriber.Subscriber, pms handler.Parameters) error {
		ctx := context.Background()
		hdl := handler.NewArticleDelete(&pms)
		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          []string{env.TopicArticleDelete},
			NumberOfWorkers: env.NumberOfWorkers,
		}

		pms.Metrics.AddStructuredDataMetrics()

		go func() {
			for evt := range scf.Events {
				pms.Metrics.Inc(pr.SDTtlErrs)

				if evt.Message == nil {
					log.Println(evt.Error)
				}

				if evt.Message != nil {
					rms := &schema.RetryMessage{
						Config:          schema.ConfigArticle,
						TopicError:      env.TopicArticleDeleteError,
						TopicDeadLetter: env.TopicArticleDeleteDeadLetter,
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

		go func() {
			<-sgs
			err := pms.Tracer.Shutdown(ctx)

			if err != nil {
				log.Println(err)
			}
		}()

		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    pms.Env.PrometheusPort,
				Metrics: pms.Metrics,
			}); err != nil {
				log.Println(err)
			}
		}()

		return sbs.
			Subscribe(ctx, hdl, scf)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Panic(err)
	}
}
