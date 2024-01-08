package filter_test

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"wikimedia-enterprise/services/event-bridge/packages/filter"
)

type filterTestSuite struct {
	suite.Suite
}

func (s *filterTestSuite) TestNew() {
	fr, err := filter.New()

	s.Assert().NoError(err)
	s.Assert().NotNil(fr.Projects)
	s.Assert().NotNil(fr.Namespaces)
}

func TestFilter(t *testing.T) {
	suite.Run(t, new(filterTestSuite))
}
