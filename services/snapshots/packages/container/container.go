// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/kafka"
	"wikimedia-enterprise/services/snapshots/libraries/s3api"
	"wikimedia-enterprise/services/snapshots/libraries/stream"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(uploader.New),
		cnt.Provide(s3api.New),
		cnt.Provide(stream.New),
		cnt.Provide(kafka.NewPool),
		cnt.Provide(config.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}
