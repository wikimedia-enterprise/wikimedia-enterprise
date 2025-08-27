package redis_test

import (
	"testing"

	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/libraries/redis"

	"github.com/stretchr/testify/suite"
)

type redisTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *redisTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.RedisAddr = "localhost:8080"
	s.env.RedisPassword = "12345"
}

func (s *redisTestSuite) TestNew() {
	s.Assert().NotNil(redis.NewClient(s.env))
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(redisTestSuite))
}
