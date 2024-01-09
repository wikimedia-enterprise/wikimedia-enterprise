// Package confirmuser provides sign up confirmation.
// Uses confirmation code received by email.
package confirmuser

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

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

// Model structure represents input data format.
type Model struct {
	Username         string `json:"username" form:"username" binding:"required,min=1,max=255"`
	ConfirmationCode string `json:"confirmation_code" form:"confirmation_code" binding:"required,min=1,max=6"`
}

// Response structure represents response data format.
type Response struct {
	UserId string `json:"userId"`
	Email  string `json:"email"`
}

// NewHandler creates a new gin handler function for confirm user endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := gcx.ShouldBind(mdl); err != nil {
			log.Error(err, log.Tip("problem binding request input to confirmation code struct"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		csgi := new(cognitoidentityprovider.ConfirmSignUpInput)
		csgi.SetUsername(mdl.Username)
		csgi.SetClientId(p.Env.CognitoClientID)
		csgi.SetConfirmationCode(mdl.ConfirmationCode)

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))

		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", mdl.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("problem writing hashed user and confirmation code"))
			httputil.InternalServerError(gcx, err)
			return
		}

		csgi.SetSecretHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))

		if _, err := p.Cognito.ConfirmSignUpWithContext(gcx.Request.Context(), csgi); err != nil {
			log.Error(err, log.Tip("problem confirming sign up"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gui := new(cognitoidentityprovider.AdminGetUserInput)
		gui.SetUsername(mdl.Username)
		gui.SetUserPoolId(p.Env.CognitoUserPoolID)
		usr, err := p.Cognito.AdminGetUserWithContext(gcx.Request.Context(), gui)

		if err != nil {
			log.Error(err, log.Tip("problem with cognito admin get user in sign up"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		rsp := new(Response)

		for _, att := range usr.UserAttributes {
			if *att.Name == "sub" {
				rsp.UserId = *att.Value
			}
			if *att.Name == "email" {
				rsp.Email = *att.Value
			}
		}

		gcx.JSON(http.StatusOK, rsp)
	}
}
