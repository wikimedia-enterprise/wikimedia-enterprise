package v2_test

import (
	"context"
	"os"
	"testing"
	v2 "wikimedia-enterprise/api/realtime/handlers/v2"
	"wikimedia-enterprise/api/realtime/packages/container"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type v2TestSuite struct {
	suite.Suite
	cont   *dig.Container
	router *gin.Engine
}

func (s *v2TestSuite) SetupTest() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:2000")
	os.Setenv("KAFKA_CONSUMER_GROUP_ID", "unique")
	os.Setenv("COGNITO_CLIENT_ID", "cg_id")
	os.Setenv("COGNITO_CLIENT_SECRET", "cg_sc")
	os.Setenv("REDIS_ADDR", "addr")
	os.Setenv("REDIS_PASSWORD", "pwd")
	os.Setenv("SCHEMA_REGISTRY_URL", "localhost:22")
	os.Setenv("ACCESS_MODEL", "")
	os.Setenv("ACCESS_POLICY", "")

	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	var err error
	s.cont, err = container.New()
	s.Assert().NoError(err)
}

func (s *v2TestSuite) TestNewGroup() {
	group, err := v2.NewGroup(context.Background(), s.cont, s.router)
	s.Assert().NoError(err)
	s.Assert().NotNil(group)
	s.Assert().NotZero(len(group.Handlers))
}

func TestV1(t *testing.T) {
	suite.Run(t, new(v2TestSuite))
}
