package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/builder"

	"github.com/stretchr/testify/suite"
)

type deltaTableTestSuite struct {
	suite.Suite
	tokens []string
	table  builder.FrequencyTable
}

func (s *deltaTableTestSuite) TestDeltaTable() {
	s.Assert().Equal(s.table, builder.
		NewFrequencyTableBuilder().
		Tokens(s.tokens).
		Build())
}

func TestDeltaTable(t *testing.T) {
	for _, testCase := range []*deltaTableTestSuite{
		{
			tokens: []string{"apple", "cat", "shoes", "cat", "apple", "apple", "viking"},
			table: map[string]int{
				"apple":  3,
				"cat":    2,
				"shoes":  1,
				"viking": 1,
			},
		},
		{
			tokens: []string{"apple", "shoes", "shoes", "viking", "viking"},
			table: map[string]int{
				"apple":  1,
				"shoes":  2,
				"viking": 2,
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
