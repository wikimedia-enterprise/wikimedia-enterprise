package enforcer_test

import (
	"os"
	"testing"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/libraries/enforcer"

	"github.com/stretchr/testify/suite"
)

type enforcerTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *enforcerTestSuite) SetupSuite() {
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("AWS_ID", "foo")
	os.Setenv("AWS_KEY", "bar")
	os.Setenv("AWS_URL", "url")
	os.Setenv("AWS_BUCKET", "bucket")
	os.Setenv("REDIS_ADDR", "some-addr")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("COGNITO_CLIENT_ID", "ab21ad1")
	os.Setenv("COGNITO_CLIENT_SECRET", "dv34ad21")
	os.Setenv("ACCESS_MODEL", "")
	os.Setenv("ACCESS_POLICY", "")
	s.env, _ = env.New()
}

func (s *enforcerTestSuite) TestEnforcerSuccess() {
	enf, err := enforcer.New(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(enf)
}

func TestEnforcer(t *testing.T) {
	suite.Run(t, new(enforcerTestSuite))
}
