package main

import (
	"context"
	"log"
	"time"
	pr "wikimedia-enterprise/general/prometheus"
	"wikimedia-enterprise/services/event-bridge/handlers/pagedelete/handler"
	"wikimedia-enterprise/services/event-bridge/packages/container"
	"wikimedia-enterprise/services/event-bridge/packages/filter"
	"wikimedia-enterprise/services/event-bridge/packages/shutdown"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/redis/go-redis/v9"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

func main() {
	ctx := context.Background()
	sh := shutdown.NewHelper(ctx)
	cont, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	fr, err := filter.New()

	if err != nil {
		log.Panic(err)
	}

	err = cont.Invoke(func(p handler.Parameters, prod *kafka.Producer) {
		since := time.Now()

		// Prometherus metrics.
		prommet := new(pr.Metrics)
		prommet.Init()
		prommet.AddEventStreamMetrics()
		prommet.AddRedisMetrics()

		if data, err := p.Redis.Get(ctx, handler.LastEventTimeKey).Time(); err == nil {
			prommet.Inc(pr.RedisReqTtl, "LastEventTimeKey_miss")
			since = data
		}

		wg := sh.WG()
		hl := handler.PageDelete(ctx, &p, fr)
		stream := eventstream.
			NewClient().
			PageDelete(sh.Ctx(), since, func(evt *eventstream.PageDelete) error {
				prommet.Inc(pr.EsTtlEvents)
				wg.Add(1)
				defer wg.Done()
				return hl(evt)
			})

		// Start event stream on a separate goroutine.
		go func() {
			for err := range stream.Sub() {
				prommet.Inc(pr.EsTtlErrs)
				log.Println(err)
			}
		}()

		// Prometheus server run.
		go func() {
			if err := pr.Run(pr.Parameters{
				Port:    p.Env.PrometheusPort,
				Redis:   p.Redis.(*redis.Client),
				Metrics: prommet,
			}); err != nil {
				log.Println(err)
			}
		}()

		go sh.Shutdown()

		sh.Wait(prod)
	})

	if err != nil {
		log.Panic(err)
	}
}
