// Package redis simple wrapper around redis client for providing default configuration.
package redis

import (
	"crypto/tls"
	"wikimedia-enterprise/services/content-integrity/config/env"

	redis "github.com/redis/go-redis/v9"
)

// NewClient create new redis client.
func NewClient(env *env.Environment) redis.Cmdable {
	cfg := &redis.Options{
		Addr: env.RedisAddr,
	}

	if len(env.RedisPassword) > 0 {
		cfg.Password = env.RedisPassword
		cfg.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return redis.NewClient(cfg)
}
