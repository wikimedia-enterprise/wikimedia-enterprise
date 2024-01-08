package filter_test

import (
	"testing"
	"wikimedia-enterprise/services/event-bridge/packages/filter"

	"github.com/stretchr/testify/suite"
)

type namespacesTestSuite struct {
	suite.Suite
	namespaces *filter.Namespaces
}

func (s *namespacesTestSuite) SetupTest() {
	s.namespaces = filter.NewNamespaces()
}

func (s *namespacesTestSuite) TestIsSupported() {
	for _, ns := range []int{0, 6, 14, 10} {
		s.Assert().True(s.namespaces.IsSupported(ns))
	}

	for _, ns := range []int{100, 500, 800} {
		s.Assert().False(s.namespaces.IsSupported(ns))
	}
}

func TestNamespaces(t *testing.T) {
	suite.Run(t, new(namespacesTestSuite))
}
