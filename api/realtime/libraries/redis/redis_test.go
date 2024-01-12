package redis_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/redis"

	"github.com/stretchr/testify/suite"
)

type redisTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *redisTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.RedisAddr = "REDIS_ADDR"
	s.env.RedisPassword = "REDIS_PASSWORD"
}

func (s *redisTestSuite) TestRedisSuccess() {
	cmdable, err := redis.New(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(cmdable)
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(redisTestSuite))
}
