package filter_test

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"wikimedia-enterprise/services/event-bridge/packages/filter"
)

type projectsTestSuite struct {
	suite.Suite
	projects *filter.Projects
}

func (s *projectsTestSuite) SetupTest() {
	var err error
	s.projects, err = filter.NewProjects()
	s.Assert().NoError(err)
}

func (s *projectsTestSuite) TestIsSupported() {
	for _, dbname := range []string{"advisorswiki", "amwikimedia", "arbcom_dewiki"} {
		s.Assert().False(s.projects.IsSupported(dbname))
	}

	for _, dbname := range []string{"enwiki", "afwikibooks", "simplewiki"} {
		s.Assert().True(s.projects.IsSupported(dbname))
	}
}

func TestProjects(t *testing.T) {
	suite.Run(t, new(projectsTestSuite))
}
