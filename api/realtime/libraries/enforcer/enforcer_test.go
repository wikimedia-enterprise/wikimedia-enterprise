package enforcer_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/enforcer"

	"github.com/stretchr/testify/suite"
)

type enforcerTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *enforcerTestSuite) SetupSuite() {
	s.env = &env.Environment{
		AccessModel:  new(env.AccessModel),
		AccessPolicy: new(env.AccessPolicy),
	}

	_ = s.env.AccessPolicy.UnmarshalEnvironmentValue("")
	_ = s.env.AccessModel.UnmarshalEnvironmentValue("")
}

func (s *enforcerTestSuite) TestEnforcerSuccess() {
	enf, err := enforcer.New(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(enf)
}

func TestEnforcer(t *testing.T) {
	suite.Run(t, new(enforcerTestSuite))
}
