package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"wikimedia-enterprise/services/structured-contents/config/env"
	"wikimedia-enterprise/services/structured-contents/handler"
	"wikimedia-enterprise/services/structured-contents/packages/container"
	"wikimedia-enterprise/services/structured-contents/submodules/log"
	pr "wikimedia-enterprise/services/structured-contents/submodules/prometheus"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	"wikimedia-enterprise/services/structured-contents/submodules/subscriber"
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
		hdl := handler.New(&pms)
		scf := &subscriber.Config{
			Events:          make(chan *subscriber.Event, env.EventChannelSize),
			Topics:          env.InputTopics,
			NumberOfWorkers: env.NumberOfWorkers,
		}

		pms.Metrics.AddStructuredContentsMetrics()

		// Process error events from the subscriber
		go func() {
			for evt := range scf.Events {
				if evt.Error != nil {
					log.Info("error processing upstream message", log.Any("error", evt.Error))
				}

				pms.Metrics.Inc(pr.SCTtlErrs)

				// Sending failed msgs directly to dead letter by setting MAX_FAIL_COUNT to 0
				if len(env.DeadLetterTopic) > 0 && evt.Message != nil {
					rms := &schema.RetryMessage{
						Config:          schema.ConfigArticle,
						TopicError:      env.DeadLetterTopic,
						TopicDeadLetter: env.DeadLetterTopic,
						MaxFailCount:    env.MaxFailCount,
						Error:           evt.Error,
						Message:         evt.Message,
					}

					if err := rtr.Retry(ctx, rms); err != nil {
						log.Error("error sending failed message to retry topics", log.Any("error", err))
					}
				}
			}
		}()

		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    env.PrometheusPort,
				Metrics: pms.Metrics,
			}); err != nil {
				log.Error("error starting prometheus", log.Any("error", err))
			}
		}()

		log.Info("subscribing to kafka", log.Any("topics", env.InputTopics), log.Any("workers", env.NumberOfWorkers))

		return sbs.
			Subscribe(ctx, hdl, scf)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Fatal("error starting app", log.Any("error", err))
	}
}
