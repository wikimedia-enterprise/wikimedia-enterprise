package cognito

import (
	"log"
	"strings"
	"wikimedia-enterprise/api/auth/config/env"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

// New creates new cognito identity provider with session.
func New(env *env.Environment) cognitoidentityprovideriface.CognitoIdentityProviderAPI {
	cfg := &aws.Config{
		Region: aws.String(env.AWSRegion),
	}

	if len(env.AWSID) > 0 && len(env.AWSKey) > 0 {
		cfg.Credentials = credentials.NewStaticCredentials(env.AWSID, env.AWSKey, "")
	}

	if len(env.AWSURL) > 0 {
		cfg.Endpoint = aws.String(env.AWSURL)

		if strings.HasPrefix(env.AWSURL, "http://") {
			cfg.DisableSSL = aws.Bool(true)
		}
		log.Println("Cognito: Using endpoint", env.AWSURL)
	} else {
		log.Println("Cognito: No endpoint configured")
	}

	return cognitoidentityprovider.New(session.Must(session.NewSession(cfg)))
}
