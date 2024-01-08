package exponential

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type exponentialTestSuite struct {
	suite.Suite
	tc  map[uint]uint
	bse uint
}

func (s *exponentialTestSuite) SetupSuite() {
	s.tc = map[uint]uint{
		0: 1,
		1: 2,
		2: 4,
		3: 8,
		4: 16,
		5: 32,
		6: 64,
		7: 128,
		8: 256,
		9: 512,
	}
	s.bse = 2
}

func (s *exponentialTestSuite) TestGetNth() {
	for input, expectedOutput := range s.tc {
		result := GetNth(input, s.bse)
		s.Equal(expectedOutput, result)
	}
}

func TestExponential(t *testing.T) {
	suite.Run(t, new(exponentialTestSuite))
}
