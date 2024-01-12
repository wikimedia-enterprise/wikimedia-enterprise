package auth_test

import (
	"testing"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/libraries/auth"

	"github.com/stretchr/testify/suite"
)

type authTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *authTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.AWSRegion = "AWS_REGION"
	s.env.CognitoClientID = "COGNITO_CLIENT_ID"
	s.env.CognitoSecret = "COGNITO_CLIENT_SECRET"
}

func (s *authTestSuite) TestauthSuccess() {
	idp := auth.New(s.env)
	s.Assert().NotNil(idp)
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}
