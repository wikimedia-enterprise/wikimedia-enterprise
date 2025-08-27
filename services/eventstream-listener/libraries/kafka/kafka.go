// Package kafka is a wrapper around kafka library to provide default configuration and interfaces.
package kafka

import (
	"wikimedia-enterprise/services/eventstream-listener/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// NewProducer create new producer with default configuration.
func NewProducer(env *env.Environment) (*kafka.Producer, error) {
	cfg := kafka.ConfigMap{
		"bootstrap.servers":      env.KafkaBootstrapServers,
		"message.max.bytes":      "20971520",
		"compression.type":       "gzip",
		"queue.buffering.max.ms": 10,
		"go.delivery.reports":    false,
	}

	if env.KafkaCreds != nil && len(env.KafkaCreds.Username) > 0 && len(env.KafkaCreds.Password) > 0 {
		cfg["security.protocol"] = "SASL_SSL"
		cfg["sasl.mechanism"] = "SCRAM-SHA-512"
		cfg["sasl.username"] = env.KafkaCreds.Username
		cfg["sasl.password"] = env.KafkaCreds.Password
	}

	return kafka.NewProducer(&cfg)
}
