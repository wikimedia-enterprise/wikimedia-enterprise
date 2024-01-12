package redis

import (
	"crypto/tls"
	"wikimedia-enterprise/api/realtime/config/env"

	"github.com/go-redis/redis/v8"
)

// New creates redis cache instance.
func New(env *env.Environment) (redis.Cmdable, error) {
	cfg := &redis.Options{
		Addr: env.RedisAddr,
	}

	if len(env.RedisPassword) > 0 {
		cfg.Password = env.RedisPassword
		cfg.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return redis.NewClient(cfg), nil
}
