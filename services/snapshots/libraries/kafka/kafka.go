// Package kafka wraps kafka communications and consumer pooling into interfaces.
package kafka

import (
	"context"
	"errors"
	"strings"
	"time"
	"wikimedia-enterprise/services/snapshots/config/env"

	"github.com/avast/retry-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/wikimedia-enterprise/log"
)

// IsTransientKafkaErr returns true for errors
// that are expected during broker restarts / network blips.
// We retry and allow librdkafka to recover automatically; we don't enforce a local retry cap.
// The guardrail is the gRPC ctx deadline set by the Airflow DAG clients (batches/snapshots),
// which will cancel the request and cause this loop to exit via ctx.Done().
func IsTransientKafkaErr(err error) bool {
	if err == nil {
		return false
	}
	var ke kafka.Error
	if errors.As(err, &ke) {
		switch ke.Code() {
		case kafka.ErrTransport,
			kafka.ErrAllBrokersDown,
			kafka.ErrNotLeaderForPartition,
			kafka.ErrLeaderNotAvailable,
			kafka.ErrNetworkException,
			kafka.ErrRequestTimedOut,
			kafka.ErrTimedOut,
			kafka.ErrNotCoordinator,
			kafka.ErrCoordinatorLoadInProgress,
			kafka.ErrNotController:
			return true
		default:
			return false
		}
	}

	// Fallback for non-typed / wrapped errors (e.g. "connection refused" bubbles up)
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "all brokers down") ||
		strings.Contains(s, "local: timed out") ||
		strings.Contains(s, "timed out") ||
		strings.Contains(s, "connection setup timed out") ||
		strings.Contains(s, "setup timed out in state connect") ||
		strings.Contains(s, "broker transport failure") ||
		strings.Contains(s, "failed: connect") ||
		strings.Contains(s, "connect to ipv4") ||
		strings.Contains(s, "connect to ipv6") ||
		strings.Contains(s, "not leader for partition") ||
		strings.Contains(s, "leader not available")
}

// kafkaRetryOpts builds the shared retry-go configuration for Kafka calls.
// It retries transient Kafka errors, respects context cancellation, and applies
// a bounded backoff strategy.
func kafkaRetryOpts(ctx context.Context) []retry.Option {
	return []retry.Option{
		retry.Context(ctx),
		retry.RetryIf(IsTransientKafkaErr),
		retry.Delay(500 * time.Millisecond),
		retry.MaxDelay(5 * time.Second),
		retry.LastErrorOnly(true),
	}
}

// ErrReadAllEndMessage this message can be passed to ReadAll callback function to stop the consumption.
var ErrReadAllEndMessage = errors.New("read all end message")

// KafkaConsumer wraps kafka consumer into interface.
type KafkaConsumer interface {
	Assign([]kafka.TopicPartition) error
	ReadMessage(time.Duration) (*kafka.Message, error)
	GetMetadata(*string, bool, int) (*kafka.Metadata, error)
	QueryWatermarkOffsets(topic string, partition int32, timeoutMs int) (low, high int64, err error)
	GetWatermarkOffsets(string, int32) (int64, int64, error)
	OffsetsForTimes([]kafka.TopicPartition, int) ([]kafka.TopicPartition, error)
	Close() error
}

// ReadAllCloser is an interface that wraps kafka consumer to retrieve all messages from the topic.
type ReadAllCloser interface {
	ReadAll(ctx context.Context, since int, topic string, partition int, identifier string, callback func(msg *kafka.Message) error) error
	GetWatermarkOffsets(string, int) (int, int, error)
	GetMetadata(string) (*kafka.TopicMetadata, error)
	Close() error
}

// Consumer kafka library wrapper to read all messages by partition from the topic.
type Consumer struct {
	Consumer     KafkaConsumer
	LogWatermark bool
	UseWatermark bool
}

// GetMeta gets topic metadata from broker.
func (c *Consumer) GetMetadata(tpc string) (*kafka.TopicMetadata, error) {
	mtd, err := c.Consumer.GetMetadata(&tpc, false, 1000)

	if err != nil {
		return nil, err
	}

	tpm := mtd.Topics[tpc]
	return &tpm, nil
}

// GetWatermarkOffsets returns high and a low offset for the topic.
func (c *Consumer) GetWatermarkOffsets(tpc string, ptn int) (int, int, error) {
	lof, hof, err := c.Consumer.GetWatermarkOffsets(tpc, int32(ptn)) // #nosec G115
	return int(lof), int(hof), err
}

// ReadAll retrieves all messages from provided partitions and topics.
func (c *Consumer) ReadAll(ctx context.Context, since int, topic string, partition int, identifier string, callback func(msg *kafka.Message) error) error {
	idrField := log.Any("idr", identifier)
	partitionField := log.Any("partition", partition)

	tp := kafka.TopicPartition{
		Topic:     &topic,
		Partition: int32(partition), // #nosec G115
		Offset:    kafka.OffsetBeginning,
	}

	if since > 0 {
		if err := tp.Offset.Set(since); err != nil {
			return err
		}
	}

	if since > 0 {
		var otp []kafka.TopicPartition

		if err := retry.Do(
			func() error {
				var err error
				otp, err = c.Consumer.OffsetsForTimes([]kafka.TopicPartition{tp}, 60000)
				return err
			},
			kafkaRetryOpts(ctx)...,
		); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		tp = otp[0]

		if tp.Offset == kafka.OffsetEnd {
			log.Info("skipping partition with no messages since specified time", idrField, partitionField)
			return nil
		}
	}

	var highWatermark int64
	if c.LogWatermark || c.UseWatermark {
		var low int64

		if err := retry.Do(
			func() error {
				var err error
				low, highWatermark, err = c.Consumer.QueryWatermarkOffsets(topic, int32(partition), 1000)
				return err
			},
			kafkaRetryOpts(ctx)...,
		); err != nil {
			log.Warn("error retrieving watermark", idrField, partitionField)

			if c.UseWatermark {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
		}

		log.Info("obtained watermark offsets", idrField, partitionField, log.Any("low", kafka.Offset(low)), log.Any("high", kafka.Offset(highWatermark)))

		if highWatermark == 0 {
			// Empty partition.
			log.Info("empty partition based on watermark", idrField, partitionField)
			return nil
		}
	}

	log.Info("assigning partition offset", idrField, partitionField, log.Any("offset", tp.Offset))
	if err := retry.Do(
		func() error {
			return c.Consumer.Assign([]kafka.TopicPartition{tp})
		},
		kafkaRetryOpts(ctx)...,
	); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}

	foundLast := false
	timeoutRetries := 5
	for {
		if foundLast {
			log.Info("read all, finished with last message", idrField, partitionField)
			return nil
		}
		msg, err := c.Consumer.ReadMessage(time.Second * 1)

		select {
		case <-ctx.Done():
			return nil
		default:
			if err != nil {

				if ke, ok := err.(kafka.Error); ok && ke.Code() == kafka.ErrTimedOut {
					log.Warn("received timeout err", log.Any("error", err), idrField, partitionField)

					if c.UseWatermark {
						if timeoutRetries <= 0 {
							log.Error("exhausted retries, finished with timeout err", log.Any("error", err), idrField, partitionField)
							return nil
						}

						timeoutRetries -= 1
						time.Sleep(time.Second)
						continue
					}

					log.Error("read all, finished with timeout err", log.Any("error", err), idrField, partitionField)
					return nil
				}
				if IsTransientKafkaErr(err) {
					log.Info(
						"read all, transient kafka error (will retry)",
						log.Any("error", err),
						idrField, partitionField,
					)
					time.Sleep(500 * time.Millisecond)
					continue
				}

				// Everything else bubbles
				log.Error("read all, default error finished with err", log.Any("error", err), idrField, partitionField)
				return err
			}

			if c.UseWatermark {
				offset := int64(msg.TopicPartition.Offset)
				if offset > highWatermark-1 {
					// Rare but possible: high-1 was the last message when we started consuming, but compaction
					// deleted it because we got a duplicate later. Technically we might be missing articles here,
					// but this is very unlikely.
					log.Info("read all, finished past watermark", log.Any("offset", offset), idrField, partitionField)
					return nil
				}

				if offset == highWatermark-1 {
					// We'll pass it to callback, but then we're done.
					foundLast = true
					log.Info("read all, got last message", log.Any("offset", offset), idrField, partitionField)
				}
			}

			if err := callback(msg); err != nil {
				if err == ErrReadAllEndMessage {
					log.Info("read all, finishing with end err", log.Any("error", err), idrField, partitionField)
					return nil
				}

				log.Error("read all, finishing with unknown err", log.Any("error", err), idrField, partitionField)
				return err
			}
		}
	}
}

// Close closes consumer connection.
func (c *Consumer) Close() error {
	return c.Consumer.Close()
}

// ConsumerGetter wrapper for the consumer pool struct to get a consumer by group id.
type ConsumerGetter interface {
	GetConsumer(string) (ReadAllCloser, error)
}

// Pool provides the ability to generate new consumer by group id.
type Pool struct {
	Config       *kafka.ConfigMap
	LogWatermark bool
	UseWatermark bool
}

// GetConsumer creates new kafka consumer within provided group id.
func (p *Pool) GetConsumer(gid string) (ReadAllCloser, error) {
	cfg := kafka.ConfigMap{}

	for key, val := range *p.Config {
		cfg[key] = val
	}

	cfg["group.id"] = gid

	kcs, err := kafka.NewConsumer(&cfg)

	if err != nil {
		return nil, err
	}

	return &Consumer{
		Consumer:     kcs,
		LogWatermark: p.LogWatermark,
		UseWatermark: p.UseWatermark,
	}, nil
}

// NewPool creates new kafka consumer with default configuration.
func NewPool(env *env.Environment) ConsumerGetter {
	cfg := kafka.ConfigMap{
		"bootstrap.servers":    env.KafkaBootstrapServers,
		"message.max.bytes":    "20971520",
		"max.poll.interval.ms": "86400000",
	}

	if env.KafkaCreds != nil && len(env.KafkaCreds.Username) > 0 && len(env.KafkaCreds.Password) > 0 {
		cfg["security.protocol"] = "SASL_SSL"
		cfg["sasl.mechanism"] = "SCRAM-SHA-512"
		cfg["sasl.username"] = env.KafkaCreds.Username
		cfg["sasl.password"] = env.KafkaCreds.Password
	}

	if len(env.KafkaAutoOffsetReset) > 0 {
		cfg["auto.offset.reset"] = env.KafkaAutoOffsetReset
	}

	if len(env.KafkaDebugOptions) > 0 {
		cfg["debug"] = env.KafkaDebugOptions
	}

	return &Pool{
		Config:       &cfg,
		LogWatermark: env.KafkaLogWatermark,
		UseWatermark: env.KafkaUseWatermark,
	}
}
