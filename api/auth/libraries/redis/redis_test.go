package redis_test

import (
	"testing"

	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/libraries/redis"

	"github.com/stretchr/testify/suite"
)

type redisTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *redisTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.RedisAddr = "localhost:8080"
	s.env.RedisPassword = "password"
}

func (s *redisTestSuite) TestNew() {
	s.Assert().NotNil(redis.New(s.env))
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(redisTestSuite))
}
