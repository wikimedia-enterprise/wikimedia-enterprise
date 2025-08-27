// Package changepassword creates an HTTP handler for change passord endpoint.
package changepassword

import (
	"errors"
	"net/http"

	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

var (
	// Here, "internal" means the error is on our side (Wikimedia Enterprise), not necessarily in the auth API server.
	internalErr = errors.New("Internal error, please try again later.")
)

// NewHandler creates a new handler function for change password endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input to change password struct"), log.Any("url", gcx.Request.URL.String()))
			httputil.UnprocessableEntity(gcx, internalErr)
			return
		}

		_, err := p.Cognito.ChangePasswordWithContext(gcx.Request.Context(), &cognitoidentityprovider.ChangePasswordInput{
			AccessToken:      aws.String(req.AccessToken),
			PreviousPassword: aws.String(req.PreviousPassword),
			ProposedPassword: aws.String(req.ProposedPassword),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				log.Error("password change AWS error", log.Any("error", err))
				switch awsErr.Code() {
				case cognitoidentityprovider.ErrCodeNotAuthorizedException, cognitoidentityprovider.ErrCodeInvalidPasswordException, cognitoidentityprovider.ErrCodeUserNotFoundException:
					httputil.Unauthorized(gcx, errors.New("Incorrect username or password."))
				default:
					httputil.InternalServerError(gcx, internalErr)
				}
				return
			}

			log.Error("password change unknown error", log.Any("error", err))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
