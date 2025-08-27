package builder_test

import (
	"testing"

	"wikimedia-enterprise/services/structured-data/packages/builder"
	"wikimedia-enterprise/services/structured-data/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type deltaTestSuite struct {
	suite.Suite
	prev []string
	cur  []string
	del  *schema.Delta
}

func (s *deltaTestSuite) TestDelta() {
	s.Assert().Equal(s.del, builder.
		NewDeltaBuilder().
		Current(s.cur).
		Previous(s.prev).
		Build())
}

func TestDelta(t *testing.T) {
	for _, testCase := range []*deltaTestSuite{
		{
			prev: []string{"apple", "cat", "shoes", "cat", "apple", "apple", "viking"},
			cur:  []string{"apple", "shoes", "shoes", "viking", "viking"},
			del: &schema.Delta{
				Increase:             2,
				Decrease:             -4,
				Sum:                  -2,
				ProportionalIncrease: 1,
				ProportionalDecrease: -1.6666666666666665,
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
