package content_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/content"

	"github.com/stretchr/testify/suite"
)

type integrityTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *integrityTestSuite) SetupSuite() {
	s.env = new(env.Environment)
}

func (s *integrityTestSuite) TestNew() {
	clt, err := content.New(s.env)

	if err == nil {
		s.Assert().NotNil(clt)
	} else {
		s.Assert().Nil(clt)
	}
}

func TestIntegrity(t *testing.T) {
	suite.Run(t, new(integrityTestSuite))
}
