package main

import (
	"context"
	"log"
	"time"
	pr "wikimedia-enterprise/general/prometheus"
	"wikimedia-enterprise/services/event-bridge/handlers/pagechange/handler"
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
	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	fr, err := filter.New()

	if err != nil {
		log.Panic(err)
	}

	err = cnt.Invoke(func(p handler.Parameters, prod *kafka.Producer) {
		since := time.Now()
		prommet := new(pr.Metrics)
		prommet.Init()
		prommet.AddEventStreamMetrics()
		prommet.AddRedisMetrics()

		if data, err := p.Redis.Get(ctx, handler.LastEventTimeKey).Time(); err == nil {
			since = data
		}

		wg := sh.WG()
		hl := handler.PageChange(ctx, &p, fr)
		stream := eventstream.
			NewClient().
			PageChange(sh.Ctx(), since, func(evt *eventstream.PageChange) error {
				prommet.Inc(pr.EsTtlEvents, "")
				wg.Add(1)
				defer wg.Done()
				return hl(evt)
			})

		go func() {
			for err := range stream.Sub() {
				prommet.Inc(pr.EsTtlErrs, "")
				log.Println(err)
			}
		}()

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
