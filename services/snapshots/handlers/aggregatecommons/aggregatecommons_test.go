package aggregatecommons_test

import (
	"bytes"
	"context"
	"fmt"
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
	keys       []string
	objectData map[string]string
	s3tracerproxy.S3TracerProxy
}

func (s *s3Mock) ListObjectsV2PagesWithContext(_ aws.Context, inp *s3.ListObjectsV2Input, cb func(*s3.ListObjectsV2Output, bool) bool, _ ...request.Option) error {
	out := &s3.ListObjectsV2Output{}

	for _, key := range s.keys {
		out.Contents = append(out.Contents, &s3.Object{
			Key: aws.String(key),
		})
	}

	cb(out, true)
	return nil
}

func (s *s3Mock) GetObjectWithContext(_ aws.Context, inp *s3.GetObjectInput, _ ...request.Option) (*s3.GetObjectOutput, error) {
	data, ok := s.objectData[*inp.Key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", *inp.Key)
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
	if input.Body != nil {
		_, err := io.ReadAll(input.Body)
		if err != nil {
			return nil, err
		}
	}
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
	s.s3Mock = &s3Mock{
		keys: []string{
			"commons/batches/20230101/file1.json",
			"commons/batches/20230101/file2.json",
		},
		objectData: map[string]string{
			"commons/batches/20230101/file1.json": `{"key": "value"}`,
			"commons/batches/20230101/file2.json": `{"key": "value"}`,
		},
	}
	s.upMock = new(uploaderMock)
	s.env = &env.Environment{
		AWSBucketCommons:                 "wme-data",
		AWSBucket:                        "wme-primary",
		LineLimit:                        2,
		CommonsAggregationWorkercount:    10,
		CommonsAggregationKeyChannelSize: 10,
		PipeBufferSize:                   1024 * 1024,
		UploadPartSize:                   5 * 1024 * 1024,
		UploadConcurrency:                3,
		LogInterval:                      15,
	}
	s.handler = &aggregatecommons.Handler{
		S3:       s.s3Mock,
		Uploader: s.upMock,
		Env:      s.env,
	}
}

func (s *handlerTestSuite) TestAggregateCommonsSuccess() {
	req := &pb.AggregateCommonsRequest{TimePeriod: "20230101"}

	s.upMock.On("Upload", mock.Anything).Return(nil).Once()

	res, err := s.handler.AggregateCommons(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(int32(2), res.Total)
	s.Equal(int32(0), res.Errors)

	s.upMock.AssertExpectations(s.T())
}

func TestAggregateCommonsHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
