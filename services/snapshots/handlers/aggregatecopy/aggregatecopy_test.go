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
		AWSBucket:     "wme-primary",
		FreeTierGroup: "group_1",
		Prefix:        "snapshots",
	}
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

	errors := len(s.req.Projects) * len(s.req.Namespaces)
	if s.cpErr == nil {
		errors = 0
	}

	s.Assert().Equal(int32(errors), res.Errors)

	copied := len(s.req.Projects) * len(s.req.Namespaces)
	if len(s.req.Projects) == 0 && len(s.req.Namespaces) == 0 {
		copied = 1 // copy the root metadata file
	}

	s.Assert().Equal(int32(copied), res.Total)
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			req: &pb.AggregateCopyRequest{
				Projects:   []string{"project_1", "project_2"},
				Namespaces: []int32{0, 16},
			},
			cpErr: nil,
		},
		{
			req: &pb.AggregateCopyRequest{
				Projects:   []string{"project_1", "project_2"},
				Namespaces: []int32{0, 16},
			},
			cpErr: errors.New("copy object error"),
		},
		{
			req: &pb.AggregateCopyRequest{
				Projects:   []string{"project_1"},
				Namespaces: []int32{0},
			},
			cpErr: nil,
		},
		{
			req: &pb.AggregateCopyRequest{
				Projects:   []string{},
				Namespaces: []int32{},
			},
			cpErr: nil,
		},
	} {
		suite.Run(t, testCase)
	}
}
