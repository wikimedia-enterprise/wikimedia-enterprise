package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articlevisibility/handler"
	"wikimedia-enterprise/services/structured-data/libraries/health"
	"wikimedia-enterprise/services/structured-data/packages/container"
	"wikimedia-enterprise/services/structured-data/submodules/log"
	pr "wikimedia-enterprise/services/structured-data/submodules/prometheus"
	"wikimedia-enterprise/services/structured-data/submodules/subscriber"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err)
	}

	sgs := make(chan os.Signal, 1)
	signal.Notify(sgs, os.Interrupt, syscall.SIGTERM)

	app := func(env *env.Environment, sbs *subscriber.Subscriber, pms handler.Parameters, consumer *kafka.Consumer, producer *kafka.Producer) error {
		pms.Metrics.AddStructuredDataMetrics()

		ctx := context.Background()
		hdl := handler.NewArticleVisibility(&pms)

		producerTopics := []string{env.TopicArticles}
		cancel, rebalanceCb, err := health.SetUpHealthChecks(producerTopics, ctx, env, producer, consumer)
		if err != nil {
			log.Fatal(err)
		}
		defer cancel()

		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          []string{env.TopicArticleVisibilityChange},
			NumberOfWorkers: env.NumberOfWorkers,
			RebalanceCb:     rebalanceCb,
		}

		go func() {
			for evt := range scf.Events {
				log.Error(evt.Error)
				pms.Metrics.Inc(pr.SDTtlErrs)
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
