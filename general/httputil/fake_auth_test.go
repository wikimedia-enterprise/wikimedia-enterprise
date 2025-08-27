package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type fakeAuthTestSuite struct {
	suite.Suite
}

func (s *fakeAuthTestSuite) TestFakeAuth() {
	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	called := false
	router := gin.New()
	router.Use(FakeAuth("group_2"))
	router.GET("/", func(gcx *gin.Context) {
		called = true

		val, exists := gcx.Get("user")
		user := val.(*User)
		s.Assert().True(exists)
		s.Assert().Equal("fakeuser", user.Username)
		s.Assert().Equal([]string{"group_2"}, user.Groups)
	})

	router.ServeHTTP(w, req)

	s.Assert().True(called)
}

func TestFakeAuth(t *testing.T) {
	suite.Run(t, new(fakeAuthTestSuite))
}
