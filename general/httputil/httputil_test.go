package httputil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type testModel struct {
	Field string `json:"field"`
}

type utilTestSuite struct {
	suite.Suite
	cnt string
	err error
}

func (s *utilTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.cnt = "test content"
	s.err = errors.New("test error")
}

func (s *utilTestSuite) TestBindModelSuccess() {
	mdl := new(testModel)
	rtr := gin.New()

	rtr.POST("/", func(ctx *gin.Context) {
		err := BindModel(ctx, mdl)
		s.Assert().NoError(err)
	})

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	_, err := http.Post(fmt.Sprintf("%s%s", srv.URL, "/"), "application/json", bytes.NewReader([]byte(`{"field": "value"}`)))
	s.Assert().NoError(err)
}

func (s *utilTestSuite) TestBindModelError() {
	mdl := new(testModel)
	rtr := gin.New()

	rtr.POST("/", func(ctx *gin.Context) {
		err := BindModel(ctx, mdl)
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "Content-Type")
	})

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	_, err := http.Post(fmt.Sprintf("%s%s", srv.URL, "/"), "multipart/form-data", bytes.NewReader([]byte("")))
	s.Assert().NoError(err)
}

func (s *utilTestSuite) TestDelimiterSuccess() {
	rtr := gin.New()

	rtr.POST("/", func(ctx *gin.Context) {
		dtr := RecordDelimiter(ctx)
		s.Assert().Equal(",", dtr)
	})

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	_, err := http.Post(fmt.Sprintf("%s%s", srv.URL, "/"), "application/json", strings.NewReader(""))
	s.Assert().NoError(err)
}

func (s *utilTestSuite) TestRenderSuccess() {
	rtr := gin.New()

	rtr.GET("/", func(ctx *gin.Context) {
		err := Render(ctx, http.StatusOK, s.cnt)

		s.Assert().NoError(err)
	})

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, "/"))
	s.Assert().NoError(err)
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(s.cnt, string(data))
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *utilTestSuite) TestRenderErrorSuccess() {
	rtr := gin.New()

	rtr.GET("/", func(ctx *gin.Context) {
		RenderError(ctx, http.StatusInternalServerError, s.err)
	})

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, "/"))
	s.Assert().NoError(err)
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
	s.Assert().Contains(string(data), s.err.Error())
}

func (s *utilTestSuite) TestNoRoute() {
	rtr := gin.New()
	rtr.Use(NoRoute())

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s/not-found", srv.URL))
	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
	s.Assert().Contains(string(dta), http.StatusText(http.StatusNotFound))
}

func (s *utilTestSuite) TestCORS() {
	rtr := gin.New()
	rtr.Use(CORS())

	srv := httptest.NewServer(rtr)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodHead, fmt.Sprintf("%s/test", srv.URL), nil)
	s.Assert().NoError(err)
	req.Header.Set("Origin", "http://localhost:1010")

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(res.Header.Get("Access-Control-Allow-Origin"), "*")
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(utilTestSuite))
}
