package captcha_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"wikimedia-enterprise/api/auth/handlers/v1/captcha"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/suite"
)

const (
	testGetUrl  = "/captcha"
	testShowUrl = "/captcha/:identifier"
)

type catpchaGetHandlerTestSuite struct {
	suite.Suite
	srv  *httptest.Server
	db   *redis.Client
	mock redismock.ClientMock
}

func (c *catpchaGetHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (c *catpchaGetHandlerTestSuite) SetupTest() {
	c.db, c.mock = redismock.NewClientMock()
	c.srv = httptest.NewServer(c.createServer(&captcha.Parameters{
		Redis: c.db,
	}))
}

func (c *catpchaGetHandlerTestSuite) createServer(p *captcha.Parameters) http.Handler {
	router := gin.New()
	router.GET(testGetUrl, captcha.NewGetHandler(p))
	return router
}

func (c *catpchaGetHandlerTestSuite) TearDownTest() {
	c.srv.Close()
}

func (c *catpchaGetHandlerTestSuite) TestNewGetHandler() {
	c.mock.Regexp().ExpectSet(`.`, `.`, 10*time.Minute).SetVal("anything")

	resp, err := http.Get(fmt.Sprintf("%s%s", c.srv.URL, testGetUrl))
	c.Assert().NoError(err)
	c.Assert().Equal(resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()

	target := new(captcha.Response)
	c.Assert().NoError(json.NewDecoder(resp.Body).Decode(target))
	c.Assert().NotZero(target.Identifier)
}

func (c *catpchaGetHandlerTestSuite) TestNewGetHandlerErrors() {
	c.mock.Regexp().ExpectSet(`.`, `.`, 10*time.Minute).SetErr(errors.New("FAIL"))

	resp, err := http.Get(fmt.Sprintf("%s%s", c.srv.URL, testGetUrl))
	c.Assert().NoError(err)
	defer resp.Body.Close()
	c.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

type catpchaShowHandlerTestSuite struct {
	suite.Suite
	srv  *httptest.Server
	db   *redis.Client
	mock redismock.ClientMock
}

func (c *catpchaShowHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (c *catpchaShowHandlerTestSuite) SetupTest() {
	c.db, c.mock = redismock.NewClientMock()
	c.srv = httptest.NewServer(c.createServer(&captcha.Parameters{
		Redis: c.db,
	}))
}

func (c *catpchaShowHandlerTestSuite) createServer(p *captcha.Parameters) http.Handler {
	router := gin.New()
	router.GET(testShowUrl, captcha.NewShowHandler(p))
	return router
}

func (c *catpchaShowHandlerTestSuite) TearDownTest() {
	c.srv.Close()
}

func (c *catpchaShowHandlerTestSuite) TestNewShowHandler() {
	c.mock.Regexp().ExpectGet(`.`).SetVal("654321")

	resp, err := http.Get(fmt.Sprintf("%s%s", c.srv.URL, testShowUrl))
	c.Assert().NoError(err)
	defer resp.Body.Close()

	c.Assert().Equal(http.StatusOK, resp.StatusCode)
}

func (c *catpchaShowHandlerTestSuite) TestNewShowHandlerGetError() {
	c.mock.Regexp().ExpectGet(`.`).SetErr(errors.New("FAIL"))

	resp, err := http.Get(fmt.Sprintf("%s%s", c.srv.URL, testShowUrl))
	c.Assert().NoError(err)
	defer resp.Body.Close()

	c.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

func (c *catpchaShowHandlerTestSuite) TestNewShowHandlerConvertError() {
	c.mock.Regexp().ExpectGet(`.`).SetVal("654a21")

	resp, err := http.Get(fmt.Sprintf("%s%s", c.srv.URL, testShowUrl))
	c.Assert().NoError(err)
	defer resp.Body.Close()

	c.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

func TestCaptcha(t *testing.T) {
	suite.Run(t, new(catpchaGetHandlerTestSuite))
	suite.Run(t, new(catpchaShowHandlerTestSuite))
}
