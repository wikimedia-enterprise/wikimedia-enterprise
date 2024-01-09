// Package forgotpassword creates an HTTP handler for forgot passord endpoint.
package forgotpassword

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
	Username string `json:"username" form:"username" binding:"required,min=1,max=255"`
}

// NewHandler creates a new gin handler function for forgot password endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input to v1 username struct"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))
		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", req.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("failed to write output for v1 username and cognito client id"))
			httputil.InternalServerError(gcx, err)
			return
		}

		_, err := p.Cognito.ForgotPasswordWithContext(gcx.Request.Context(), &cognitoidentityprovider.ForgotPasswordInput{
			ClientId:   aws.String(p.Env.CognitoClientID),
			SecretHash: aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
			Username:   aws.String(req.Username),
		})

		if err != nil {
			log.Error(err, log.Tip("failed to forget password in cognito"))
			httputil.Unauthorized(gcx, err)
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
