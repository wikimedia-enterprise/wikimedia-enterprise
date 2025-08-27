package export_test

import (
	"context"
	"crypto/md5" // #nosec G501
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/export"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	libkafka "wikimedia-enterprise/services/snapshots/libraries/kafka"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"
	"wikimedia-enterprise/services/snapshots/submodules/config"
	"wikimedia-enterprise/services/snapshots/submodules/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	nbf "github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type hashReaderTestSuite struct {
	suite.Suite
	rdr *export.HashReader
	pld []byte
	sum string
}

func (s *hashReaderTestSuite) SetupTest() {
	s.pld = []byte("test")

	buf := nbf.New(int64(len(s.pld)))
	prr, pwr := nio.Pipe(buf)
	_, err := pwr.Write(s.pld)
	s.Assert().NoError(err)

	s.sum = fmt.Sprintf("%x", md5.Sum(s.pld))
	s.rdr = new(export.HashReader)
	s.rdr.PipeReader = prr
}

func (s *hashReaderTestSuite) TestRead() {
	pld := make([]byte, len(s.pld))
	_, err := s.rdr.Read(pld)

	s.Assert().NoError(err)
	s.Assert().Equal(s.pld, pld)
}

func (s *hashReaderTestSuite) TestSum() {
	pld := make([]byte, len(s.pld))
	_, err := s.rdr.Read(pld)

	s.Assert().NoError(err)
	s.Assert().Equal(s.pld, pld)
	s.Assert().Equal(s.sum, s.rdr.Sum())
}

func TestHashReader(t *testing.T) {
	suite.Run(t, new(hashReaderTestSuite))
}

type consumerMock struct {
	mock.Mock
	libkafka.Consumer
}

func (m *consumerMock) ReadAll(_ context.Context, snc int, tpc string, partition int, _ string, cb func(msg *kafka.Message) error) error {
	ags := m.Called(snc, tpc, partition, cb)
	return ags.Error(0)
}

func (m *consumerMock) GetMetadata(tpc string) (*kafka.TopicMetadata, error) {
	ags := m.Called(tpc)
	return ags.Get(0).(*kafka.TopicMetadata), ags.Error(1)
}

func (m *consumerMock) Close() error {
	return m.Called().Error(0)
}

type poolMock struct {
	mock.Mock
}

func (s *poolMock) GetConsumer(idr string) (libkafka.ReadAllCloser, error) {
	ags := s.Called(idr)
	return ags.Get(0).(libkafka.ReadAllCloser), ags.Error(1)
}

type uploaderMock struct {
	mock.Mock
}

func (m *uploaderMock) Upload(uin *s3manager.UploadInput, _ ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	ags := m.Called(*uin)

	_, err := io.ReadAll(uin.Body)
	if err != nil {
		fmt.Printf("error in Upload reading body: %v\n", err)
		return nil, err
	}

	return nil, ags.Error(0)
}

type s3Mock struct {
	mock.Mock
	s3tracerproxy.S3TracerProxy
}

func (m *s3Mock) HeadObjectWithContext(_ aws.Context, hin *s3.HeadObjectInput, _ ...request.Option) (*s3.HeadObjectOutput, error) {
	hop := &s3.HeadObjectOutput{
		ContentLength: aws.Int64(0),
		ETag:          aws.String(""),
	}

	return hop, m.Called(*hin.Key).Error(0)
}

func (m *s3Mock) PutObjectWithContext(_ aws.Context, pin *s3.PutObjectInput, _ ...request.Option) (*s3.PutObjectOutput, error) {
	return nil, m.
		Called(*pin).
		Error(0)
}

func (m *s3Mock) ListObjectsV2PagesWithContext(_ aws.Context, input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool, _ ...request.Option) error {
	args := m.Called(input.Bucket, input.Prefix)
	lst := args.Get(0).([]*s3.Object)
	err := args.Error(1)

	if err != nil {
		return err
	}

	var sbj []*s3.Object
	Pfx := *input.Prefix
	for _, obj := range lst {
		if strings.HasPrefix(*obj.Key, Pfx) {
			sbj = append(sbj, obj)
		}
	}

	output := &s3.ListObjectsV2Output{
		Contents: sbj,
	}
	fn(output, true)
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

type configMock struct {
	config.API
	mock.Mock
}

func (c *configMock) GetPartitions(dtb string, nid int) []int {
	return c.Called(dtb, nid).Get(0).([]int)
}

func (c *configMock) GetStructuredProjects() []string {
	return c.Called().Get(0).([]string)
}

type mockUnmarshaler struct {
	mock.Mock
}

func (m *mockUnmarshaler) Unmarshal(ctx context.Context, data []byte, v interface{}) error {
	args := m.Called(ctx, data, v)
	return args.Error(0)
}

func (m *mockUnmarshaler) UnmarshalNoCache(ctx context.Context, data []byte, v interface{}) error {
	args := m.Called(ctx, data, v)
	return args.Error(0)
}

type handlerTestSuite struct {
	suite.Suite
	ctx            context.Context
	tpc            string
	req            *pb.ExportRequest
	res            *pb.ExportResponse
	hdr            *export.Handler
	partitions     []int
	stp            []string
	expectedPrefix string
	ecr            error
	eur            error
	era            error
	eho            error
	epo            error
	mss            []*kafka.Message
	unm            *mockUnmarshaler
}

func matchTaggedPut(path string, tag string) any {
	return mock.MatchedBy(func(in s3.PutObjectInput) bool {
		return *in.Key == path && *in.Tagging == tag
	})
}

func matchTaggedUpload(path string, tag string) any {
	return mock.MatchedBy(func(in s3manager.UploadInput) bool {
		return *in.Key == path && *in.Tagging == tag
	})
}

func (s *handlerTestSuite) SetupTest() {
	s.ctx = context.Background()

	cfm := new(configMock)
	cfm.On("GetPartitions", s.req.GetProject(), int(s.req.GetNamespace())).Return(s.partitions)
	cfm.On("GetStructuredProjects").Return(s.stp)
	csr := new(consumerMock)
	csr.On("Close").Return(nil)

	idr := fmt.Sprintf("%s_namespace_%d", s.req.Project, s.req.Namespace)

	pol := new(poolMock)
	upr := new(uploaderMock)
	s3m := new(s3Mock)

	expectedTag := "type=" + s.req.Prefix
	expectedTagChunks := "type=" + s.req.Prefix + "_chunks"

	s3m.On("ListObjectsV2PagesWithContext", mock.Anything, mock.Anything).Return([]*s3.Object{}, nil)

	if s.req.Type == "structured" {
		s.mss = []*kafka.Message{
			{
				Value:     []byte(`{"title": "Test Article 1", "content": "This is test content 1"}`),
				Timestamp: time.Now().Add(-time.Hour),
			},
			{
				Value:     []byte(`{"title": "Test Article 2", "content": "This is test content 2"}`),
				Timestamp: time.Now().Add(-30 * time.Minute),
			},
		}

		csr.On("ReadAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			callback := args.Get(3).(func(*kafka.Message) error)
			for _, msg := range s.mss {
				if err := callback(msg); err != nil {
					return
				}
			}
		})

		pol.On("GetConsumer", fmt.Sprintf("structured-contents_%s", idr)).
			Return(csr, s.ecr)

		upr.On("Upload", matchTaggedUpload(fmt.Sprintf("structured-contents/%s.tar.gz", idr), expectedTag)).
			Return(s.eur)
		if s.req.EnableChunking {
			upr.On("Upload", matchTaggedUpload(fmt.Sprintf("structured-contents/%s.tar.gz", idr), expectedTagChunks)).
				Return(s.eur)
		}

		if s.req.EnableChunking {
			for i := 0; i < 20; i++ {
				s3m.On("PutObjectWithContext", matchTaggedPut(fmt.Sprintf("structured-contents/%s/chunk_%d.tar.gz", idr, i), expectedTagChunks)).
					Return(s.eur)
				s3m.On("PutObjectWithContext", matchTaggedPut(fmt.Sprintf("structured-contents/%s/chunk_%d.json", idr, i), expectedTagChunks)).
					Return(s.eur)
			}
		}

		s3m.On("HeadObjectWithContext", fmt.Sprintf("structured-contents/%s.tar.gz", idr)).
			Return(s.eho)
		s3m.On("PutObjectWithContext", matchTaggedPut(fmt.Sprintf("structured-contents/%s.json", idr), expectedTag)).
			Return(s.epo)

	} else {
		s.mss = []*kafka.Message{
			{
				Value:     []byte(`{"name": "Test Article 1", "abstract": "This is test content 1"}`),
				Timestamp: time.Now().Add(-time.Hour),
			},
			{
				Value:     []byte(`{"name": "Test Article 2", "abstract": "This is test content 2"}`),
				Timestamp: time.Now().Add(-30 * time.Minute),
			},
		}
		for _, pt := range s.partitions {
			csr.On("ReadAll", int(s.req.Since), s.tpc, pt, mock.Anything).Return(s.era).Run(func(args mock.Arguments) {
				callback := args.Get(3).(func(*kafka.Message) error)
				for _, msg := range s.mss {
					if err := callback(msg); err != nil {
						return
					}
				}
			})
		}

		pol.On("GetConsumer", fmt.Sprintf("%s_%s", s.req.GetPrefix(), idr)).
			Return(csr, s.ecr)

		expectedTarPath := fmt.Sprintf("%s/%s.tar.gz", s.expectedPrefix, idr)
		upr.On("Upload", matchTaggedUpload(expectedTarPath, expectedTag)).Return(s.eur)

		if s.req.EnableChunking {
			for i := 0; i < 20; i++ {
				s3m.On("PutObjectWithContext", matchTaggedPut(fmt.Sprintf("chunks/%s/chunk_%d.tar.gz", idr, i), expectedTagChunks)).
					Return(s.eur)
				s3m.On("PutObjectWithContext", matchTaggedPut(fmt.Sprintf("chunks/%s/chunk_%d.json", idr, i), expectedTagChunks)).
					Return(s.eur)
			}
		}

		s3m.On("HeadObjectWithContext", expectedTarPath).
			Return(s.eho)
		s3m.On("PutObjectWithContext", matchTaggedPut(fmt.Sprintf("%s/%s.json", s.expectedPrefix, idr), expectedTag)).
			Return(s.epo)
	}

	tps := new(schema.Topics)

	// Different service name used by UnmarshalEnvironmentValue and GetNameByVersion
	if s.req.Type == "structured" {
		tps.ServiceName = "structured-contents"
	}

	s.Assert().NoError(
		tps.UnmarshalEnvironmentValue(""),
	)

	s.unm = new(mockUnmarshaler)
	s.unm.On("Unmarshal", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		data := args.Get(1).([]byte)
		obj := args.Get(2)

		objType := reflect.TypeOf(obj)
		if objType == nil {
			s.Assert().Fail("unexpected nil in Unmarshal cbk")
		}

		objValue := reflect.ValueOf(obj)
		if objValue.Kind() != reflect.Ptr {
			s.Assert().Fail("unexpected non-pointer type in Unmarshal cbk")
		}

		switch objType.Elem().Name() {
		case "Article":
			v, ok := obj.(*schema.Article)
			if !ok {
				s.Assert().Fail("unexpected Article type in Unmarshal cbk")
			}
			err := json.Unmarshal(data, v)
			if err != nil {
				s.Assert().Fail("unexpected Article error in Unmarshal cbk")
			}
		case "AvroStructured":
			v, ok := obj.(*schema.AvroStructured)
			if !ok {
				s.Assert().Fail("unexpected Structured type in Unmarshal cbk")
			}
			err := json.Unmarshal(data, v)
			if err != nil {
				s.Assert().Fail("unexpected Structured error in Unmarshal cbk")
			}
		default:
			s.Assert().Fail("unexpected type neither Article or Structured in Unmarshal cbk")
		}
	})

	s.hdr = &export.Handler{
		Env: &env.Environment{
			Topics:         tps,
			PipeBufferSize: 1024,
			ChunkWorkers:   3,
			FreeTierGroup:  "group_1",
		},
		Cfg:      cfm,
		Uploader: upr,
		Pool:     pol,
		S3:       s3m,
		Stream:   s.unm,
	}
}

func (s *handlerTestSuite) TestExport() {
	s.hdr.Env.PipeBufferSize = 64
	s.hdr.Env.BufferSize = 16

	res, err := s.hdr.Export(s.ctx, s.req)

	if s.ecr != nil {
		s.Assert().Equal(s.ecr, err)
	} else if s.eur != nil {
		s.Assert().Equal(s.eur, err)
	} else if s.era != nil {
		s.Assert().Equal(s.era, err)
	} else if s.eho != nil {
		s.Assert().Equal(s.eho, err)
	} else if s.epo != nil {
		s.Assert().Equal(s.epo, err)
	} else {
		s.Assert().Equal(s.res, res)
		s.Assert().NoError(err)
	}
}

func TestHandler(t *testing.T) {
	stp := []string{"enwiki", "eswiki", "frwiki"}
	for _, testCase := range []*handlerTestSuite{
		{
			req: &pb.ExportRequest{
				Project:  "enwiki",
				Language: "en",
			},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
			ecr:        errors.New("could not get a consumer"),
		},
		{
			req: &pb.ExportRequest{
				Project:  "enwiki",
				Language: "en",
			},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
			eur:        errors.New("upload failed"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
			era:        errors.New("read all error"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			res:        &pb.ExportResponse{},
			partitions: []int{20, 40},
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
			eho:        errors.New("head object error"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			res:        &pb.ExportResponse{},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
			epo:        errors.New("put object error"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			res:        &pb.ExportResponse{Total: 4},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
				Type:      "structured",
				Prefix:    "structured-contents",
			},
			res:        &pb.ExportResponse{Total: 4},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-contents.enwiki-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace:      0,
				Project:        "enwiki",
				Language:       "en",
				EnableChunking: true,
			},
			res:        &pb.ExportResponse{Total: 4},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-data.enwiki-articles-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace:      0,
				Project:        "enwiki",
				Language:       "en",
				Type:           "structured",
				Prefix:         "structured-contents",
				EnableChunking: true,
			},
			res:        &pb.ExportResponse{Total: 4},
			partitions: []int{20, 40},
			stp:        stp,
			tpc:        "aws.structured-contents.enwiki-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace:      0,
				Project:        "dewiki",
				Language:       "de",
				Type:           "article",
				EnableChunking: true,
			},
			res:        &pb.ExportResponse{Total: 8},
			partitions: []int{20, 40, 60, 80},
			stp:        stp,
			tpc:        "aws.structured-data.dewiki-articles-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "dewiki",
				Language:  "de",
				Type:      "article",
				Prefix:    "batches",
				Since:     time.Date(2025, 6, 26, 1, 0, 0, 0, time.UTC).UnixMilli(),
			},
			expectedPrefix: "batches/2025-06-26",
			res:            &pb.ExportResponse{Total: 8},
			partitions:     []int{20, 40, 60, 80},
			stp:            stp,
			tpc:            "aws.structured-data.dewiki-articles-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace:                  0,
				Project:                    "dewiki",
				Language:                   "de",
				Type:                       "article",
				Prefix:                     "batches",
				Since:                      time.Date(2025, 6, 26, 1, 0, 0, 0, time.UTC).UnixMilli(),
				EnableNonCumulativeBatches: true,
			},
			expectedPrefix: "batches/2025-06-26/01",
			res:            &pb.ExportResponse{Total: 8},
			partitions:     []int{20, 40, 60, 80},
			stp:            stp,
			tpc:            "aws.structured-data.dewiki-articles-compacted.v1",
		},
	} {
		suite.Run(t, testCase)
	}
}

type chunkTestSuite struct {
	suite.Suite
	s3m *s3Mock
	hdr *export.Handler
	ctx context.Context
	opt *export.ChunkOptions
}

func (s *chunkTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.s3m = new(s3Mock)
	s.hdr = &export.Handler{
		Env: &env.Environment{
			AWSBucket: "bucket",
		},
		S3: s.s3m,
	}
	s.opt = &export.ChunkOptions{
		Ctx: s.ctx,
		Key: "enwiki_namespace_0",
		Request: &pb.ExportRequest{
			Project:   "enwiki",
			Namespace: 0,
			Language:  "en",
		},
		Counter: &export.Counter{Itr: -1},
	}
}

func (s *chunkTestSuite) TestHandleChunkSuccess() {
	key1 := "chunks/enwiki_namespace_0/chunk_0.tar.gz"
	key2 := "chunks/enwiki_namespace_0/chunk_0.json"

	s.s3m.On("PutObjectWithContext", matchTaggedPut(key1, "type=_chunks")).Return(nil)
	s.s3m.On("PutObjectWithContext", matchTaggedPut(key2, "type=_chunks")).Return(nil)

	cnm, err := s.hdr.HandleChunk([]byte(`{"test":"data"}`), s.opt)
	s.NoError(err)
	s.NotNil(cnm)
}

func (s *chunkTestSuite) TestHandleChunkPutError() {
	key := "chunks/enwiki_namespace_0/chunk_0.tar.gz"

	s.s3m.On("PutObjectWithContext", matchTaggedPut(key, "type=_chunks")).Return(errors.New("put error"))

	_, err := s.hdr.HandleChunk([]byte(`{"test":"data"}`), s.opt)
	s.Error(err)
}

func TestChunk(t *testing.T) {
	suite.Run(t, new(chunkTestSuite))
}

type cleanupTestSuite struct {
	suite.Suite
	s3m *s3Mock
	hdr *export.Handler
	ctx context.Context
	key string
}

func (s *cleanupTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.s3m = new(s3Mock)
	s.hdr = &export.Handler{
		Env: &env.Environment{
			AWSBucket:     "bucket",
			FreeTierGroup: "group_1",
		},
		S3: s.s3m,
	}
	s.key = "enwiki_namespace_0"
}

func (s *cleanupTestSuite) TestCleanupChunksSuccess() {
	pfx := aws.String(fmt.Sprintf("chunks/%s/", s.key))
	bkt := aws.String(s.hdr.Env.AWSBucket)

	lst := []*s3.Object{
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_0.tar.gz", s.key))},
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_0.json", s.key))},
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_2.txt", s.key))},                               // This should be filtered by CleanupChunks as it's not a .tar.gz or .json
		{Key: aws.String(fmt.Sprintf("chunks/%schunk_3_%s.tar.gz", s.key, s.hdr.Env.FreeTierGroup))}, // Should be filtered
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_3_%s.json", s.key, s.hdr.Env.FreeTierGroup))},  // Should be filtered
		{Key: aws.String("other_key/chunk_0.tar.gz")},                                                // Should be filtered by CleanupChunks
	}
	dbj := []string{
		fmt.Sprintf("chunks/%s/chunk_0.tar.gz", s.key),
		fmt.Sprintf("chunks/%s/chunk_0.json", s.key),
	}

	s.s3m.On("ListObjectsV2PagesWithContext", bkt, pfx).Return(lst, nil).Once()

	s.s3m.On("DeleteObjectsWithContext", dbj).Return(nil).Once()

	s.hdr.CleanupChunks(s.ctx, s.key)

	s.s3m.AssertExpectations(s.T())
}

func (s *cleanupTestSuite) TestCleanupChunksListError() {
	pfx := aws.String(fmt.Sprintf("chunks/%s/", s.key))
	bkt := aws.String(s.hdr.Env.AWSBucket)

	ler := errors.New("list objects failed")

	s.s3m.On("ListObjectsV2PagesWithContext", bkt, pfx).Return([]*s3.Object{}, ler).Once()
	s.s3m.AssertNotCalled(s.T(), "DeleteObjectsWithContext", mock.Anything)

	s.hdr.CleanupChunks(s.ctx, s.key)
	s.s3m.AssertExpectations(s.T())
}

func (s *cleanupTestSuite) TestCleanupChunksDeleteError() {
	pfx := aws.String(fmt.Sprintf("chunks/%s/", s.key))
	bkt := aws.String(s.hdr.Env.AWSBucket)

	lst := []*s3.Object{
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_0.tar.gz", s.key))},
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_0.json", s.key))},
	}

	dbj := []string{
		fmt.Sprintf("chunks/%s/chunk_0.tar.gz", s.key),
		fmt.Sprintf("chunks/%s/chunk_0.json", s.key),
	}

	s.s3m.On("ListObjectsV2PagesWithContext", bkt, pfx).Return(lst, nil).Once()

	der := errors.New("delete objects failed")
	s.s3m.On("DeleteObjectsWithContext", dbj).Return(der).Once()

	s.hdr.CleanupChunks(s.ctx, s.key)
	s.s3m.AssertExpectations(s.T())
}

func (s *cleanupTestSuite) TestCleanupChunksFreeTierChunks() {
	pfx := aws.String(fmt.Sprintf("chunks/%s/", s.key))
	bkt := aws.String(s.hdr.Env.AWSBucket)

	lst := []*s3.Object{
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_0_%s.tar.gz", s.key, s.hdr.Env.FreeTierGroup))},
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_1_%s.json", s.key, s.hdr.Env.FreeTierGroup))},
		{Key: aws.String(fmt.Sprintf("chunks/%s/chunk_1.tar.gz", s.key))},
	}

	dbj := []string{
		fmt.Sprintf("chunks/%s/chunk_1.tar.gz", s.key),
	}

	s.s3m.On("ListObjectsV2PagesWithContext", bkt, pfx).Return(lst, nil).Once()

	s.s3m.On("DeleteObjectsWithContext", dbj).Return(nil).Once()

	s.hdr.CleanupChunks(s.ctx, s.key)

	s.s3m.AssertExpectations(s.T())
}

func TestCleanup(t *testing.T) {
	suite.Run(t, new(cleanupTestSuite))
}
