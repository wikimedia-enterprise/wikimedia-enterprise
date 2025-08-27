// Package forgotpassword creates an HTTP handler for forgot passord endpoint.
package forgotpassword

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/libraries/utils"
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type Parameters struct {
	dig.In
	Cognito cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Env     *env.Environment
}
type UserGroupGetter interface {
	GetUserGroup(username string) string
}
type UsernameByEmailGetter interface {
	GetUsernameByEmail(email string) (string, error)
}

type UserInfoService interface {
	UsernameByEmailGetter
	UserGroupGetter
}

// Request structure represents the endpoint input data.
type Request struct {
	Username string `json:"username" form:"username" binding:"required,min=1,max=255"`
}

var (
	// Here, "internal" means the error is on our side (Wikimedia Enterprise), not necessarily in the auth API server.
	internalErr = errors.New("Internal error, please try again later.")
)

// NewHandler creates a new gin handler function for forgot password endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input to v1 username struct"), log.Any("url", gcx.Request.URL.String()))
			httputil.UnprocessableEntity(gcx, internalErr)
			return
		}

		username := req.Username
		utl := utils.NewParameters(p.Cognito, p.Env)
		// User can input their email address instead of username
		if strings.Contains(req.Username, "@") {
			var eml = username
			var err error
			username, err = utl.GetUsernameByEmail(eml)
			if err != nil {
				log.Warn(err, log.Tip("failed to find username by email"))
				httputil.Unauthorized(gcx, errors.New("User not found."))
				return
			}
		}

		grp, err := utl.GetUserGroup(username)
		if err != nil {
			log.Error(err, log.Tip("failed to find user group"), log.Any("username", username))
			httputil.Unauthorized(gcx, internalErr)
			return
		}

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))
		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("failed to write output for v1 username and cognito client id"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		_, err = p.Cognito.ForgotPasswordWithContext(gcx.Request.Context(), &cognitoidentityprovider.ForgotPasswordInput{
			ClientId:   aws.String(p.Env.CognitoClientID),
			SecretHash: aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
			Username:   aws.String(username),
			ClientMetadata: map[string]*string{
				"group":    aws.String(grp),
				"username": aws.String(username),
			},
		})

		if err != nil {
			log.Error(err, log.Tip("failed to request password reset in cognito"))
			httputil.Unauthorized(gcx, internalErr)
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
