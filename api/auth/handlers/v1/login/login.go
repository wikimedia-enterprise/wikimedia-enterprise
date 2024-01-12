// Package creates HTTP handler for login endpoint.
package login

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
	ChallengeName string `json:"challenge_name,omitempty"`
	IDToken       string `json:"id_token,omitempty"`
	AccessToken   string `json:"access_token,omitempty"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	ExpiresIn     int    `json:"expires_in,omitempty"`
	Session       string `json:"session,omitempty"`
}

// Model structure represents input data format.
type Model struct {
	Username string `json:"username" form:"username" binding:"required,min=1,max=255"`
	Password string `json:"password" form:"password" binding:"required,min=1,max=255"`
}

// NewHandler creates new login HTTP handler.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := gcx.ShouldBind(mdl); err != nil {
			log.Error(err, log.Tip("problem binding request input to new login model"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))

		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", mdl.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("problem in login with writing user and cognito client id"))
			httputil.InternalServerError(gcx, err)
			return
		}

		out, err := p.Cognito.InitiateAuthWithContext(gcx.Request.Context(), &cognitoidentityprovider.InitiateAuthInput{
			ClientId: aws.String(p.Env.CognitoClientID),
			AuthFlow: aws.String("USER_PASSWORD_AUTH"),
			AuthParameters: map[string]*string{
				"USERNAME":    aws.String(mdl.Username),
				"PASSWORD":    aws.String(mdl.Password),
				"SECRET_HASH": aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
			},
		})

		if err != nil {
			log.Error(err, log.Tip("problem authentication user during login, unauthorized"))
			httputil.Unauthorized(gcx, err)
			return
		}

		rsp := new(Response)

		if out.AuthenticationResult != nil {
			rsp.IDToken = *out.AuthenticationResult.IdToken
			rsp.AccessToken = *out.AuthenticationResult.AccessToken
			rsp.RefreshToken = *out.AuthenticationResult.RefreshToken
			rsp.ExpiresIn = int(*out.AuthenticationResult.ExpiresIn)
		}

		if out.ChallengeName != nil && len(*out.ChallengeName) > 0 {
			rsp.ChallengeName = *out.ChallengeName
			rsp.Session = *out.Session
		}

		gcx.JSON(http.StatusOK, rsp)
	}
}
