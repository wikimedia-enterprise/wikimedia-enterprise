// Package container provides the dependency injection management setup.
// Injects and resolves default dependencies.
package container

import (
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/libraries/auth"
	"wikimedia-enterprise/api/main/libraries/enforcer"
	"wikimedia-enterprise/api/main/libraries/metrics"
	"wikimedia-enterprise/api/main/libraries/redis"
	"wikimedia-enterprise/api/main/libraries/s3"
	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/wmf"

	"go.uber.org/dig"
)

// New create container with dependency injection and default dependencies.
func New() (*dig.Container, error) {
	cnt := dig.New()

	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(s3.New),
		cnt.Provide(redis.New),
		cnt.Provide(auth.New),
		cnt.Provide(enforcer.New),
		cnt.Provide(config.New),
		cnt.Provide(metrics.New),
		cnt.Provide(wmf.NewAPI),
	} {
		if err != nil {
			log.Error(err, log.Tip("problem creating container for dependency injection"))
			return nil, err
		}
	}

	return cnt, nil
}
