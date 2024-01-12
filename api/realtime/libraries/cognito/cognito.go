package cognito

import (
	"wikimedia-enterprise/api/realtime/config/env"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

// New creates new cognito identity provider with session.
func New(env *env.Environment) cognitoidentityprovideriface.CognitoIdentityProviderAPI {
	return cognitoidentityprovider.New(session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(env.AWSRegion),
		Credentials: credentials.NewStaticCredentials(env.AWSID, env.AWSKey, ""),
	})))
}
