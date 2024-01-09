package httputil

import (
	"errors"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// ErrNoUser indicates there is no user (or the user is nil) in the request's context.
var ErrNoUser = errors.New("missing user in request context")

// CasbinRBACAuthorizer uses a provided Casbin enforcer to implement RBAC middleware.
// This function will look up for a `User` instance stored in the request's
// `gin.Context` using the `user` key, and will attempt to authorize the request
// using each one of the user's roles.
// If no match is made, the request will be rejected.
func CasbinRBACAuthorizer(enf *casbin.Enforcer) RBACAuthorizeFunc {
	return func(gcx *gin.Context) (bool, error) {
		usr, ok := gcx.Get("user")

		if usr == nil || !ok {
			return false, ErrNoUser
		}

		user, _ := usr.(*User)

		for _, role := range user.GetGroups() {
			res, err := enf.Enforce(role, gcx.Request.URL.Path, gcx.Request.Method)

			if err != nil {
				return false, err
			}

			if res {
				return true, nil
			}
		}

		return false, nil
	}
}
