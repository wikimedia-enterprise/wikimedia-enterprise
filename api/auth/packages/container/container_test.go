package container_test

import (
	"os"
	"testing"

	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) SetupSuite() {
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("AWS_ID", "foo")
	os.Setenv("AWS_KEY", "bar")
	os.Setenv("COGNITO_CLIENT_ID", "ab21ad1")
	os.Setenv("COGNITO_SECRET", "dv34ad21")
	os.Setenv("COGNITO_USER_POOL_ID", "ffd2as1w")
	os.Setenv("COGNITO_USER_GROUP", "free")
	os.Setenv("ACCESS_POLICY", "")
	os.Setenv("GROUP_DOWNLOAD_LIMIT", "1000")
}

func (s *containerTestSuite) TestNew() {
	cont, err := container.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(cont)
	s.Assert().NotPanics(func() {
		s.Assert().NoError(cont.Invoke(func(env *env.Environment) {
			s.Assert().NotNil(env)
		}))
	})
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
