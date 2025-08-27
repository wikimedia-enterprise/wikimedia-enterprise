package container_test

import (
	"os"
	"testing"

	"wikimedia-enterprise/services/eventstream-listener/packages/container"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"
	"wikimedia-enterprise/services/eventstream-listener/submodules/tracing"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) SetupSuite() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:8001")
	os.Setenv("REDIS_ADDR", "localhost:8002")
	os.Setenv("SCHEMA_REGISTRY_URL", "localhost:8085")
}

func (s *containerTestSuite) TestNew() {
	cont, err := container.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(cont)
	s.Assert().NotPanics(func() {
		s.Assert().NoError(cont.Invoke(func(redis redis.Cmdable, prod schema.Producer, trc tracing.Tracer) {
			s.Assert().NotNil(redis)
			s.Assert().NotNil(prod)
			s.Assert().NotNil(trc)
		}))
	})
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
