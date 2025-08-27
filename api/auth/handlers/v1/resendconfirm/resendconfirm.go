// Package resendconfirm creates HTTP handler for new resend email confirmation code endpoint.
package resendconfirm

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

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

var (
	// Here, "internal" means the error is on our side (Wikimedia Enterprise), not necessarily in the auth API server.
	internalErr = errors.New("Internal error, please try again later.")
)

// NewHandler creates a new gin handler function for resend email confirmation code endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input to email confirmation model v1"), log.Any("url", gcx.Request.URL.String()))
			httputil.UnprocessableEntity(gcx, internalErr)
			return
		}

		utl := utils.NewParameters(p.Cognito, p.Env)
		grp, err := utl.GetUserGroup(req.Username)
		if err != nil {
			log.Error(err, log.Tip("failed to find user group"))
			httputil.Unauthorized(gcx, errors.New("User not found."))
			return
		}

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))
		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", req.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("problem in email confirmation v1 writing user and cognito client id"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		_, err = p.Cognito.ResendConfirmationCodeWithContext(gcx.Request.Context(), &cognitoidentityprovider.ResendConfirmationCodeInput{
			ClientId:   aws.String(p.Env.CognitoClientID),
			SecretHash: aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
			Username:   aws.String(req.Username),
			ClientMetadata: map[string]*string{
				"username": aws.String(req.Username),
				"group":    aws.String(grp),
			},
		})

		if err != nil {
			log.Error(err, log.Tip("problem in new password v1 user unauthorized"))
			httputil.Unauthorized(gcx, errors.New("Unauthorized."))
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
