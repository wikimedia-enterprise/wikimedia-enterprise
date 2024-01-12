// Package kafka is a wrapper around kafka library to provide default configuration and interfaces.
package kafka

import (
	"log"
	"wikimedia-enterprise/services/structured-data/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// NewProducer create new producer with default configuration.
func NewProducer(env *env.Environment) (*kafka.Producer, error) {
	cfg := kafka.ConfigMap{
		"bootstrap.servers":      env.KafkaBootstrapServers,
		"compression.type":       "gzip",
		"message.max.bytes":      "20971520",
		"go.batch.producer":      true,
		"queue.buffering.max.ms": 10,
		// "go.delivery.reports":    false,
	}

	if env.KafkaCreds != nil && len(env.KafkaCreds.Username) > 0 && len(env.KafkaCreds.Password) > 0 {
		cfg["security.protocol"] = "SASL_SSL"
		cfg["sasl.mechanism"] = "SCRAM-SHA-512"
		cfg["sasl.username"] = env.KafkaCreds.Username
		cfg["sasl.password"] = env.KafkaCreds.Password
	}

	psr, err := kafka.NewProducer(&cfg)

	if err != nil {
		return nil, err
	}

	go func() {
		for evt := range psr.Events() {
			switch evt := evt.(type) {
			case *kafka.Message:
				if evt.TopicPartition.Error != nil {
					log.Printf("delivery failed for topic partition `%v` with error `%v`", evt.TopicPartition, evt.TopicPartition.Error)
				}
			}
		}
	}()

	return psr, nil
}

// NewConsumer creates new kafka consumer with default configuration.
func NewConsumer(env *env.Environment) (*kafka.Consumer, error) {
	cfg := kafka.ConfigMap{
		"bootstrap.servers": env.KafkaBootstrapServers,
		"group.id":          env.KafkaConsumerGroupID,
		"message.max.bytes": "20971520",
	}

	if env.KafkaCreds != nil && len(env.KafkaCreds.Username) > 0 && len(env.KafkaCreds.Password) > 0 {
		cfg["security.protocol"] = "SASL_SSL"
		cfg["sasl.mechanism"] = "SCRAM-SHA-512"
		cfg["sasl.username"] = env.KafkaCreds.Username
		cfg["sasl.password"] = env.KafkaCreds.Password
	}

	return kafka.NewConsumer(&cfg)
}
