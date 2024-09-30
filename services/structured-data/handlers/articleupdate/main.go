package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articleupdate/handler"
	"wikimedia-enterprise/services/structured-data/packages/container"

	pr "wikimedia-enterprise/general/prometheus"
)

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err)
	}

	sgs := make(chan os.Signal, 1)
	signal.Notify(sgs, os.Interrupt, syscall.SIGTERM)

	app := func(env *env.Environment, rtr schema.Retryer, sbs *subscriber.Subscriber, pms handler.Parameters) error {
		ctx := context.Background()
		hdl := handler.NewArticleUpdate(&pms)
		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          []string{env.TopicArticleUpdate},
			NumberOfWorkers: env.NumberOfWorkers,
		}

		pms.Metrics.AddStructuredDataMetrics()

		go func() {
			for evt := range scf.Events {
				pms.Metrics.Inc(pr.SDTtlErrs)

				if evt.Message == nil {
					log.Error(evt.Error)
				}

				if evt.Message != nil {
					rms := &schema.RetryMessage{
						Config:          schema.ConfigArticle,
						TopicError:      env.TopicArticleUpdateError,
						TopicDeadLetter: env.TopicArticleUpdateDeadLetter,
						MaxFailCount:    env.MaxFailCount,
						Error:           evt.Error,
						Message:         evt.Message,
					}

					log.Info(
						"trying to send retry message",
						log.Any("err", evt.Error),
					)

					if err := rtr.Retry(ctx, rms); err != nil {
						log.Error(err)
					}
				}
			}
		}()

		go func() {
			<-sgs
			err := pms.Tracer.Shutdown(ctx)

			if err != nil {
				log.Error(err)
			}
		}()

		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    pms.Env.PrometheusPort,
				Metrics: pms.Metrics,
			}); err != nil {
				log.Error(err)
			}
		}()

		return sbs.
			Subscribe(ctx, hdl, scf)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Fatal(err)
	}
}
