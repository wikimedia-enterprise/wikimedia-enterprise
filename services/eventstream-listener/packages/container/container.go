// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/libraries/kafka"
	"wikimedia-enterprise/services/eventstream-listener/libraries/producer"
	"wikimedia-enterprise/services/eventstream-listener/libraries/redis"
	trc "wikimedia-enterprise/services/eventstream-listener/libraries/tracing"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/operations"
	"wikimedia-enterprise/services/eventstream-listener/packages/transformer"
	"wikimedia-enterprise/services/eventstream-listener/submodules/config"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cont := dig.New()

	for _, err := range []error{
		cont.Provide(env.New),
		cont.Provide(trc.NewAPI),
		cont.Provide(kafka.NewProducer),
		cont.Provide(redis.NewClient),
		cont.Provide(producer.New),
		cont.Provide(config.New),
		cont.Provide(filter.New),
		cont.Provide(transformer.New),
		cont.Provide(operations.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cont, nil
}
