package httputil

import "github.com/gin-gonic/gin"

// RBACAuthorizeFunc is the type alias for a RBAC Authorize function signature.
type RBACAuthorizeFunc func(*gin.Context) (bool, error)

// RBAC implements RBAC using the provided authorizer function.
func RBAC(fn RBACAuthorizeFunc) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		ok, err := fn(gcx)

		if err != nil {
			AbortWithInternalServerError(gcx, err)
			return
		}

		if !ok {
			AbortWithForbidden(gcx)
			return
		}

		gcx.Next()
	}
}
