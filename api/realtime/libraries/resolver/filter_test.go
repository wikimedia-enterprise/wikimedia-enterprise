package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type filterTestSuite struct {
	suite.Suite
	filters []resolver.Filter
}

func (s *filterTestSuite) TestGetFilter() {
	fr := resolver.GetFilter(s.filters...)

	if len(s.filters) == 0 {
		s.Assert().True(fr(new(resolver.Field)))
	}

	if len(s.filters) > 0 {
		result := false

		for _, fr := range s.filters {
			if ok := fr(new(resolver.Field)); ok {
				result = true
				break
			}
		}

		s.Assert().Equal(result, fr(new(resolver.Field)))
	}
}

func TestFilter(t *testing.T) {
	for _, testCase := range []*filterTestSuite{
		{
			filters: []resolver.Filter{
				func(fld *resolver.Field) bool {
					return false
				},
			},
		},
		{
			filters: []resolver.Filter{
				func(fld *resolver.Field) bool {
					return false
				},
				func(fld *resolver.Field) bool {
					return true
				},
			},
		},
		{
			filters: []resolver.Filter{
				func(fld *resolver.Field) bool {
					return false
				},
				func(fld *resolver.Field) bool {
					return true
				},
				func(fld *resolver.Field) bool {
					return false
				},
			},
		},
		{
			filters: []resolver.Filter{},
		},
	} {

		suite.Run(t, testCase)
	}
}
