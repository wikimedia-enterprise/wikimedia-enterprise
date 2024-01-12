package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type diffTestSuite struct {
	suite.Suite
	diff *Diff
}

func (s *diffTestSuite) SetupTest() {
	s.diff = &Diff{
		LongestNewRepeatedCharacter: 5,
		DictionaryWords: &Delta{
			Increase: 10,
		},
		NonDictionaryWords: &Delta{
			Decrease: 5,
		},
		NonSafeWords: &Delta{
			Increase: 1,
		},
		InformalWords: &Delta{
			Decrease: 2,
		},
		UppercaseWords: &Delta{
			Increase: 5,
		},
		Size: &Size{
			Value:    500,
			UnitText: "MB",
		},
	}
}

func (s *diffTestSuite) TestNewDiffSchema() {
	sch, err := NewDiffSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.diff)
	s.Assert().NoError(err)

	diff := new(Diff)
	s.Assert().NoError(avro.Unmarshal(sch, data, diff))
	s.Assert().Equal(s.diff.LongestNewRepeatedCharacter, diff.LongestNewRepeatedCharacter)
	s.Assert().Equal(s.diff.DictionaryWords, diff.DictionaryWords)
	s.Assert().Equal(s.diff.NonDictionaryWords, diff.NonDictionaryWords)
	s.Assert().Equal(s.diff.NonSafeWords, diff.NonSafeWords)
	s.Assert().Equal(s.diff.InformalWords, diff.InformalWords)
	s.Assert().Equal(s.diff.UppercaseWords, diff.UppercaseWords)
	s.Assert().Equal(s.diff.Size, diff.Size)
}

func TestDiff(t *testing.T) {
	suite.Run(t, new(diffTestSuite))
}
