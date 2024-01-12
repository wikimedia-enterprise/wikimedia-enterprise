// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/general/parser"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/libraries/content"
	"wikimedia-enterprise/services/structured-data/libraries/kafka"
	"wikimedia-enterprise/services/structured-data/libraries/stream"
	"wikimedia-enterprise/services/structured-data/libraries/text"
	"wikimedia-enterprise/services/structured-data/libraries/wmf"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(kafka.NewProducer),
		cnt.Provide(kafka.NewConsumer),
		cnt.Provide(text.New),
		cnt.Provide(content.New),
		cnt.Provide(stream.New),
		cnt.Provide(schema.NewRetry),
		cnt.Provide(aggregate.New),
		cnt.Provide(wmf.NewAPI),
		cnt.Provide(parser.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}