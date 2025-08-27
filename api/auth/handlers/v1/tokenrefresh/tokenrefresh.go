// Package tokenrefresh provides a handler to refresh an access token.
package tokenrefresh

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	Cognito cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Env     *env.Environment
	Redis   redis.Cmdable
}

// Request structure represents the endpoint input data.
type Request struct {
	Username     string `json:"username" form:"username" binding:"required,min=1,max=255"`
	RefreshToken string `json:"refresh_token" form:"refresh_token" binding:"required"`
}

// Response structure represents response data format.
type Response struct {
	IdToken     string `json:"id_token,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

var (
	// Here, "internal" means the error is on our side (Wikimedia Enterprise), not necessarily in the auth API server.
	internalErr = errors.New("Internal error, please try again later.")
)

// NewHandler makes a access token refresh.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		req := new(Request)

		if err := gcx.ShouldBind(req); err != nil {
			log.Error(err, log.Tip("problem binding request input in token refresh model v1"), log.Any("url", gcx.Request.URL.String()))
			httputil.UnprocessableEntity(gcx, internalErr)
			return
		}

		// Check if the user can generate more access tokens using the same refresh token
		key := fmt.Sprintf("refresh_token:%s:%s:access_tokens", req.Username, req.RefreshToken)
		card, err := p.Redis.SCard(gcx.Request.Context(), key).Result()

		if err != nil {
			log.Error(err, log.Tip("problem in token refresh with redis read scard v1"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		if card >= p.Env.MaxAccessTokens {
			httputil.ToManyRequests(gcx, errors.New("Refresh tokens limit has been exceeded."))
			return
		}

		// A keyed-hash message authentication code (HMAC) calculated using
		// the secret key of a user pool client and username plus the client
		// ID in the message.
		hmac := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))

		if _, err := hmac.Write([]byte(fmt.Sprintf("%s%s", req.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("problem in token refresh with writing user and cognito client id v1"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		out, err := p.Cognito.InitiateAuthWithContext(gcx.Request.Context(), &cognitoidentityprovider.InitiateAuthInput{
			ClientId: aws.String(p.Env.CognitoClientID),
			AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
			AuthParameters: map[string]*string{
				"REFRESH_TOKEN": aws.String(req.RefreshToken),
				"SECRET_HASH":   aws.String(base64.StdEncoding.EncodeToString(hmac.Sum(nil))),
			},
		})

		if err != nil {
			log.Error(err, log.Tip("problem in token refresh user unauthorized v1"))
			httputil.Unauthorized(gcx, errors.New("Unauthorized."))
			return
		}

		if out.AuthenticationResult == nil {
			log.Error(err, log.Tip("problem in token refresh user auth result in nil v1"))
			httputil.InternalServerError(gcx)
			return
		}

		if err := p.Redis.SAdd(gcx.Request.Context(), key, *out.AuthenticationResult.AccessToken).Err(); err != nil {
			log.Error(err, log.Tip("problem in token refresh adding redis token v1"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		if card == 0 {
			if err := p.Redis.Expire(gcx.Request.Context(), key, time.Second*60*60*time.Duration(p.Env.AccessTokensExpHours)).Err(); err != nil {
				httputil.InternalServerError(gcx, internalErr)
				return
			}
		}

		rsp := new(Response)
		rsp.IdToken = *out.AuthenticationResult.IdToken
		rsp.AccessToken = *out.AuthenticationResult.AccessToken
		rsp.ExpiresIn = int(*out.AuthenticationResult.ExpiresIn)

		gcx.JSON(http.StatusOK, rsp)
	}
}
