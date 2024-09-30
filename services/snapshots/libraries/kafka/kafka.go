// Package kafka wraps kafka communications and consumer pooling into interfaces.
package kafka

import (
	"context"
	"errors"
	"log"
	"time"
	"wikimedia-enterprise/services/snapshots/config/env"

	"github.com/avast/retry-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ErrReadAllEndMessage this message can be passed to ReadAll callback function to stop the consumption.
var ErrReadAllEndMessage = errors.New("read all end message")

// KafkaConsumer wraps kafka consumer into interface.
type KafkaConsumer interface {
	Assign([]kafka.TopicPartition) error
	ReadMessage(time.Duration) (*kafka.Message, error)
	GetMetadata(*string, bool, int) (*kafka.Metadata, error)
	GetWatermarkOffsets(string, int32) (int64, int64, error)
	OffsetsForTimes([]kafka.TopicPartition, int) ([]kafka.TopicPartition, error)
	Close() error
}

// ReadAllCloser is an interface that wraps kafka consumer to retrieve all messages from the topic.
type ReadAllCloser interface {
	ReadAll(context.Context, int, string, []int, func(msg *kafka.Message) error) error
	GetWatermarkOffsets(string, int) (int, int, error)
	GetMetadata(string) (*kafka.TopicMetadata, error)
	Close() error
}

// Consumer kafka library wrapper to read all messages by partition from the topic.
type Consumer struct {
	Consumer KafkaConsumer
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
	lof, hof, err := c.Consumer.GetWatermarkOffsets(tpc, int32(ptn))
	return int(lof), int(hof), err
}

// ReadAll retrieves all messages from provided partitions and topics.
func (c *Consumer) ReadAll(ctx context.Context, snc int, tpc string, pts []int, cbk func(msg *kafka.Message) error) error {
	tps := []kafka.TopicPartition{}

	for _, pid := range pts {
		tpn := kafka.TopicPartition{
			Topic:     &tpc,
			Partition: int32(pid),
			Offset:    kafka.OffsetBeginning,
		}

		if snc > 0 {
			if err := tpn.Offset.Set(snc); err != nil {
				return err
			}
		}

		tps = append(tps, tpn)
	}

	if snc > 0 {
		var otp []kafka.TopicPartition

		gof := func() (err error) {
			otp, err = c.Consumer.OffsetsForTimes(tps, 60000)
			return err
		}

		if err := retry.Do(gof); err != nil {
			return err.(retry.Error)[0]
		}

		tps = otp
	}

	if err := c.Consumer.Assign(tps); err != nil {
		return err
	}

	for {
		msg, err := c.Consumer.ReadMessage(time.Second * 1)

		select {
		case <-ctx.Done():
			return nil
		default:
			if err != nil {
				switch err.Error() {
				case kafka.ErrTransport.String(), kafka.ErrBadMsg.String():
					log.Printf("read all, message error: `%v`\n", err)
				case kafka.ErrTimedOut.String():
					log.Printf("read all, finished with timeout err: `%v`\n", err)
					return nil
				default:
					log.Printf("read all, default error finished with err: `%v`\n", err)
					return err
				}
			} else {
				if err := cbk(msg); err != nil {
					if err == ErrReadAllEndMessage {
						log.Printf("read all, finishing with end err: `%v`\n", err)
						return nil
					}

					log.Printf("read all, finishing with unknown err: `%v`\n", err)
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

	if len(env.KafkaAutoOffsetReset) > 0 {
		cfg["auto.offset.reset"] = env.KafkaAutoOffsetReset
	}

	return &Pool{
		Config: &cfg,
	}
}
