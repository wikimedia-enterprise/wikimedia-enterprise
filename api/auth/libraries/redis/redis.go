// Package redis simple wrapper around redis client for providing default configuration.
package redis

import (
	"crypto/tls"
	"wikimedia-enterprise/api/auth/config/env"

	"github.com/go-redis/redis/v8"
)

// New used to initialize a redis client.
func New(env *env.Environment) redis.Cmdable {
	cfg := &redis.Options{
		Addr: env.RedisAddr,
	}

	if len(env.RedisPassword) > 0 {
		cfg.Password = env.RedisPassword
		cfg.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return redis.NewClient(cfg)
}
