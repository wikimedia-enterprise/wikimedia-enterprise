package container_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (c *containerTestSuite) TestNew() {
	cnt, err := container.New()

	c.Assert().NotNil(cnt)
	c.Assert().NoError(err)
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
