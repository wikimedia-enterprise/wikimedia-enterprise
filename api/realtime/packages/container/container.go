// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/auth"
	"wikimedia-enterprise/api/realtime/libraries/enforcer"
	"wikimedia-enterprise/api/realtime/libraries/kafka"
	"wikimedia-enterprise/api/realtime/libraries/ksqldb"
	"wikimedia-enterprise/api/realtime/libraries/metrics"
	"wikimedia-enterprise/api/realtime/libraries/redis"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/libraries/stream"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"go.uber.org/dig"
)

// Schemas provides supported schemas for resolver.
// Update to add more support such as entity, structured-contents.
var Schemas = map[string]interface{}{
	schema.KeyTypeArticle: new(schema.Article),
}

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
		cnt.Provide(kafka.NewPool),
		cnt.Provide(stream.New),
		cnt.Provide(stream.NewGetterCreator),

		cnt.Provide(func(e *env.Environment) (resolver.Resolvers, error) {
			if e.UseKsqldb {
				return resolver.NewResolvers(Schemas, func(r *resolver.Resolver) {
					r.Keywords = map[string]string{
						"namespace": "`NAMESPACE`",
					}
				})
			}

			return resolver.NewResolvers(Schemas)
		})} {
		if err != nil {
			return nil, err
		}
	}

	return cnt, nil
}
