package copy_test

import (
	"context"
	"errors"
	"regexp"
	"strings"
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
	TestSuite *handlerTestSuite
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

func (m *s3Mock) ListObjectsV2PagesWithContext(_ aws.Context, input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool, _ ...request.Option) error {
	prefix := aws.StringValue(input.Prefix)
	project := ""

	if parts := strings.Split(prefix, "/"); len(parts) > 1 {
		if matches := regexp.MustCompile(`^(.+?)_namespace_\d+$`).FindStringSubmatch(parts[1]); len(matches) == 2 {
			project = matches[1]
		}
	}

	suite := m.TestSuite

	if objs, ok := suite.chks[project]; ok {
		fn(&s3.ListObjectsV2Output{
			Contents: objs,
		}, true)
	} else {
		fn(&s3.ListObjectsV2Output{
			Contents: []*s3.Object{},
		}, true)
	}

	return nil
}

func (m *s3Mock) DeleteObjectsWithContext(_ aws.Context, input *s3.DeleteObjectsInput, _ ...request.Option) (*s3.DeleteObjectsOutput, error) {
	var dbj []string
	if input.Delete != nil {
		for _, obj := range input.Delete.Objects {
			dbj = append(dbj, *obj.Key)
		}
	}
	args := m.Called(dbj)
	return nil, args.Error(0)
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
	chks   map[string][]*s3.Object
}

func (s *handlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket:     "wme-primary",
		FreeTierGroup: "group_1",
	}
}

func (s *handlerTestSuite) SetupTest() {
	s.s3m = new(s3Mock)
	s.s3m.TestSuite = s
	s.hdl = &copy.Handler{
		S3:  s.s3m,
		Env: s.env,
	}

	s.s3m.On("HeadObjectWithContext").Return(s.cln, s.hdErr)
	s.s3m.On("CopyObjectWithContext").Return(s.cpErr)
	s.s3m.On("CreateMultipartUploadWithContext").Return(s.uid, s.cmuErr)
	s.s3m.On("UploadPartCopyWithContext").Return(s.tag, s.upErr)
	s.s3m.On("CompleteMultipartUploadWithContext").Return(s.cmpErr)
	s.s3m.On("DeleteObjectsWithContext", mock.Anything).Return(nil)

}

const maxUploadSizeBytes = 4294967296

func (s *handlerTestSuite) TestCopy() {

	res, err := s.hdl.Copy(s.ctx, s.req)

	s.Assert().NoError(err)

	cnt := 0
	for _, chunks := range s.chks {
		cnt = cnt + len(chunks)
	}

	if s.hdErr != nil || s.cpErr != nil || s.cmuErr != nil || s.upErr != nil || s.cmpErr != nil {
		if s.cln < maxUploadSizeBytes {
			s.Assert().Equal(int32(len(s.req.Projects)*2), res.Total)
		} else {
			s.Assert().Equal(int32(len(s.req.Projects)*2+cnt), res.Total)
		}
	} else {
		s.Assert().Equal(int32(0), res.Errors)
	}

	if s.hdErr != nil || s.cpErr != nil || s.cmuErr != nil || s.upErr != nil || s.cmpErr != nil {
		s.Assert().Equal(int32(len(s.req.Projects)*2+cnt), res.Total)
	} else {
		if s.cln < maxUploadSizeBytes {
			s.Assert().Equal(int32(len(s.req.Projects)*2), res.Total)
		} else {
			s.Assert().Equal(int32(len(s.req.Projects)*2+cnt), res.Total)
		}
	}
}

func (s *handlerTestSuite) TestCopyWithNoProjects() {
	s.req = &pb.CopyRequest{
		Workers:   10,
		Projects:  []string{},
		Namespace: 0,
	}

	res, err := s.hdl.Copy(s.ctx, s.req)
	s.Require().NoError(err)
	s.Assert().Equal(int32(0), res.Total)
	s.Assert().Equal(int32(0), res.Errors)
}

func (s *handlerTestSuite) TestCopyWithNilProjects() {
	s.req = &pb.CopyRequest{
		Workers:   10,
		Projects:  nil,
		Namespace: 0,
	}

	res, err := s.hdl.Copy(s.ctx, s.req)
	s.Require().NoError(err)
	s.Assert().Equal(int32(0), res.Total)

	if s.hdErr != nil || s.cpErr != nil || s.cmuErr != nil || s.upErr != nil || s.cmpErr != nil {
		s.Assert().Equal(int32(len(s.req.Projects)*2), res.Errors)
		return
	}

	s.Assert().Equal(int32(0), res.Errors)
}

func (s *handlerTestSuite) TestCopyWithDefaultWorkers() {
	s.req = &pb.CopyRequest{
		Projects:  []string{"dewiki"},
		Namespace: 0,
	}
	s.cln = 3221225472
	s.chks = map[string][]*s3.Object{
		"dewiki": {},
	}

	s.s3m.On("HeadObjectWithContext").Return(s.cln, nil)
	s.s3m.On("CopyObjectWithContext").Return(nil)

	res, err := s.hdl.Copy(s.ctx, s.req)
	s.Require().NoError(err)
	s.Assert().Equal(int32(2), res.Total)

	if s.hdErr != nil || s.cpErr != nil || s.cmuErr != nil || s.upErr != nil || s.cmpErr != nil {
		s.Assert().Equal(int32(len(s.req.Projects)*2), res.Errors)
		return
	}

	s.Assert().Equal(int32(0), res.Errors)
}

func (s *handlerTestSuite) TestInvalidChunkFilenamesIgnored() {
	s.req = &pb.CopyRequest{
		Projects:  []string{"enwiki"},
		Namespace: 0,
		Workers:   1,
	}
	s.cln = 3221225472
	s.chks = map[string][]*s3.Object{
		"enwiki": {
			{Key: aws.String("chunks/enwiki_namespace_0/invalid_chunk_file.txt")},
		},
	}

	s.s3m.On("HeadObjectWithContext").Return(s.cln, nil)
	s.s3m.On("CopyObjectWithContext").Return(nil)

	res, err := s.hdl.Copy(s.ctx, s.req)
	if err != nil {
		s.T().Fatalf("Error: %v", err)
		return
	}

	s.Require().NoError(err)
	s.Assert().Equal(int32(2), res.Total)

	if s.hdErr != nil || s.cpErr != nil {
		s.Assert().Equal(int32(len(s.req.Projects)*2), res.Errors)
		return
	}

	s.Assert().Equal(int32(0), res.Errors)
}

func (s *handlerTestSuite) TestGetChunkCopyPaths() {
	s.chks = map[string][]*s3.Object{
		"enwiki": {
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_1_group_1.json")},
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_2.tar.gz")},
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_2_group_1.tar.gz")},
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_3_group_1.json")},
			{Key: aws.String("chunks/enwiki_namespace_0/random_file.txt")},
		},
	}

	s.s3m.On("DeleteObjectsWithContext", mock.MatchedBy(func(input *s3.DeleteObjectsInput) bool {
		var deletedKeys []string
		for _, obj := range input.Delete.Objects {
			deletedKeys = append(deletedKeys, *obj.Key)
		}
		s.Assert().Contains(deletedKeys, "chunks/enwiki_namespace_0/chunk_1_group_1.json")
		s.Assert().Contains(deletedKeys, "chunks/enwiki_namespace_0/chunk_2_group_1.tar.gz")
		s.Assert().Contains(deletedKeys, "chunks/enwiki_namespace_0/chunk_3_group_1.json")
		s.Assert().NotContains(deletedKeys, "chunks/enwiki_namespace_0/chunk_1.json")
		s.Assert().NotContains(deletedKeys, "chunks/enwiki_namespace_0/chunk_2.tar.gz")
		s.Assert().NotContains(deletedKeys, "chunks/enwiki_namespace_0/random_file.txt")

		s.Assert().Len(deletedKeys, 3)

		return true
	})).Return(nil)

	prefix := "chunks/enwiki_namespace_0/"
	paths, err := s.hdl.CopyChunkFiles(s.ctx, prefix, "enwiki", 0)
	s.Require().NoError(err)

	s.Assert().Len(paths, 2)
	s.Assert().Equal("chunks/enwiki_namespace_0/chunk_1_group_1.json", paths["chunks/enwiki_namespace_0/chunk_1.json"])
	s.Assert().Equal("chunks/enwiki_namespace_0/chunk_2_group_1.tar.gz", paths["chunks/enwiki_namespace_0/chunk_2.tar.gz"])
	s.s3m.AssertNotCalled(s.T(), "DeleteObjectsWithContext", mock.Anything)
}

func (s *handlerTestSuite) TestGetChunkCopyPathsWithDeletionError() {
	s.chks = map[string][]*s3.Object{
		"enwiki": {
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
			{Key: aws.String("chunks/enwiki_namespace_0/chunk_1_group_1.json")},
		},
	}
	s.s3m.On("DeleteObjectsWithContext", mock.Anything).Return(errors.New("deletion failed"))

	prefix := "chunks/enwiki_namespace_0/"
	paths, err := s.hdl.CopyChunkFiles(s.ctx, prefix, "enwiki", 0)
	s.Require().NoError(err)
	s.Assert().Len(paths, 1)
	s.Assert().Equal("chunks/enwiki_namespace_0/chunk_1_group_1.json", paths["chunks/enwiki_namespace_0/chunk_1.json"])
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
			chks: map[string][]*s3.Object{
				"dewiki": {},
				"enwiki": {},
			},
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			chks: map[string][]*s3.Object{
				"dewiki": {
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.tar.gz")},
				},
				"enwiki": {
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.tar.gz")},
				},
			},
			tag: "part-upload",
			uid: "uploadId",
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 3221225472,
			chks: map[string][]*s3.Object{
				"dewiki": {},
				"enwiki": {},
			},
			hdErr: errors.New("head object error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 3221225472,
			chks: map[string][]*s3.Object{
				"dewiki": {},
				"enwiki": {},
			},
			cpErr: errors.New("copy object error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			chks: map[string][]*s3.Object{
				"dewiki": {
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.tar.gz")},
				},
				"enwiki": {
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.tar.gz")},
				},
			},
			cmuErr: errors.New("create multi part upload error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			chks: map[string][]*s3.Object{
				"dewiki": {
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.tar.gz")},
				},
				"enwiki": {
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.tar.gz")},
				},
			},
			upErr: errors.New("upload part error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			chks: map[string][]*s3.Object{
				"dewiki": {},
				"enwiki": {},
			},
			cmpErr: errors.New("complete multi part upload error"),
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			chks: map[string][]*s3.Object{
				"dewiki": {
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.tar.gz")},
				},
				"enwiki": {
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.tar.gz")},
				},
			},
		},
		{
			req: &pb.CopyRequest{
				Workers:   10,
				Projects:  []string{"dewiki", "enwiki"},
				Namespace: 0,
			},
			cln: 5368709120,
			chks: map[string][]*s3.Object{
				"dewiki": {
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/dewiki_namespace_0/chunk_0.tar.gz")},
				},
				"enwiki": {
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_0.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_1.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_2.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_2.tar.gz")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_3.json")},
					{Key: aws.String("chunks/enwiki_namespace_0/chunk_3.tar.gz")},
				},
			},
		},
	} {

		suite.Run(t, testCase)
	}
}
