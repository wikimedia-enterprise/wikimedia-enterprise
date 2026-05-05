// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/services/structured-contents/config/env"
	"wikimedia-enterprise/services/structured-contents/libraries/aggregate"
	"wikimedia-enterprise/services/structured-contents/libraries/kafka"
	"wikimedia-enterprise/services/structured-contents/libraries/prometheus"
	"wikimedia-enterprise/services/structured-contents/libraries/stream"
	"wikimedia-enterprise/services/structured-contents/libraries/tracing"
	"wikimedia-enterprise/services/structured-contents/libraries/wmf"
	"wikimedia-enterprise/services/structured-contents/submodules/config"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
	"wikimedia-enterprise/services/structured-contents/submodules/subscriber"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(kafka.NewProducer),
		cnt.Provide(kafka.NewConsumer),
		cnt.Provide(subscriber.New),
		cnt.Provide(stream.New),
		cnt.Provide(wmf.NewAPI),
		cnt.Provide(parser.New),
		cnt.Provide(aggregate.New),
		cnt.Provide(config.New),
		cnt.Provide(prometheus.New),
		cnt.Provide(tracing.NewAPI),
		cnt.Provide(schema.NewRetry),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}
