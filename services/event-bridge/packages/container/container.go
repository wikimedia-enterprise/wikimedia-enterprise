// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/libraries/kafka"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/libraries/producer"
	"wikimedia-enterprise/services/event-bridge/libraries/redis"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cont := dig.New()

	for _, err := range []error{
		cont.Provide(env.New),
		cont.Provide(langid.NewDictionary),
		cont.Provide(kafka.NewProducer),
		cont.Provide(redis.NewClient),
		cont.Provide(producer.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cont, nil
}
