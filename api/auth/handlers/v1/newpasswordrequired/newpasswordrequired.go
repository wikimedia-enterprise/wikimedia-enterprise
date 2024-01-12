// Package newpasswordrequired creates HTTP handler for new password required challenge endpoint.
package newpasswordrequired

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

// Response structure represents response data format.
type Response struct {
	IDToken      string `json:"id_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// Model structure represents input data format.
type Model struct {
	Username    string `json:"username" form:"username" binding:"required,min=1,max=255"`
	Session     string `json:"session" form:"session" binding:"required,min=1"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required,min=6,max=255"`
}

// NewHandler creates new password required HTTP handler.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := gcx.ShouldBind(mdl); err != nil {
			log.Error(err, log.Tip("problem binding request input to new password model v1"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))

		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", mdl.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("problem in new password v1 writing user and cognito client id"))
			httputil.InternalServerError(gcx, err)
			return
		}

		out, err := p.Cognito.RespondToAuthChallengeWithContext(gcx.Request.Context(), &cognitoidentityprovider.RespondToAuthChallengeInput{
			Session:       aws.String(mdl.Session),
			ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
			ClientId:      aws.String(p.Env.CognitoClientID),
			ChallengeResponses: map[string]*string{
				"NEW_PASSWORD": aws.String(mdl.NewPassword),
				"USERNAME":     aws.String(mdl.Username),
				"SECRET_HASH":  aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
			},
		})

		if err != nil {
			log.Error(err, log.Tip("problem in new password v1 user unauthorized"))
			httputil.Unauthorized(gcx, err)
			return
		}

		if out.AuthenticationResult == nil {
			log.Error("problem in new password v1 authentication result is nil")
			httputil.InternalServerError(gcx)
			return
		}

		rsp := new(Response)
		rsp.IDToken = *out.AuthenticationResult.IdToken
		rsp.AccessToken = *out.AuthenticationResult.AccessToken
		rsp.RefreshToken = *out.AuthenticationResult.RefreshToken
		rsp.ExpiresIn = int(*out.AuthenticationResult.ExpiresIn)

		gcx.JSON(http.StatusOK, rsp)
	}
}
