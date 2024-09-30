package aggregatecopy_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/aggregatecopy"
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

func (m *s3Mock) CopyObjectWithContext(_ aws.Context, _ *s3.CopyObjectInput, _ ...request.Option) (*s3.CopyObjectOutput, error) {
	return nil, m.Called().Error(0)
}

type handlerTestSuite struct {
	suite.Suite
	hdl   *aggregatecopy.Handler
	ctx   context.Context
	s3m   *s3Mock
	env   *env.Environment
	cpErr error
	req   *pb.AggregateCopyRequest
}

func (s *handlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		FreeTierGroup: "group_1",
		Prefix:        "snapshots",
	}

	s.req = &pb.AggregateCopyRequest{}
}

func (s *handlerTestSuite) SetupTest() {
	s.s3m = new(s3Mock)
	s.hdl = &aggregatecopy.Handler{
		S3:  s.s3m,
		Env: s.env,
	}

	s.s3m.On("CopyObjectWithContext").Return(s.cpErr)
}

func (s *handlerTestSuite) TestCopy() {
	res, err := s.hdl.AggregateCopy(s.ctx, s.req)

	s.Assert().NoError(err)

	if s.cpErr != nil {
		s.Assert().Equal(int32(1), res.Errors)
	} else {
		s.Assert().Equal(int32(0), res.Errors)
	}

	s.Assert().Equal(int32(1), res.Total)
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			cpErr: nil,
		},
		{
			cpErr: errors.New("copy object error"),
		},
	} {
		suite.Run(t, testCase)
	}
}
