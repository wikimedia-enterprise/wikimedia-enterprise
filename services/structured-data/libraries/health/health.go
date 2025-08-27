package health

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/submodules/log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/wikimedia-enterprise/health-checker/health"
)

func logHealthCheck(name string, _ string, result error) {
	if result != nil {
		log.Error(fmt.Sprintf("health check %s returned an error: %v", name, result))
	} else {
		log.Debug(fmt.Sprintf("health check %s succeeded", name))
	}
}

// SetUpHealthChecks configures and starts the healthchecks server, exposing the /healthz endpoint.
func SetUpHealthChecks(requiredProducerTopics []string, ctx context.Context, env *env.Environment, producer *kafka.Producer, consumer *kafka.Consumer) (cancel func(), rebalanceCb func(*kafka.Consumer, kafka.Event) error, err error) {
	ignoreOffsets := false
	maxLag := env.HealthChecksMaxConsumerLag
	if env.HealthChecksMaxConsumerLag <= 0 {
		log.Debug("max consumer lag for kafka health checks <= 0, will track only consumed topics presence")
		ignoreOffsets = true
		maxLag = 0
	}
	sync, err := health.NewSyncKafkaChecker(health.SyncKafkaChecker{
		Name:                  "kafka-health-check",
		Interval:              time.Duration(env.HealthChecksAsyncIntervalMs) * time.Millisecond,
		Producer:              producer,
		Consumer:              consumer,
		ProducerTopics:        requiredProducerTopics,
		ConsumerTopics:        env.HealthChecksConsumerTopics,
		MaxLag:                int64(maxLag),
		ConsumerIgnoreOffsets: ignoreOffsets,
	}, health.NewConsumerOffsetStore())
	if err != nil {
		return nil, nil, fmt.Errorf("error creating sync Kafka checker: %w", err)
	}

	async := health.NewAsyncKafkaChecker(sync)
	h, err := health.SetupHealthChecks(env.HealthChecksComponentName, "1.0.0", true, logHealthCheck, 3, async)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set up health checks: %w", err)
	}

	kafkaCtx, cancel := context.WithCancel(ctx)
	async.Start(kafkaCtx)

	health.StartHealthCheckServer(h, ":"+strconv.Itoa(env.HealthChecksPort))

	return cancel, sync.ProcessRebalance, nil
}
