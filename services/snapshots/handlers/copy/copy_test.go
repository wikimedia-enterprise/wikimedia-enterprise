package copy_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/copy"
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

func (m *s3Mock) HeadObjectWithContext(_ aws.Context, _ *s3.HeadObjectInput, _ ...request.Option) (*s3.HeadObjectOutput, error) {
	hop := &s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(m.Called().Int(0))),
	}

	return hop, m.Called().Error(1)
}

func (m *s3Mock) CopyObjectWithContext(_ aws.Context, _ *s3.CopyObjectInput, _ ...request.Option) (*s3.CopyObjectOutput, error) {
	return nil, m.Called().Error(0)
}

func (m *s3Mock) CreateMultipartUploadWithContext(_ aws.Context, _ *s3.CreateMultipartUploadInput, _ ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
	cmr := &s3.CreateMultipartUploadOutput{
		UploadId: aws.String(m.Called().String(0)),
	}

	return cmr, m.Called().Error(1)
}

func (m *s3Mock) UploadPartCopyWithContext(_ aws.Context, _ *s3.UploadPartCopyInput, _ ...request.Option) (*s3.UploadPartCopyOutput, error) {
	upr := &s3.UploadPartCopyOutput{
		CopyPartResult: &s3.CopyPartResult{ETag: aws.String(m.Called().String(0))},
	}

	return upr, m.Called().Error(1)
}

func (m *s3Mock) CompleteMultipartUploadWithContext(_ aws.Context, _ *s3.CompleteMultipartUploadInput, _ ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
	return nil, m.Called().Error(0)
}

type handlerTestSuite struct {
	suite.Suite
	hdl    *copy.Handler
	ctx    context.Context
	s3m    *s3Mock
	env    *env.Environment
	cln    int
	tag    string
	uid    string
	hdErr  error
	cpErr  error
	cmuErr error
	upErr  error
	cmpErr error
	req    *pb.CopyRequest
}

func (s *handlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		FreeTierGroup: "group_1",
	}
}

func (s *handlerTestSuite) SetupTest() {
	s.s3m = new(s3Mock)
	s.hdl = &copy.Handler{
		S3:  s.s3m,
		Env: s.env,
	}

	s.s3m.On("HeadObjectWithContext").Return(s.cln, s.hdErr)
	s.s3m.On("CopyObjectWithContext").Return(s.cpErr)
	s.s3m.On("CreateMultipartUploadWithContext").Return(s.uid, s.cmuErr)
	s.s3m.On("UploadPartCopyWithContext").Return(s.tag, s.upErr)
	s.s3m.On("CompleteMultipartUploadWithContext").Return(s.cmpErr)
}

func (s *handlerTestSuite) TestCopy() {
	res, err := s.hdl.Copy(s.ctx, s.req)

	s.Assert().NoError(err)

	if s.hdErr != nil || s.cpErr != nil || s.cmuErr != nil || s.upErr != nil || s.cmpErr != nil {
		s.Assert().Equal(int32(len(s.req.Projects)*2), res.Errors)
	} else {
		s.Assert().Equal(int32(0), res.Errors)
	}

	s.Assert().Equal(int32(len(s.req.Projects)*2), res.Total)
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 3221225472,
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			tag: "part-upload",
			uid: "uploadId",
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln:   3221225472,
			hdErr: errors.New("head object error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln:   3221225472,
			cpErr: errors.New("copy object error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln:    5368709120,
			cmuErr: errors.New("create multi part upload error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln:   5368709120,
			upErr: errors.New("upload part error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln:    5368709120,
			cmpErr: errors.New("complete multi part upload error"),
		},
	} {
		suite.Run(t, testCase)
	}
}
