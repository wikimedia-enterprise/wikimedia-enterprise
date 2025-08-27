package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/handlers/articles/handler"
	"wikimedia-enterprise/services/on-demand/packages/container"
	"wikimedia-enterprise/services/on-demand/submodules/log"
	pr "wikimedia-enterprise/services/on-demand/submodules/prometheus"
	"wikimedia-enterprise/services/on-demand/submodules/subscriber"

	hutils "wikimedia-enterprise/services/on-demand/libraries/healthcheckutils"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	sgs := make(chan os.Signal, 1)
	signal.Notify(sgs, os.Interrupt, syscall.SIGTERM)

	app := func(env *env.Environment, sbs *subscriber.Subscriber, prs handler.Parameters, consumer *kafka.Consumer, S3Interface s3iface.S3API) error {
		ctx := context.Background()
		hdl := handler.New(&prs)

		cancel, rebalanceCb, err := hutils.SetUpHealthChecks(
			ctx,
			"on-demand-articles",
			"1.0.0",
			env,
			consumer,
			S3Interface,
			nil,
		)

		if err != nil {
			log.Fatal("failed to setup healthchecks", log.Any("error", err))
		}

		defer cancel()

		scf := &subscriber.Config{
			NumberOfWorkers:    env.NumberOfWorkers,
			Topics:             env.TopicArticles,
			Events:             make(chan *subscriber.Event, env.EventChannelSize),
			RebalanceCb:        rebalanceCb,
			MessagesChannelCap: env.WorkerBufferSize,
		}

		prs.Metrics.AddOnDemandMetrics()
		err = subscriber.RegisterMetrics(prometheus.DefaultRegisterer)
		if err != nil {
			return err
		}

		go func() {
			for evt := range scf.Events {
				prs.Metrics.Inc(pr.OdmTtlErrs)
				if evt.Error != nil {
					var tpr *kafka.TopicPartition
					if evt.Message != nil {
						tpr = &evt.Message.TopicPartition
					}
					log.Error("error from consumer",
						log.Any("error", evt.Error),
						log.Any("message", tpr),
					)
				}
			}
		}()

		go func() {
			<-sgs
			err := prs.Tracer.Shutdown(ctx)

			if err != nil {
				log.Warn(err)
			}
		}()

		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    env.PrometheusPort,
				Metrics: prs.Metrics,
			}); err != nil {
				log.Error(err)
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
