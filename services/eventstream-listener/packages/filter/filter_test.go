package filter_test

import (
	"testing"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/submodules/config"

	"github.com/stretchr/testify/suite"
)

type filterTestSuite struct {
	suite.Suite
}

func (s *filterTestSuite) TestNew() {
	cfg, err := config.New()
	fr := filter.New(cfg)

	s.Assert().NoError(err)
	s.Assert().True(fr.IsSupported("dewiki", 6, 0))
	s.Assert().True(fr.IsSupported("dewiki", 6, 100))
	s.Assert().True(fr.IsSupported("dewiki", 6))
	s.Assert().True(fr.IsSupported("commonswiki", 6, 6))
	s.Assert().True(fr.IsSupported("commonswiki", 6, 0))

	for _, prj := range []string{"enwiki", "commonswiki"} {
		lng, err := fr.ResolveLang(prj)
		s.Assert().NoError(err)
		s.Assert().Equal("en", lng)
	}

	s.Assert().False(fr.IsSupported("tigwiki", 6))
	s.Assert().False(fr.IsSupported("enwiki", 5))
	s.Assert().False(fr.IsSupported("commonswiki", 100, 2))
	s.Assert().False(fr.IsSupported("commonswiki", 100))

	lng, err := fr.ResolveLang("tigwiki")
	s.Assert().Error(err)
	s.Assert().Equal("", lng)
}

func TestFilter(t *testing.T) {
	suite.Run(t, new(filterTestSuite))
}
