// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/libraries/auth"
	"wikimedia-enterprise/api/auth/libraries/cognito"
	"wikimedia-enterprise/api/auth/libraries/metrics"
	"wikimedia-enterprise/api/auth/libraries/redis"
	"wikimedia-enterprise/general/log"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cont := dig.New()

	for _, err := range []error{
		cont.Provide(env.New),
		cont.Provide(cognito.New),
		cont.Provide(redis.New),
		cont.Provide(auth.New),
		cont.Provide(metrics.New),
	} {
		if err != nil {
			log.Error(err, log.Tip("problem in container initialization"))
			return nil, err
		}
	}

	return cont, nil
}
