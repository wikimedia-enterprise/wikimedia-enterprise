package status_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"wikimedia-enterprise/api/main/handlers/v2/status"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type statusTestSuite struct {
	suite.Suite
	srv *httptest.Server
	url string
}

func (s *statusTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.url = "/status"
}

func (s *statusTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)

	rtr := gin.New()
	rtr.GET(s.url, status.NewHandler())

	return rtr
}

func (s *statusTestSuite) SetupTest() {
	s.srv = httptest.NewServer(s.createServer())
}

func (s *statusTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *statusTestSuite) TestHandler() {
	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, s.url))
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(statusTestSuite))
}
