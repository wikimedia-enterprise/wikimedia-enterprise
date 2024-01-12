// Package changepassword creates an HTTP handler for change passord endpoint.
package changepassword

import (
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
	AccessToken      string `json:"access_token" form:"access_token" binding:"required"`
	PreviousPassword string `json:"previous_password" form:"previous_password" binding:"required,min=6,max=256"`
	ProposedPassword string `json:"proposed_password" form:"proposed_password" binding:"required,min=6,max=256"`
}

// NewHandler creates a new handler function for change password endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input to change password struct"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		_, err := p.Cognito.ChangePasswordWithContext(gcx.Request.Context(), &cognitoidentityprovider.ChangePasswordInput{
			AccessToken:      aws.String(req.AccessToken),
			PreviousPassword: aws.String(req.PreviousPassword),
			ProposedPassword: aws.String(req.ProposedPassword),
		})

		if err != nil {
			log.Error(err, log.Tip("problem change password was unauthorized"))
			httputil.Unauthorized(gcx, err)
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
