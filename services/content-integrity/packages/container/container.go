// Package container provides the dependency injection management setup.
// Injects and resolves all dependencies.
package container

import (
	"wikimedia-enterprise/services/content-integrity/config/env"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"
	"wikimedia-enterprise/services/content-integrity/libraries/integrity"
	"wikimedia-enterprise/services/content-integrity/libraries/kafka"
	"wikimedia-enterprise/services/content-integrity/libraries/redis"
	"wikimedia-enterprise/services/content-integrity/libraries/stream"

	"go.uber.org/dig"
)

// New creates new container instance for dependency injection.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(redis.NewClient),
		cnt.Provide(collector.NewArticle),
		cnt.Provide(kafka.NewConsumer),
		cnt.Provide(stream.New),
		cnt.Provide(integrity.New),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}
