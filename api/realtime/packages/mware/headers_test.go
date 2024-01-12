package mware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"wikimedia-enterprise/api/realtime/packages/mware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

const streamTestURL = "/stream"

func createStreamServer() http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(mware.Headers())
	router.Handle(http.MethodGet, streamTestURL, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

type headerTestSuite struct {
	suite.Suite
	srv *httptest.Server
}

// SetupTest creates an instance of httptest.Server with test URL "/stream" and method Get.
func (s *headerTestSuite) SetupTest() {
	s.srv = httptest.NewServer(createStreamServer())
}

// TestHearders sends a HTTP GET request to the test server on the test URL, and verifies the status code and response headers.
func (s *headerTestSuite) TestHeaders() {
	defer s.srv.Close()
	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, streamTestURL))
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusOK, res.StatusCode)
	s.Assert().Equal("no-cache", res.Header.Get("Cache-Control"))
	s.Assert().Equal("keep-alive", res.Header.Get("Connection"))
}

func TestHearders(t *testing.T) {
	suite.Run(t, new(headerTestSuite))
}
