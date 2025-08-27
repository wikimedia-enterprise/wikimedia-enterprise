// healthcheckutils/healthcheckutils.go
package healthcheckutils

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	"wikimedia-enterprise/services/on-demand/config/env"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	healthchecks "github.com/wikimedia-enterprise/health-checker/health"
)

// LogHealthCheck is a default logging function for health check results.
func LogHealthCheck(name string, _ string, result error) {
	if result != nil {
		log.Printf("health check %s returned an error: %v", name, result)
	} else {
		log.Printf("health check %s succeeded", name)
	}
}

// KafkaHealthCheckConfig holds configuration for Kafka health checks.
type KafkaHealthCheckConfig struct {
	Interval       time.Duration
	Consumer       *kafka.Consumer
	ConsumerTopics env.List
	MaxLag         int64
}

// S3HealthCheckConfig holds configuration for S3 health checks
type S3HealthCheckConfig struct {
	BucketName string
	Name       string
	Region     string
	S3Client   s3iface.S3API
}

// SetupHealthChecks sets up Kafka and S3 health checks.
func SetUpHealthChecks(
	ctx context.Context,
	serviceName string,
	version string,
	env *env.Environment,
	consumer *kafka.Consumer,
	s3Client s3iface.S3API,
	logFunc func(string, string, error),
) (cancel func(), rebalanceCb kafka.RebalanceCb, err error) {

	var checkers []healthchecks.HealthChecker

	syncKafka, err := healthchecks.NewSyncKafkaChecker(healthchecks.SyncKafkaChecker{
		Name:           "kafka-health-check",
		Interval:       time.Duration(env.KafkaHealthCheckIntervalMs) * time.Millisecond,
		Producer:       nil,
		ProducerTopics: []string{},
		Consumer:       consumer,
		ConsumerTopics: env.HealthChecksConsumerTopics,
		MaxLag:         int64(env.KafkaHealthCheckLagNMS),
	}, healthchecks.NewConsumerOffsetStore())

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kafka health checker: %w", err)
	}
	asyncKafka := healthchecks.NewAsyncKafkaChecker(syncKafka)
	checkers = append(checkers, asyncKafka)

	s3Checker, err := healthchecks.NewS3Checker(healthchecks.S3CheckerConfig{
		BucketName: env.AWSBucket,
		Name:       "on-demand-s3-articles",
		Region:     env.AWSRegion,
		S3Client:   s3Client,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create S3 health checker: %w", err)
	}

	checkers = append(checkers, s3Checker)
	logFn := LogHealthCheck

	h, err := healthchecks.SetupHealthChecks(serviceName, version, true, logFn, env.HealthCheckMaxRetries, checkers...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup health checks: %w", err)
	}

	kafkaCtx, cancel := context.WithCancel(ctx)
	asyncKafka.Start(kafkaCtx)

	healthchecks.StartHealthCheckServer(h, ":"+strconv.Itoa(env.HealthCheckPort))

	cb := func(c *kafka.Consumer, event kafka.Event) error {
		log.Printf("Kafka rebalance event: %s", event.String())
		return syncKafka.ProcessRebalance(c, event)
	}

	return cancel, cb, nil
}
