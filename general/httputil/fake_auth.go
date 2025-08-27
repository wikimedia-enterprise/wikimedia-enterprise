package httputil

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

var (
	ErrFakeAuth = errors.New("using fake authentication, this is only OK in local development")
)

// Fake Auth middleware for local development
func FakeAuth(fakeGroup string) gin.HandlerFunc {
	log.Print(ErrFakeAuth)

	group := fakeGroup
	if group == "" {
		group = "group_3"
	}
	return func(gcx *gin.Context) {
		gcx.Set("user", &User{Username: "fakeuser", Groups: []string{group}})
		gcx.Next()
	}
}
