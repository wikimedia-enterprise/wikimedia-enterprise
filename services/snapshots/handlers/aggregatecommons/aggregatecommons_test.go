package aggregatecommons_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/aggregatecommons"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
)

type s3Mock struct {
	mock.Mock
	s3tracerproxy.S3TracerProxy
}

func (s *s3Mock) ListObjectsV2PagesWithContext(_ aws.Context, inp *s3.ListObjectsV2Input, cb func(*s3.ListObjectsV2Output, bool) bool, _ ...request.Option) error {
	args := s.Called(inp)
	out := &s3.ListObjectsV2Output{}

	if keys, ok := args.Get(0).([]string); ok {
		for _, key := range keys {
			out.Contents = append(out.Contents, &s3.Object{
				Key: aws.String(key),
			})
		}
	}

	cb(out, true)
	return args.Error(1)
}

func (s *s3Mock) GetObjectWithContext(_ aws.Context, inp *s3.GetObjectInput, _ ...request.Option) (*s3.GetObjectOutput, error) {
	args := s.Called(inp)
	data := args.String(0)
	err := args.Error(1)

	if err != nil {
		return nil, err
	}

	out := &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader([]byte(data))),
	}

	return out, nil
}

type uploaderMock struct {
	mock.Mock
}

func (u *uploaderMock) Upload(input *s3manager.UploadInput, _ ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := u.Called(input)
	return &s3manager.UploadOutput{}, args.Error(0)
}

type handlerTestSuite struct {
	suite.Suite
	handler *aggregatecommons.Handler
	ctx     context.Context
	s3Mock  *s3Mock
	upMock  *uploaderMock
	env     *env.Environment
}

func (s *handlerTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.s3Mock = new(s3Mock)
	s.upMock = new(uploaderMock)
	s.env = &env.Environment{
		AWSBucket: "wme-data",
		LineLimit: 2,
	}
	s.handler = &aggregatecommons.Handler{
		S3:       s.s3Mock,
		Uploader: s.upMock,
		Env:      s.env,
	}
}

func (s *handlerTestSuite) TearDownTest() {
	s.s3Mock = new(s3Mock)
	s.upMock = new(uploaderMock)
}

func (s *handlerTestSuite) TestAggregateCommonsSuccess() {
	req := &pb.AggregateCommonsRequest{TimePeriod: "20230101"}
	keys := []string{"commons/batches/20230101/file1.json", "commons/batches/20230101/file2.json"}

	s.s3Mock.On("ListObjectsV2PagesWithContext", mock.Anything).Return(keys, nil).Once()
	s.s3Mock.On("GetObjectWithContext", mock.Anything).Return(`{"key": "value"}`, nil).Times(2)
	s.upMock.On("Upload", mock.AnythingOfType("*s3manager.UploadInput")).Return(nil).Once()

	res, err := s.handler.AggregateCommons(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(int32(2), res.Total)
	s.Equal(int32(0), res.Errors)

	s.s3Mock.AssertExpectations(s.T())
	s.upMock.AssertExpectations(s.T())
}

func TestAggregateCommonsHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
