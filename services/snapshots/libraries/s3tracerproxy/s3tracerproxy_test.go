package s3tracerproxy

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"wikimedia-enterprise/general/tracing"
)

// MockTracer is a mock implementation of the Tracer interface
type MockTracer struct {
	mock.Mock
}

// Trace is a mock implementation of the Trace method in the Tracer interface
func (m *MockTracer) Trace(ctx context.Context, tags map[string]string) (func(error, string), context.Context) {
	args := m.Called(ctx, tags)
	return args.Get(0).(func(error, string)), args.Get(1).(context.Context)
}

// StorageTraceSuite is a test suite for the StorageTracer
type StorageTraceSuite struct {
	suite.Suite
	tracer       tracing.Tracer
	mockS3API    *MockS3API
	storageTrace S3TracerProxy
}

// MockS3API is a mock implementation of the s3iface.S3API interface
type MockS3API struct {
	s3iface.S3API
	GetObjectWithContextFunc               func(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
	DeleteObjectWithContextFunc            func(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error)
	PutObjectWithContextFunc               func(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error)
	CompleteMultipartUploadWithContextFunc func(ctx context.Context, input *s3.CompleteMultipartUploadInput, opts ...request.Option) (*s3.CompleteMultipartUploadOutput, error)
	UploadPartCopyWithContextFunc          func(ctx context.Context, input *s3.UploadPartCopyInput, opts ...request.Option) (*s3.UploadPartCopyOutput, error)
	CreateMultipartUploadWithContextFunc   func(ctx context.Context, input *s3.CreateMultipartUploadInput, opts ...request.Option) (*s3.CreateMultipartUploadOutput, error)
	CopyObjectWithContextFunc              func(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, error)
	ListObjectsV2PagesWithContextFunc      func(ctx context.Context, input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool, opts ...request.Option) error
}

// GetObjectWithContext is a mock implementation of the GetObjectWithContext method in the s3iface.S3API interface
func (m *MockS3API) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return m.GetObjectWithContextFunc(ctx, input, opts...)
}

// DeleteObjectWithContext is a mock implementation of the DeleteObjectWithContext method in the s3iface.S3API interface
func (m *MockS3API) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	return m.DeleteObjectWithContextFunc(ctx, input, opts...)
}

// PutObjectWithContext is a mock implementation of the PutObjectWithContext method in the s3iface.S3API interface
func (m *MockS3API) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	return m.PutObjectWithContextFunc(ctx, input, opts...)
}

// CompleteMultipartUploadWithContext is a mock implementation of the CompleteMultipartUploadWithContext method in the s3iface.S3API interface
func (m *MockS3API) CompleteMultipartUploadWithContext(ctx context.Context, input *s3.CompleteMultipartUploadInput, opts ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
	return m.CompleteMultipartUploadWithContextFunc(ctx, input, opts...)
}

// UploadPartCopyWithContext is a mock implementation of the UploadPartCopyWithContext method in the s3iface.S3API interface
func (m *MockS3API) UploadPartCopyWithContext(ctx context.Context, input *s3.UploadPartCopyInput, opts ...request.Option) (*s3.UploadPartCopyOutput, error) {
	return m.UploadPartCopyWithContextFunc(ctx, input, opts...)
}

// CreateMultipartUploadWithContext is a mock implementation of the CreateMultipartUploadWithContext method in the s3iface.S3API interface
func (m *MockS3API) CreateMultipartUploadWithContext(ctx context.Context, input *s3.CreateMultipartUploadInput, opts ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
	return m.CreateMultipartUploadWithContextFunc(ctx, input, opts...)
}

// CopyObjectWithContext is a mock implementation of the CopyObjectWithContext method in the s3iface.S3API interface
func (m *MockS3API) CopyObjectWithContext(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, error) {
	return m.CopyObjectWithContextFunc(ctx, input, opts...)
}

// ListObjectsV2PagesWithContext is a mock implementation of the ListObjectsV2PagesWithContext method in the s3iface.S3API interface
func (m *MockS3API) ListObjectsV2PagesWithContext(ctx context.Context, input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool, opts ...request.Option) error {
	return m.ListObjectsV2PagesWithContextFunc(ctx, input, fn, opts...)
}

// SetupTest is a setup function for the StorageTraceSuite
func (suite *StorageTraceSuite) SetupTest() {
	tracer, err := tracing.NewAPI(tracing.WithServiceName("test-service"), tracing.WithSamplingRate(1.0), tracing.WithGRPCHost("localhost:4317"))
	if err != nil {
		suite.FailNow("Failed to initialize tracer: " + err.Error())
	}

	suite.tracer = tracer
	suite.mockS3API = &MockS3API{
		GetObjectWithContextFunc: func(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{}, nil
		},
		DeleteObjectWithContextFunc: func(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
			return &s3.DeleteObjectOutput{}, nil
		},
		PutObjectWithContextFunc: func(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
		CompleteMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.CompleteMultipartUploadInput, opts ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
			return &s3.CompleteMultipartUploadOutput{}, nil
		},
		UploadPartCopyWithContextFunc: func(ctx context.Context, input *s3.UploadPartCopyInput, opts ...request.Option) (*s3.UploadPartCopyOutput, error) {
			return &s3.UploadPartCopyOutput{}, nil
		},
		CreateMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.CreateMultipartUploadInput, opts ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
			return &s3.CreateMultipartUploadOutput{}, nil
		},
		CopyObjectWithContextFunc: func(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, error) {
			return &s3.CopyObjectOutput{}, nil
		},
		ListObjectsV2PagesWithContextFunc: func(ctx context.Context, input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool, opts ...request.Option) error {
			return nil
		},
	}
	suite.storageTrace = NewStorageTrace(suite.mockS3API, suite.tracer)
}

// TestGetObjectWithContext is a test function for the GetObjectWithContext method
func (suite *StorageTraceSuite) TestGetObjectWithContext() {
	input := &s3.GetObjectInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("my-key"),
	}

	_, err := suite.storageTrace.GetObjectWithContext(context.Background(), input)
	suite.Nil(err)
}

// TestDeleteObjectWithContext is a test function for the DeleteObjectWithContext method
func (suite *StorageTraceSuite) TestDeleteObjectWithContext() {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("my-key"),
	}

	_, err := suite.storageTrace.DeleteObjectWithContext(context.Background(), input)
	suite.Nil(err)

}

// TestPutObjectWithContext is a test function for the PutObjectWithContext method
func (suite *StorageTraceSuite) TestPutObjectWithContext() {
	input := &s3.PutObjectInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("my-key"),
		Body:   aws.ReadSeekCloser(strings.NewReader("test content")),
	}

	_, err := suite.storageTrace.PutObjectWithContext(context.Background(), input)
	suite.Nil(err)
}

// TestCompleteMultipartUploadWithContext is a test function for the CompleteMultipartUploadWithContext method
func (suite *StorageTraceSuite) TestCompleteMultipartUploadWithContext() {
	input := &s3.CompleteMultipartUploadInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("my-key"),
	}

	_, err := suite.storageTrace.CompleteMultipartUploadWithContext(context.Background(), input)
	suite.Nil(err)
}

// TestUploadPartCopyWithContext is a test function for the UploadPartCopyWithContext method
func (suite *StorageTraceSuite) TestUploadPartCopyWithContext() {
	input := &s3.UploadPartCopyInput{
		Bucket:     aws.String("my-bucket"),
		Key:        aws.String("my-key"),
		CopySource: aws.String("source-bucket/source-key"),
	}

	_, err := suite.storageTrace.UploadPartCopyWithContext(context.Background(), input)
	suite.Nil(err)
}

// TestCreateMultipartUploadWithContext is a test function for the CreateMultipartUploadWithContext method
func (suite *StorageTraceSuite) TestCreateMultipartUploadWithContext() {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("my-key"),
	}

	_, err := suite.storageTrace.CreateMultipartUploadWithContext(context.Background(), input)
	suite.Nil(err)
}

// TestCopyObjectWithContext is a test function for the CopyObjectWithContext method
func (suite *StorageTraceSuite) TestCopyObjectWithContext() {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String("my-bucket"),
		Key:        aws.String("my-key"),
		CopySource: aws.String("source-bucket/source-key"),
	}

	_, err := suite.storageTrace.CopyObjectWithContext(context.Background(), input)
	suite.Nil(err)
}

// TestListObjectsV2PagesWithContext is a test function for the ListObjectsV2PagesWithContext method
func (suite *StorageTraceSuite) TestListObjectsV2PagesWithContext() {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("my-bucket"),
	}

	err := suite.storageTrace.ListObjectsV2PagesWithContext(context.Background(), input, func(*s3.ListObjectsV2Output, bool) bool {
		return true
	})
	suite.Nil(err)
}

// TestStorageTraceSuite is a test function for the StorageTraceSuite
func TestStorageTraceSuite(t *testing.T) {
	suite.Run(t, new(StorageTraceSuite))
}
