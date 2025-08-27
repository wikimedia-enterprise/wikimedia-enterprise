package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/handlers/structured/handler"
	"wikimedia-enterprise/services/on-demand/packages/container"
	pr "wikimedia-enterprise/services/on-demand/submodules/prometheus"
	"wikimedia-enterprise/services/on-demand/submodules/subscriber"

	hutils "wikimedia-enterprise/services/on-demand/libraries/healthcheckutils"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

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
			"on-demand-structured",
			"1.0.0",
			env,
			consumer,
			S3Interface,
			nil,
		)

		if err != nil {
			log.Fatalf("failed to setup healthchecks: %v", err)
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
