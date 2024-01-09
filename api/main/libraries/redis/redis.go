// Package redis holds initialization of the Redis connection
// for dependency injection.
package redis

import (
	"crypto/tls"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/general/log"

	"github.com/go-redis/redis/v8"
)

// New creates redis cache.
func New(env *env.Environment) (redis.Cmdable, error) {
	cfg := &redis.Options{
		Addr: env.RedisAddr,
	}

	if len(env.RedisPassword) > 0 {
		cfg.Password = env.RedisPassword
		cfg.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	if len(cfg.Password) == 0 {
		log.Warn("missing password env variable for redis connection")
	}

	return redis.NewClient(cfg), nil
}
