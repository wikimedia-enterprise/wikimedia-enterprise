package container_test

import (
	"os"
	"testing"

	"wikimedia-enterprise/services/on-demand/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) SetupSuite() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:8001")
}

func (s *containerTestSuite) TestNew() {
	cont, err := container.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(cont)
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
