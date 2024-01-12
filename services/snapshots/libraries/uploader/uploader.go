// Package uploader creates s3manager uploader interface for dependency injection.
package uploader

import (
	"strings"
	"wikimedia-enterprise/services/snapshots/config/env"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Uploader wraps s3manager uploader method in the interface.
type Uploader interface {
	Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// New creates new uploader instance for dependency injection.
func New(env *env.Environment) (Uploader, error) {
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

	ssn, err := session.NewSession(cfg)

	if err != nil {
		return nil, err
	}

	return s3manager.NewUploader(ssn), nil
}
