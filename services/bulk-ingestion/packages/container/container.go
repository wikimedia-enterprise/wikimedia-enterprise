// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	"wikimedia-enterprise/services/bulk-ingestion/libraries/kafka"
	"wikimedia-enterprise/services/bulk-ingestion/libraries/s3api"
	"wikimedia-enterprise/services/bulk-ingestion/libraries/stream"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cont := dig.New()

	for _, err := range []error{
		cont.Provide(env.New),
		cont.Provide(kafka.NewProducer),
		cont.Provide(stream.New),
		cont.Provide(wmf.NewAPI),
		cont.Provide(s3api.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cont, nil
}
