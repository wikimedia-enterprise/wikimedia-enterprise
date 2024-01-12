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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type s3Mock struct {
	mock.Mock
	s3iface.S3API
}

func (s *s3Mock) ListObjectsV2PagesWithContext(_ aws.Context, inp *s3.ListObjectsV2Input, cb func(*s3.ListObjectsV2Output, bool) bool, opt ...request.Option) error {
	arg := s.Called(inp)
	out := &s3.ListObjectsV2Output{}

	if kys, ok := arg.Get(0).([]string); ok {
		for _, key := range kys {
			out.Contents = append(out.Contents, &s3.Object{
				Key: &key,
			})
		}
	}

	cb(out, true)
	return arg.Error(1)
}

func (s *s3Mock) GetObjectWithContext(_ aws.Context, inp *s3.GetObjectInput, _ ...request.Option) (*s3.GetObjectOutput, error) {
	arg := s.Called(inp)
	data, err := json.Marshal(arg.Get(1))

	if err != nil {
		return nil, err
	}

	out := &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(data)),
	}

	return out, arg.Error(1)
}

func (s *s3Mock) PutObjectWithContext(_ aws.Context, inp *s3.PutObjectInput, _ ...request.Option) (*s3.PutObjectOutput, error) {
	arg := s.Called(inp.Key, inp.Bucket)
	return nil, arg.Error(0)
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

	prx := "snapshots"

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

	for i, key := range s.kys {
		s.kys[i] = key
		s.gns = append(s.gns, &s3.GetObjectInput{
			Bucket: aws.String(s.env.AWSBucket),
			Key:    aws.String(key),
		})
	}
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
	s.Assert().Equal(res, s.res)
}

func (s *handlerTestSuite) TestHandlerListErr() {
	s.s3m.On("ListObjectsV2PagesWithContext", s.lin).Return(s.kys, s.err)

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
	for _, testCase := range []*handlerTestSuite{
		{
			req: &pb.AggregateRequest{
				Prefix: "snapshots",
			},
			res: &pb.AggregateResponse{Total: 2},
			kys: []string{"snapshots/%s/enwiki.json", "snapshots/%s/afwikibooks.json"},
			hls: []*schema.Project{
				{
					Identifier: "enwiki",
				},
				{
					Identifier: "afwikibooks",
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
