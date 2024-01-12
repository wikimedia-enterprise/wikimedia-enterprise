package httputil

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

var testErr = "test error"

type testRBACSuccessSuite struct {
	suite.Suite
	clb RBACAuthorizeFunc
	url string
	srv *httptest.Server
}

func createRBACServer(url string, clb RBACAuthorizeFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()
	rtr.Use(RBAC(clb))
	rtr.GET(url, func(gcx *gin.Context) { gcx.Status(http.StatusTeapot) })

	return rtr
}

func (s *testRBACSuccessSuite) SetupTest() {
	s.url = "/test"
	s.srv = httptest.NewServer(createRBACServer(s.url, s.clb))
}

func (s *testRBACSuccessSuite) TearDownTest() {
	s.srv.Close()
}

func (s *testRBACSuccessSuite) TestSuccess() {
	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, s.url))
	s.Assert().NoError(err)

	s.Assert().Equal(http.StatusTeapot, res.StatusCode)
}

func TestRBACSuccess(t *testing.T) {
	for _, testCase := range []*testRBACSuccessSuite{
		{
			clb: func(*gin.Context) (bool, error) { return true, nil },
		},
	} {
		suite.Run(t, testCase)
	}
}

type testRBACErrorSuite struct {
	suite.Suite
	clb RBACAuthorizeFunc
	url string
	srv *httptest.Server
}

func (s *testRBACErrorSuite) SetupTest() {
	s.url = "/test"
	s.srv = httptest.NewServer(createRBACServer(s.url, s.clb))
}

func (s *testRBACErrorSuite) TearDownTest() {
	s.srv.Close()
}

func (s *testRBACErrorSuite) TestRBACError() {
	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, s.url))
	s.Assert().NoError(err)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func TestRBACAccessDenied(t *testing.T) {
	for _, testCase := range []*testRBACErrorSuite{
		{
			clb: func(*gin.Context) (bool, error) { return false, nil },
		},
	} {
		suite.Run(t, testCase)
	}
}

type testRBACWithErrorSuite struct {
	suite.Suite
	clb RBACAuthorizeFunc
	url string
	srv *httptest.Server
}

func (s *testRBACWithErrorSuite) SetupTest() {
	s.url = "/test"
	s.srv = httptest.NewServer(createRBACServer(s.url, s.clb))
}

func (s *testRBACWithErrorSuite) TearDownTest() {
	s.srv.Close()
}

func (s *testRBACWithErrorSuite) TestRBACAccessDeniedWIthError() {
	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, s.url))
	s.Assert().NoError(err)

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), testErr)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func TestRBACAccessDeniedWithError(t *testing.T) {
	for _, testCase := range []*testRBACWithErrorSuite{
		{
			clb: func(*gin.Context) (bool, error) { return true, errors.New(testErr) },
		},
	} {
		suite.Run(t, testCase)
	}
}
