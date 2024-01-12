// Package forgotpasswordconfirm creates an HTTP handler for confirm forgot passord endpoint.
package forgotpasswordconfirm

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	Cognito cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Env     *env.Environment
}

// Request structure represents the endpoint input data.
type Request struct {
	Username         string `json:"username" form:"username" binding:"required,min=1,max=255"`
	Password         string `json:"password" form:"password" binding:"required,min=1,max=255"`
	ConfirmationCode string `json:"confirmation_code" form:"confirmation_code" binding:"required,min=1,max=2048"`
}

// NewHandler creates a new gin handler function for confirm forgot password endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input to v1 change password"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))
		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", req.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("problem writing username and cognito client id to v1 change password"))
			httputil.InternalServerError(gcx, err)
			return
		}

		_, err := p.Cognito.ConfirmForgotPasswordWithContext(gcx.Request.Context(), &cognitoidentityprovider.ConfirmForgotPasswordInput{
			ClientId:         aws.String(p.Env.CognitoClientID),
			SecretHash:       aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
			ConfirmationCode: aws.String(req.ConfirmationCode),
			Username:         aws.String(req.Username),
			Password:         aws.String(req.Password),
		})

		if err != nil {
			log.Error(err, log.Tip("problem confirming cognito forgotten password to v1 change password"))
			httputil.Unauthorized(gcx, err)
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
