package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/builder"
	"wikimedia-enterprise/services/structured-data/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type diffBuilderTestSuite struct {
	suite.Suite
	bdr *builder.DiffBuilder
}

func (s *diffBuilderTestSuite) SetupTest() {
	s.bdr = builder.NewDiffBuilder()
}

func (s *diffBuilderTestSuite) TestSize() {
	diff := s.bdr.Size("current wikitext value", "previous wikitext").Build()

	s.Assert().Equal(diff.Size, &schema.Size{
		UnitText: "B",
		Value:    5,
	})
}

func (s *diffBuilderTestSuite) TestDictNonDictWords() {
	diff := s.bdr.DictNonDictWords(
		[]string{"apple"},
		[]string{"yui"},
		[]string{"apple", "pear"},
		[]string{""},
	).Build()

	s.Assert().Equal(diff, &schema.Diff{
		DictionaryWords: &schema.Delta{
			Increase:             1,
			ProportionalIncrease: 1,
			Sum:                  1,
		},
		NonDictionaryWords: &schema.Delta{
			Increase:             1,
			Decrease:             -1,
			ProportionalIncrease: 1,
			ProportionalDecrease: -1,
		},
	})
}

func (s *diffBuilderTestSuite) TestInformalWords() {
	diff := s.bdr.InformalWords([]string{"info"}, []string{}).Build()

	s.Assert().Equal(diff, &schema.Diff{
		InformalWords: &schema.Delta{
			Decrease:             -1,
			Sum:                  -1,
			ProportionalDecrease: -1,
		},
	})
}

func (s *diffBuilderTestSuite) TestNonSafeWords() {
	diff := s.bdr.NonSafeWords([]string{"gringo"}, []string{}).Build()

	s.Assert().Equal(diff, &schema.Diff{
		NonSafeWords: &schema.Delta{
			Decrease:             -1,
			Sum:                  -1,
			ProportionalDecrease: -1,
		},
	})
}

func (s *diffBuilderTestSuite) TestGetUppercaseWords() {
	diff := s.bdr.UpperCaseWords([]string{"TEST"}, []string{}).Build()

	s.Assert().Equal(diff, &schema.Diff{
		UppercaseWords: &schema.Delta{
			Decrease:             -1,
			Sum:                  -1,
			ProportionalDecrease: -1,
		},
	})
}

func (s *diffBuilderTestSuite) TestLongestNewRepeatedCharacter() {
	for _, tc := range []struct {
		cur  string
		prev string
		res  int
	}{
		// Return 4, since length of current (4) is greater than previous (1)
		{"aaaa", "abcd", 4},
		// Return 4, since length of current (4) is greater than previous (1)
		{"AAAA", "abcd", 4},
		// Return 4, since length of current (4) is greater than previous (1)
		{"aAaA", "abcd", 4},
		// Return 4, since length of current (4) is greater than previous (0)
		{"aaaa", "", 4},
		// Return 1, since length of current (1) is less than previous (4)
		{"abcd", "aaaa", 1},
		// Return 1, since length of current (1) is less than previous (4)
		{"abcd", "aAaAbc", 1},
		// Return 1, since length of current (0) is less than previous (1)
		{"", "test", 1},
		// Return 1, since both strings have the same length of repeating characters (4)
		{"aaaa", "aaaa", 1},
		//Return 1, since both strings have the same length of repeating characters (0)
		{"", "", 1},
	} {
		diff := s.bdr.LongestNewRepeatedCharacter(tc.cur, tc.prev).Build()
		s.Assert().Equal(tc.res, diff.LongestNewRepeatedCharacter)
	}
}

func TestDiffBuilderSuite(t *testing.T) {
	suite.Run(t, new(diffBuilderTestSuite))
}
