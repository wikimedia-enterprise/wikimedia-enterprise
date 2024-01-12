package abstract_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/abstract"

	"github.com/stretchr/testify/suite"
)

type isValidTestSuite struct {
	suite.Suite
	nsp int
	cml string
	res bool
}

func (s *isValidTestSuite) TestIsValid() {
	s.Assert().Equal(
		s.res,
		abstract.IsValid(s.nsp, s.cml),
	)
}

func TestIsValid(t *testing.T) {
	for _, testCase := range []*isValidTestSuite{
		{
			nsp: 10,
			cml: "html",
			res: false,
		},
		{
			nsp: 0,
			cml: "wikitext",
			res: true,
		},
	} {
		suite.Run(t, testCase)
	}
}
