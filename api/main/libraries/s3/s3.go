// Package s3 provides the constructor of s3 service for dependency injection.
package s3

import (
	"strings"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/submodules/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/aws/aws-sdk-go/service/s3"
)

// New creates S3 storage instance for dependency injection.
func New(env *env.Environment) s3iface.S3API {
	cfg := &aws.Config{
		Region: aws.String(env.AWSRegion),
	}

	if len(env.AWSID) > 0 && len(env.AWSKey) > 0 {
		cfg.Credentials = credentials.NewStaticCredentials(env.AWSID, env.AWSKey, "")
	}

	if len(env.AWSURL) > 0 {
		cfg.Endpoint = aws.String(env.AWSURL)
	}

	if strings.HasPrefix(env.AWSURL, "http://") {
		cfg.DisableSSL = aws.Bool(true)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	if cfg.DisableSSL == aws.Bool(true) || len(env.AWSURL) == 0 || len(env.AWSKey) == 0 {
		log.Warn("not using SSL for s3 connection")
	}

	return s3.New(session.Must(session.NewSession(cfg)))
}
