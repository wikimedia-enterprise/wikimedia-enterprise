package v1

import (
	"net/http"
	"wikimedia-enterprise/api/auth/handlers/v1/captcha"
	"wikimedia-enterprise/api/auth/handlers/v1/changepassword"
	"wikimedia-enterprise/api/auth/handlers/v1/confirmuser"
	"wikimedia-enterprise/api/auth/handlers/v1/createuser"
	"wikimedia-enterprise/api/auth/handlers/v1/forgotpassword"
	"wikimedia-enterprise/api/auth/handlers/v1/forgotpasswordconfirm"
	"wikimedia-enterprise/api/auth/handlers/v1/getuser"
	"wikimedia-enterprise/api/auth/handlers/v1/login"
	"wikimedia-enterprise/api/auth/handlers/v1/newpasswordrequired"
	"wikimedia-enterprise/api/auth/handlers/v1/resendconfirm"
	"wikimedia-enterprise/api/auth/handlers/v1/tokenrefresh"
	"wikimedia-enterprise/api/auth/handlers/v1/tokenrevoke"
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// NewGroup creates new router group for the API.
func NewGroup(cnt *dig.Container, rtr *gin.Engine) (*gin.RouterGroup, error) {
	v1 := rtr.Group("/v1")
	v1.GET("/status", func(gcx *gin.Context) { gcx.Status(http.StatusOK) })

	for _, err := range []error{
		cnt.Invoke(func(p login.Parameters) {
			v1.POST("/login", login.NewHandler(&p))
		}),
		cnt.Invoke(func(p tokenrefresh.Parameters) {
			v1.POST("/token-refresh", tokenrefresh.NewHandler(&p))
		}),
		cnt.Invoke(func(p tokenrevoke.Parameters) {
			v1.POST("/token-revoke", tokenrevoke.NewHandler(&p))
		}),
		cnt.Invoke(func(p forgotpassword.Parameters) {
			v1.POST("/forgot-password", forgotpassword.NewHandler(&p))
		}),
		cnt.Invoke(func(p forgotpasswordconfirm.Parameters) {
			v1.POST("/forgot-password-confirm", forgotpasswordconfirm.NewHandler(&p))
		}),
		cnt.Invoke(func(p changepassword.Parameters) {
			v1.POST("/change-password", changepassword.NewHandler(&p))
		}),
		cnt.Invoke(func(p newpasswordrequired.Parameters) {
			v1.POST("/new-password-required", newpasswordrequired.NewHandler(&p))
		}),
		cnt.Invoke(func(p createuser.Parameters) {
			v1.POST("/create-user", createuser.NewHandler(&p))
		}),
		cnt.Invoke(func(p confirmuser.Parameters) {
			v1.POST("/confirm-user", confirmuser.NewHandler(&p))
		}),
		cnt.Invoke(func(p resendconfirm.Parameters) {
			v1.POST("/resend-confirm", resendconfirm.NewHandler(&p))
		}),
		cnt.Invoke(func(p captcha.Parameters) {
			v1.GET("/captcha", captcha.NewGetHandler(&p))
			v1.GET("/captcha/:identifier", captcha.NewShowHandler(&p))
		}),
		cnt.Invoke(func(p getuser.Parameters, apv httputil.AuthProvider) {
			amw := httputil.Auth(httputil.NewAuthParams(p.Env.CognitoClientID, p.Redis, apv))
			v1.POST("/get-user", amw, getuser.NewHandler(&p))
		}),
	} {
		if err != nil {
			log.Error(err, log.Tip("problem read environment values in csv - no data "))
			return nil, err
		}
	}

	return v1, nil
}
