// Package kafka wraps kafka communications and consumer pooling into interfaces.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/submodules/log"

	"github.com/avast/retry-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ErrNoPartitions happens when no partitions are provided to read.
var ErrNoPartitions = errors.New("no partitions provided to read")

// KafkaConsumer wraps kafka consumer into interface.
type KafkaConsumer interface {
	Assign([]kafka.TopicPartition) error
	ReadMessage(time.Duration) (*kafka.Message, error)
	Poll(int) kafka.Event
	OffsetsForTimes([]kafka.TopicPartition, int) ([]kafka.TopicPartition, error)
	Close() error
}

// ReadAllCloser is an interface that wraps kafka consumer to retrieve all messages from the topic.
type ReadAllCloser interface {
	ReadAll(ctx context.Context, params *ReadParams, topic string, partitions []int, callback func(msg *kafka.Message) error) error
	SetPartitionOffsets(topic string, partitions []int, r *ReadParams) ([]kafka.TopicPartition, error)
	Close() error
}

// Consumer kafka library wrapper to read all messages by partition from the topic.
type Consumer struct {
	Consumer KafkaConsumer
}

// ReadParams configures the offsets ReadAll will start reading from.
// If more than one of these apply to a given partition, the first applicable field will be used.
// If none apply, partitions get "latest".
type ReadParams struct {
	Since              time.Time
	SincePerPartition  map[int]time.Time
	OffsetPerPartition map[int]int64
}

// ReadAll retrieves all messages from the provided partitions and topics.
func (c *Consumer) ReadAll(ctx context.Context, params *ReadParams, tpc string, pts []int, cbk func(msg *kafka.Message) error) error {
	if len(pts) == 0 {
		return ErrNoPartitions
	}

	tps, err := c.SetPartitionOffsets(tpc, pts, params)

	if err != nil {
		return fmt.Errorf("error retrieving offsets: %v", err)
	}

	if err := c.Consumer.Assign(tps); err != nil {
		return fmt.Errorf("error assigning partitions: %v", err)
	}

	// Two approaches are implemented: consumer.Poll() and consumer.Read().
	// Poll() should be more efficient as it fetches batches from the broker.
	polling := true

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// keep going
		}

		if polling {
			ev := c.Consumer.Poll(1000)

			switch e := ev.(type) {
			case *kafka.Message:
				if err := cbk(e); err != nil {
					// treat non-nil as stop streaming
					return err
				}
			case kafka.PartitionEOF:
				// keep going
			case kafka.Error:
				log.Error("kafka error", log.Any("error", ev.(kafka.Error)))
			}
		} else {
			msg, err := c.Consumer.ReadMessage(time.Second * 1)

			if err != nil {
				switch err.Error() {
				case kafka.ErrTimedOut.String():
					continue
				default:
					log.Error("error in Kafka consumer", log.Any("error", err))
				}
			} else {
				if err := cbk(msg); err != nil {
					return err
				}
			}
		}
	}
}

// Close closes consumer connection.
func (c *Consumer) Close() error {
	return c.Consumer.Close()
}

func (c *Consumer) SetPartitionOffsets(topic string, partitions []int, r *ReadParams) ([]kafka.TopicPartition, error) {
	ptsToResolve := []kafka.TopicPartition{}
	result := []kafka.TopicPartition{}

	for _, pID := range partitions {
		offset := kafka.OffsetEnd
		resolve := false
		if !r.Since.IsZero() {
			offset = kafka.Offset(r.Since.UnixMilli())
			resolve = true
		} else if since, hasSince := r.SincePerPartition[pID]; hasSince {
			offset = kafka.Offset(since.UnixMilli())
			resolve = true
		} else if pOffset, hasOffset := r.OffsetPerPartition[pID]; hasOffset {
			offset = kafka.Offset(pOffset)
		}

		ptn := kafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(pID), // #nosec G115
			Offset:    offset,
		}

		if resolve {
			ptsToResolve = append(ptsToResolve, ptn)
		} else {
			result = append(result, ptn)
		}
	}

	if len(ptsToResolve) > 0 {
		var resolved []kafka.TopicPartition

		gof := func() (err error) {
			resolved, err = c.Consumer.OffsetsForTimes(ptsToResolve, 60000)
			return err
		}

		if err := retry.Do(gof); err != nil {
			return nil, err.(retry.Error)[0]
		}

		result = append(result, resolved...)
	}

	return result, nil
}

// ConsumerGetter wrapper for the consumer pool struct to get a consumer by group id.
type ConsumerGetter interface {
	GetConsumer(string) (ReadAllCloser, error)
}

// Pool provides the ability to generate new consumer by group id.
type Pool struct {
	Config *kafka.ConfigMap
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

	return &Consumer{Consumer: kcs}, nil
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

	for k, v := range env.KafkaExtraConfig {
		cfg[k] = v
	}

	return &Pool{
		Config: &cfg,
	}
}
