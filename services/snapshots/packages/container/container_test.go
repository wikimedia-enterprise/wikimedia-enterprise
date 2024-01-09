package container_test

import (
	"os"
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"
	"wikimedia-enterprise/services/snapshots/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) SetupSuite() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:8001")
	os.Setenv("SCHEMA_REGISTRY_URL", "localhost:9121")
}

func (s *containerTestSuite) TestNew() {
	cnt, err := container.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)
	s.Assert().NotPanics(func() {
		s.Assert().NoError(cnt.Invoke(func(env *env.Environment, upd uploader.Uploader) {
			s.Assert().NotNil(env)
			s.Assert().NotNil(upd)
		}))
	})
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
