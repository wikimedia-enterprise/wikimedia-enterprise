// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/auth"
	"wikimedia-enterprise/api/realtime/libraries/enforcer"
	"wikimedia-enterprise/api/realtime/libraries/ksqldb"
	"wikimedia-enterprise/api/realtime/libraries/metrics"
	"wikimedia-enterprise/api/realtime/libraries/redis"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/general/schema"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(redis.New),
		cnt.Provide(auth.New),
		cnt.Provide(enforcer.New),
		cnt.Provide(ksqldb.New),
		cnt.Provide(metrics.New),
		cnt.Provide(func() (*resolver.Resolver, error) {
			return resolver.New(new(schema.Article))
		}),
	} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}
