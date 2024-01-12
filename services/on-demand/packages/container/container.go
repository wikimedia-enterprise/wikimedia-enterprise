// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/libraries/kafka"
	"wikimedia-enterprise/services/on-demand/libraries/s3api"
	"wikimedia-enterprise/services/on-demand/libraries/storage"
	"wikimedia-enterprise/services/on-demand/libraries/stream"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(kafka.NewProducer),
		cnt.Provide(kafka.NewConsumer),
		cnt.Provide(stream.New),
		cnt.Provide(storage.New),
		cnt.Provide(s3api.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}