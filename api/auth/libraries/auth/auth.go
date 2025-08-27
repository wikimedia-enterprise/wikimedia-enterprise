// Package auth provides the constructor of identity provider for dependency injection.
package auth

import (
	"log"
	"strings"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/submodules/httputil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// New creates cognito identity provider with new session.
func New(env *env.Environment) httputil.AuthProvider {
	cfg := &aws.Config{
		Region: aws.String(env.AWSRegion),
	}

	if len(env.CognitoClientID) > 0 && len(env.CognitoSecret) > 0 {
		cfg.Credentials = credentials.NewStaticCredentials(env.CognitoClientID, env.CognitoSecret, "")
	} else {
		log.Println("Auth: No credentials configured")
	}

	if len(env.AWSURL) > 0 {
		cfg.Endpoint = aws.String(env.AWSURL)

		if strings.HasPrefix(env.AWSURL, "http://") {
			cfg.DisableSSL = aws.Bool(true)
		}
		log.Println("Auth: Using endpoint", env.AWSURL)
	} else {
		log.Println("Auth: No endpoint configured")
	}

	ses := session.Must(session.NewSession(cfg))

	return &httputil.CognitoAuthProvider{
		CIP: cognitoidentityprovider.New(ses),
	}
}
