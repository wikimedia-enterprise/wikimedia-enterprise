package v1_test

import (
	"os"
	"testing"
	v1 "wikimedia-enterprise/api/auth/handlers/v1"
	"wikimedia-enterprise/api/auth/packages/container"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type v1TestSuite struct {
	suite.Suite
	cont   *dig.Container
	router *gin.Engine
}

func (s *v1TestSuite) SetupTest() {
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("AWS_ID", "foo")
	os.Setenv("AWS_KEY", "bar")
	os.Setenv("COGNITO_CLIENT_ID", "ab21ad1")
	os.Setenv("COGNITO_SECRET", "dv34ad21")
	os.Setenv("COGNITO_USER_POOL_ID", "ffd2as1w")
	os.Setenv("COGNITO_USER_GROUP", "free")
	os.Setenv("ACCESS_POLICY", "")
	os.Setenv("GROUP_DOWNLOAD_LIMIT", "1000")

	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	var err error
	s.cont, err = container.New()
	s.Assert().NoError(err)
}

func (s *v1TestSuite) TestNewGroup() {
	group, err := v1.NewGroup(s.cont, s.router)
	s.Assert().NoError(err)
	s.Assert().NotNil(group)
}

func TestV1(t *testing.T) {
	suite.Run(t, new(v1TestSuite))
}
