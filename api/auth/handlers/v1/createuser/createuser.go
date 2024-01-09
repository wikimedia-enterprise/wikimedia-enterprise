// Package createuser creates new congnito user and adds it to free tier group.
package createuser

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

// Model structure represents input data format.
type Model struct {
	Username         string    `json:"username" form:"username" binding:"required,min=1,max=255"`
	Email            string    `json:"email" form:"email" binding:"required,email,min=1,max=60"`
	Password         string    `json:"password" form:"password" binding:"required,min=8"`
	CaptchaID        string    `json:"captcha_id" form:"captcha_id" binding:"required,min=1,max=30"`
	CaptchaSolution  string    `json:"captcha_solution" form:"captcha_solution" binding:"required,min=1,max=6"`
	PolicyDateAccept time.Time `json:"policy_date_accept" form:"policy_date_accept" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	PolicyVersion    string    `json:"policy_version" form:"policy_version" binding:"required,min=1,max=4"`
	MarketingEmails  string    `json:"marketing_emails" form:"marketing_emails" binding:"required,min=4,max=5"`
}

var pvregex = regexp.MustCompile(`v\d{1}\.\d{1}`)

// NewHandler creates a new gin handler function for create user endpoint.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := gcx.ShouldBind(mdl); err != nil {
			log.Error(err, log.Tip("problem binding request input to create user model"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		if found := pvregex.MatchString(mdl.PolicyVersion); !found {
			log.Error("policy version format is not compatible")
			httputil.UnprocessableEntity(gcx, errors.New("Policy version format is not compatible!"))
			return
		}

		gui := new(cognitoidentityprovider.AdminGetUserInput)
		gui.SetUsername(mdl.Email)
		gui.SetUserPoolId(p.Env.CognitoUserPoolID)

		_, err := p.Cognito.AdminGetUserWithContext(gcx.Request.Context(), gui)

		if err == nil {
			log.Error("user email already exists")
			httputil.BadRequest(gcx, errors.New("An account using this email address already exists."))
			return
		}

		if err != nil {
			aerr, ok := err.(awserr.Error)

			if !ok {
				log.Error(err, log.Tip("could not extract inner error, in create user"))
				httputil.InternalServerError(gcx, err)
				return
			}

			if aerr.Code() != cognitoidentityprovider.ErrCodeUserNotFoundException {
				log.Error(aerr, log.Tip("cognito could not find user"))
				httputil.InternalServerError(gcx, err)
				return
			}
		}

		ggi := new(cognitoidentityprovider.GetGroupInput)
		ggi.SetGroupName(p.Env.CognitoUserGroup)
		ggi.SetUserPoolId(p.Env.CognitoUserPoolID)

		ggo, err := p.Cognito.GetGroupWithContext(gcx.Request.Context(), ggi)

		if err != nil {
			log.Error(err, log.Tip("create user could not find group"))
			httputil.InternalServerError(gcx, err)
			return
		}

		solution := p.Redis.Get(gcx.Request.Context(), mdl.CaptchaID).Val()

		if solution != mdl.CaptchaSolution {
			log.Error("captcha verification failed")
			httputil.BadRequest(gcx, errors.New("Captcha verification failed."))
			return
		}

		// Delete solution after check to prevent abusive use of this endpoint.
		if _, err := p.Redis.Del(gcx.Request.Context(), mdl.CaptchaID).Result(); err != nil {
			log.Error("create user failed to delete redis captcha")
			httputil.InternalServerError(gcx, err)
			return
		}

		sgi := new(cognitoidentityprovider.SignUpInput)
		sgi.SetClientId(p.Env.CognitoClientID)
		sgi.SetUsername(mdl.Username)
		sgi.SetPassword(mdl.Password)
		sgi.SetUserAttributes([]*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(mdl.Email),
			},
			{
				Name:  aws.String("custom:policy_version"),
				Value: aws.String(mdl.PolicyVersion),
			},
			{
				Name:  aws.String("custom:policy_date_accept"),
				Value: aws.String(mdl.PolicyDateAccept.Format(time.RFC3339)),
			},
			{
				Name:  aws.String("custom:marketing_emails"),
				Value: aws.String(mdl.MarketingEmails),
			},
			{
				Name:  aws.String("custom:username"),
				Value: aws.String(mdl.Username),
			},
		})

		h := hmac.New(sha256.New, []byte(p.Env.CognitoSecret))

		if _, err := h.Write([]byte(fmt.Sprintf("%s%s", mdl.Username, p.Env.CognitoClientID))); err != nil {
			log.Error(err, log.Tip("failed to write username and cognito client id"))
			httputil.InternalServerError(gcx, err)
			return
		}

		sgi.SetSecretHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))

		if _, err := p.Cognito.SignUpWithContext(gcx.Request.Context(), sgi); err != nil {
			log.Error(err, log.Tip("failed to sign up with context from cognito"))
			httputil.InternalServerError(gcx, err)
			return
		}

		// Add user to free tier group.
		agi := new(cognitoidentityprovider.AdminAddUserToGroupInput)
		agi.SetGroupName(*ggo.Group.GroupName)
		agi.SetUserPoolId(p.Env.CognitoUserPoolID)
		agi.SetUsername(mdl.Username)

		if _, err := p.Cognito.AdminAddUserToGroupWithContext(gcx.Request.Context(), agi); err != nil {
			log.Error(err, log.Tip("failed to add user to group in cognito"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.Status(http.StatusNoContent)
	}
}
