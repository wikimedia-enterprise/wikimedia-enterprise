package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articledelete/handler"
	"wikimedia-enterprise/services/structured-data/libraries/health"
	"wikimedia-enterprise/services/structured-data/packages/container"
	"wikimedia-enterprise/services/structured-data/submodules/log"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/subscriber"

	pr "wikimedia-enterprise/services/structured-data/submodules/prometheus"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err)
	}

	sgs := make(chan os.Signal, 1)
	signal.Notify(sgs, os.Interrupt, syscall.SIGTERM)

	app := func(env *env.Environment, rtr schema.Retryer, sbs *subscriber.Subscriber, pms handler.Parameters, consumer *kafka.Consumer, producer *kafka.Producer) error {
		ctx := context.Background()
		hdl := handler.NewArticleDelete(&pms)

		producerTopics := []string{env.TopicArticles, env.TopicArticleDeleteError, env.TopicArticleDeleteDeadLetter}
		cancel, rebalanceCb, err := health.SetUpHealthChecks(producerTopics, ctx, env, producer, consumer)
		if err != nil {
			log.Fatal(err)
		}
		defer cancel()

		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          []string{env.TopicArticleDelete},
			NumberOfWorkers: env.NumberOfWorkers,
			RebalanceCb:     rebalanceCb,
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
						TopicError:      env.TopicArticleDeleteError,
						TopicDeadLetter: env.TopicArticleDeleteDeadLetter,
						MaxFailCount:    env.MaxFailCount,
						Error:           evt.Error,
						Message:         evt.Message,
					}

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
