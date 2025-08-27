package main

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/handler"
	"wikimedia-enterprise/services/eventstream-listener/packages/container"
	"wikimedia-enterprise/services/eventstream-listener/packages/shutdown"
	"wikimedia-enterprise/services/eventstream-listener/submodules/log"
	pr "wikimedia-enterprise/services/eventstream-listener/submodules/prometheus"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/redis/go-redis/v9"
	"github.com/wikimedia-enterprise/health-checker/health"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

func main() {
	ctx := context.Background()
	sh := shutdown.NewHelper(ctx)
	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	log.Info("dependency intialized successfully")

	err = cnt.Invoke(func(p handler.Parameters, prod *kafka.Producer) {
		cancel, _, err := setUpHealthChecks(ctx, p.Env, p.KafkaProducer, p.Redis)
		if err != nil {
			log.Panic(err)
		}
		defer cancel()

		since := time.Now()
		prommet := new(pr.Metrics)
		prommet.Init()
		prommet.AddEventStreamMetrics()
		prommet.AddRedisMetrics()

		if data, err := p.Redis.Get(ctx, handler.LastEventTimeKey).Time(); err == nil {
			since = data
		}

		wg := sh.WG()
		hl := handler.PageChange(ctx, &p)
		stream := eventstream.
			NewClient().
			PageChange(sh.Ctx(), since, func(evt *eventstream.PageChange) error {
				prommet.Inc(pr.EsTtlEvents, "")
				wg.Add(1)
				defer wg.Done()
				return hl(evt)
			})

		log.Info("stream created successfully")

		go func() {
			for err := range stream.Sub() {
				prommet.Inc(pr.EsTtlErrs, "")
				log.Error(err)
			}
		}()

		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    p.Env.PrometheusPort,
				Redis:   p.Redis.(*redis.Client),
				Metrics: prommet,
			}); err != nil {
				log.Error(err)
			}
		}()

		go sh.Shutdown()

		sh.Wait(prod)
	})

	if err != nil {
		log.Panic(err)
	}
}

func logHealthCheck(name string, _ string, result error) {
	if result != nil {
		log.Error(fmt.Sprintf("health check %s returned an error: %v", name, result))
	} else {
		log.Debug(fmt.Sprintf("health check %s succeeded", name))
	}
}

// SetUpHealthChecks configures and starts the healthchecks server, exposing the /healthz endpoint.
func setUpHealthChecks(ctx context.Context, env *env.Environment, producer *kafka.Producer, redisClient redis.Cmdable) (cancel func(), rebalanceCb func(*kafka.Consumer, kafka.Event) error, err error) {
	producerTopics := []string{}

	// Use just a few sample topics for the health check.
	producerTopics = append(producerTopics, env.OutputTopics.GetNamesByEventType("commonswiki", schema.EventTypeCreate)...)
	producerTopics = append(producerTopics, env.OutputTopics.GetNamesByEventType("other", schema.EventTypeCreate)...)
	producerTopics = append(producerTopics, env.OutputTopics.GetNamesByEventType("other", schema.EventTypeUpdate)...)

	sync, err := health.NewSyncKafkaChecker(health.SyncKafkaChecker{
		Name:           "kafka-health-check",
		Interval:       time.Duration(env.HealthChecksAsyncIntervalMs) * time.Millisecond,
		Producer:       producer,
		ProducerTopics: producerTopics,
	}, health.NewConsumerOffsetStore())
	if err != nil {
		return nil, nil, fmt.Errorf("error creating sync Kafka checker: %w", err)
	}

	cfg := health.RedisCheckerConfig{
		Name: "redis-checker",
	}
	redisChecker, err := health.NewRedisChecker(redisClient, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Redis checker: %w", err)
	}

	async := health.NewAsyncKafkaChecker(sync)
	h, err := health.SetupHealthChecks("eventstream-listener", "1.0.0", true, logHealthCheck, 3, async, redisChecker)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set up health checks: %w", err)
	}

	kafkaCtx, cancel := context.WithCancel(ctx)
	async.Start(kafkaCtx)

	health.StartHealthCheckServer(h, ":"+strconv.Itoa(env.HealthChecksPort))

	return cancel, sync.ProcessRebalance, nil
}
