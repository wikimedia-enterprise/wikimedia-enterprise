package wmf

import (
	"testing"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"

	"github.com/stretchr/testify/suite"
)

type wmfTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *wmfTestSuite) SetupSuite() {
	s.env = new(env.Environment)
}

func (s *wmfTestSuite) TestNew() {
	clt := NewAPI(s.env)
	s.Assert().NotNil(clt)
}

func TestWMF(t *testing.T) {
	suite.Run(t, new(wmfTestSuite))
}
