package v2_test

import (
	"os"
	"testing"
	v2 "wikimedia-enterprise/api/main/handlers/v2"
	"wikimedia-enterprise/api/main/packages/container"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type v2TestSuite struct {
	suite.Suite
	cnt *dig.Container
	rtr *gin.Engine
}

func (s *v2TestSuite) SetupTest() {
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
	os.Setenv("FREE_TIER_GROUP", "group_1")

	gin.SetMode(gin.TestMode)
	s.rtr = gin.New()

	var err error
	s.cnt, err = container.New()
	s.Assert().NoError(err)
}

func (s *v2TestSuite) TestNewGroup() {
	grp, err := v2.NewGroup(s.cnt, s.rtr)
	s.Assert().NoError(err)
	s.Assert().NotNil(grp)
}

func TestV2(t *testing.T) {
	suite.Run(t, new(v2TestSuite))
}
