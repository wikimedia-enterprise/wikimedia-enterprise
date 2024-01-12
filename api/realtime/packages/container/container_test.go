package container_test

import (
	"testing"

	"wikimedia-enterprise/api/realtime/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) TestNew() {
	cnt, err := container.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
