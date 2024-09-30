package aggregate_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/aggregate"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type s3Mock struct {
	mock.Mock
	s3tracerproxy.S3TracerProxy
}

func (s *s3Mock) ListObjectsV2PagesWithContext(_ aws.Context, inp *s3.ListObjectsV2Input, cb func(*s3.ListObjectsV2Output, bool) bool, opt ...request.Option) error {
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
	data, err := json.Marshal(args.Get(0))

	if err != nil {
		return nil, err
	}

	out := &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(data)),
	}

	return out, args.Error(1)
}

func (s *s3Mock) PutObjectWithContext(_ aws.Context, inp *s3.PutObjectInput, _ ...request.Option) (*s3.PutObjectOutput, error) {
	args := s.Called(inp.Key, inp.Bucket)
	return &s3.PutObjectOutput{}, args.Error(0)
}

type handlerTestSuite struct {
	suite.Suite
	hdl *aggregate.Handler
	ctx context.Context
	s3m *s3Mock
	env *env.Environment
	req *pb.AggregateRequest
	res *pb.AggregateResponse
	kys []string
	hls []*schema.Project
	lin *s3.ListObjectsV2Input
	gns []*s3.GetObjectInput
	pin *s3.PutObjectInput
	err error
}

func (s *handlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		Prefix:        "snapshots",
		FreeTierGroup: "group_1",
	}

	s.req = &pb.AggregateRequest{
		Prefix: "snapshots",
	}

	prx := s.env.Prefix

	if len(s.req.Prefix) > 0 {
		prx = s.req.Prefix
	}

	s.lin = &s3.ListObjectsV2Input{
		Bucket: aws.String(s.env.AWSBucket),
		Prefix: aws.String(prx),
	}
	s.pin = &s3.PutObjectInput{
		Bucket: aws.String(s.env.AWSBucket),
		Key:    aws.String(fmt.Sprintf("aggregations/%[1]s/%[1]s.ndjson", prx)),
	}
	s.err = errors.New("handler test err")

	s.kys = []string{
		fmt.Sprintf("%s/enwiki.json", prx),
		fmt.Sprintf("%s/afwikibooks.json", prx),
	}

	for _, key := range s.kys {
		s.gns = append(s.gns, &s3.GetObjectInput{
			Bucket: aws.String(s.env.AWSBucket),
			Key:    aws.String(key),
		})
	}

	s.hls = []*schema.Project{
		{
			Identifier: "enwiki",
			Size:       &schema.Size{Value: 1},
		},
		{
			Identifier: "afwikibooks",
			Size:       &schema.Size{Value: 1},
		},
	}

	s.res = &pb.AggregateResponse{Total: 2}
}

func (s *handlerTestSuite) SetupTest() {
	s.s3m = new(s3Mock)
	s.hdl = &aggregate.Handler{
		S3:  s.s3m,
		Env: s.env,
	}
}

func (s *handlerTestSuite) TestHandler() {
	s.s3m.On("ListObjectsV2PagesWithContext", s.lin).Return(s.kys, nil)

	for i, inp := range s.gns {
		s.s3m.On("GetObjectWithContext", inp).Return(s.hls[i], nil)
	}

	s.s3m.On("PutObjectWithContext", s.pin.Key, s.pin.Bucket).Return(nil)

	res, err := s.hdl.Aggregate(s.ctx, s.req)
	s.Assert().NoError(err)
	s.Assert().Equal(s.res, res)
}

func (s *handlerTestSuite) TestHandlerListErr() {
	s.s3m.On("ListObjectsV2PagesWithContext", s.lin).Return(nil, s.err)

	res, err := s.hdl.Aggregate(s.ctx, s.req)
	s.Assert().Equal(s.err, err)
	s.Assert().Nil(res)
}

func (s *handlerTestSuite) TestHandlerPutErr() {
	s.s3m.On("ListObjectsV2PagesWithContext", s.lin).Return(s.kys, nil)

	for i, inp := range s.gns {
		s.s3m.On("GetObjectWithContext", inp).Return(s.hls[i], nil)
	}

	s.s3m.On("PutObjectWithContext", s.pin.Key, s.pin.Bucket).Return(s.err)

	res, err := s.hdl.Aggregate(s.ctx, s.req)
	s.Assert().Equal(s.err, err)
	s.Assert().Nil(res)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
