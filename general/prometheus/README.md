# Wikimedia Enterprise Prometheus instrumentation kit.

## About
This package allows to instrument any service and expose the `/metrics` endpoint for the Prometheus service and aims to monitor simplified version of Four golden signals:
- Traffic or Request Rate
- Errors
- Latency or Duration of the requests
- Saturation (hard to measure)

Simplified version is ussually refered as RED-method and contains following metrics:
- Request rate
- Errors
- Duration (latency) distribution

## Usage
```go
prommet := new(pr.Metrics)
prommet.AddEventStreamMetrics()
prommet.AddRedisMetrics()

cnr := rate.NewAvgRateCounter(1 * time.Second) // Averaga rate
...
cnr.Incr(1)
prommet.Set(pr.EsTtlEvntsPs, float64(cnr.Rate()))

...
prommet.Inc(pr.EsTtlEvents)
...

go func() {
    if err := pr.Run(pr.Parameters{
        Port:  p.Env.PrometheusPort,
        Redis: p.Redis.(*redis.Client),
        Metrics: prommet,
    }); err != nil {
        log.Println(err)
    }
}()
```

###Track counter metrics:
```go

```

###Track duration metrics (summaries and histograms):
```go
timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
    prommet.Observe(pr.RedisReqDur, v)
}))

defer timer.ObserveDuration()
```