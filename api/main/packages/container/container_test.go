package container_test

import (
	"os"
	"testing"

	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) SetupSuite() {
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("AWS_ID", "foo")
	os.Setenv("AWS_KEY", "bar")
	os.Setenv("AWS_URL", "url")
	os.Setenv("AWS_BUCKET", "bucket")
	os.Setenv("REDIS_ADDR", "some-addr")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("COGNITO_CLIENT_ID", "ab21ad1")
	os.Setenv("COGNITO_CLIENT_SECRET", "dv34ad21")
	os.Setenv("ACCESS_MODEL", "model")
	os.Setenv("ACCESS_POLICY", "policy")
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
