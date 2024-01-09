package v1_test

import (
	"os"
	"testing"
	v1 "wikimedia-enterprise/api/main/handlers/v1"
	"wikimedia-enterprise/api/main/packages/container"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type v1TestSuite struct {
	suite.Suite
	cnt *dig.Container
	rtr *gin.Engine
}

func (s *v1TestSuite) SetupTest() {
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("AWS_ID", "foo")
	os.Setenv("AWS_KEY", "bar")
	os.Setenv("AWS_BUCKET", "bar")
	os.Setenv("AWS_URL", "bar")
	os.Setenv("REDIS_ADDR", "some-addr")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("COGNITO_CLIENT_ID", "id")
	os.Setenv("COGNITO_CLIENT_SECRET", "secret")
	os.Setenv("ACCESS_MODEL", "id")
	os.Setenv("ACCESS_POLICY", "secret")

	gin.SetMode(gin.TestMode)
	s.rtr = gin.New()

	var err error
	s.cnt, err = container.New()
	s.Assert().NoError(err)
}

func (s *v1TestSuite) TestNewGroup() {
	grp, err := v1.NewGroup(s.cnt, s.rtr)
	s.Assert().NoError(err)
	s.Assert().NotNil(grp)
}

func TestV1(t *testing.T) {
	suite.Run(t, new(v1TestSuite))
}
