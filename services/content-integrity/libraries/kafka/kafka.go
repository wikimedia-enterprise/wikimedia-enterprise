// Package kafka is a wrapper around kafka library to provide default configuration and interfaces.
package kafka

import (
	"wikimedia-enterprise/services/content-integrity/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

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
