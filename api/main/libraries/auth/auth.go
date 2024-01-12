// Package auth provides the constructor of identity provider for dependency injection.
package auth

import (
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/general/httputil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// New creates cognito identity provider with new session.
func New(env *env.Environment) httputil.AuthProvider {
	cfg := &aws.Config{
		Region:      aws.String(env.AWSRegion),
		Credentials: credentials.NewStaticCredentials(env.CognitoClientID, env.CognitoClientSecret, ""),
	}

	ses := session.Must(session.NewSession(cfg))

	return &httputil.CognitoAuthProvider{
		CIP: cognitoidentityprovider.New(ses),
	}
}
