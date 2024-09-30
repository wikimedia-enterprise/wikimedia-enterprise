package export_test

import (
	"context"
	"crypto/md5" // #nosec G501
	"encoding/json"
	"errors"

	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/export"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	libkafka "wikimedia-enterprise/services/snapshots/libraries/kafka"
	"wikimedia-enterprise/services/snapshots/libraries/s3tracerproxy"

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

func (m *consumerMock) ReadAll(_ context.Context, snc int, tpc string, pts []int, cb func(msg *kafka.Message) error) error {
	ags := m.Called(snc, tpc, pts, cb)
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
	ags := m.Called(*uin.Key)

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
		Called(*pin.Key).
		Error(0)
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
	ctx context.Context
	tpc string
	req *pb.ExportRequest
	res *pb.ExportResponse
	hdr *export.Handler
	snc int
	pts []int
	sps [][]int
	stp []string
	ecr error
	eur error
	era error
	eho error
	epo error
	mss []*kafka.Message
	unm *mockUnmarshaler
}

func (s *handlerTestSuite) SetupSuite() {
	cfm := new(configMock)
	cfm.On("GetPartitions", s.req.GetProject(), int(s.req.GetNamespace())).Return(s.pts)
	cfm.On("GetStructuredProjects").Return(s.stp)
	csr := new(consumerMock)
	csr.On("Close").Return(nil)

	idr := fmt.Sprintf("%s_namespace_%d", s.req.Project, s.req.Namespace)

	pol := new(poolMock)
	upr := new(uploaderMock)
	s3m := new(s3Mock)

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
	for _, pts := range s.sps {
		csr.On("ReadAll", s.snc, s.tpc, pts, mock.Anything).Return(s.era).Run(func(args mock.Arguments) {
			callback := args.Get(3).(func(*kafka.Message) error)
			for _, msg := range s.mss {
				if err := callback(msg); err != nil {
					return
				}
			}
		})
	}

	pol.On("GetConsumer", fmt.Sprintf("_%s", idr)).
		Return(csr, s.ecr)

	upr.On("Upload", fmt.Sprintf("/%s.tar.gz", idr)).
		Return(s.eur)

	if s.req.EnableChunking {
		for i := 0; i < 20; i++ {
			s3m.On("PutObjectWithContext", fmt.Sprintf("chunks/%s/chunk_%d.tar.gz", idr, i)).
				Return(s.eur)
			s3m.On("PutObjectWithContext", fmt.Sprintf("chunks/%s/chunk_%d.json", idr, i)).
				Return(s.eur)
		}
	}

	s3m.On("HeadObjectWithContext", fmt.Sprintf("/%s.tar.gz", idr)).
		Return(s.eho)
	s3m.On("PutObjectWithContext", fmt.Sprintf("/%s.json", idr)).
		Return(s.epo)

	tps := new(schema.Topics)

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
		default:
			s.Assert().Fail("unexpected type Article in Unmarshal cbk")
		}
	})

	s.hdr = &export.Handler{
		Env: &env.Environment{
			Topics:         tps,
			PipeBufferSize: 1024,
			ChunkWorkers:   3,
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
			pts: []int{20, 40},
			stp: stp,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
			ecr: errors.New("could not get a consumer"),
		},
		{
			req: &pb.ExportRequest{
				Project:  "enwiki",
				Language: "en",
			},
			pts: []int{20, 40},
			sps: [][]int{{20}, {40}},
			stp: stp,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
			eur: errors.New("upload failed"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			pts: []int{20, 40},
			sps: [][]int{{20}, {40}},
			stp: stp,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
			era: errors.New("read all error"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			res: &pb.ExportResponse{},
			pts: []int{20, 40},
			sps: [][]int{{20}, {40}},
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
			eho: errors.New("head object error"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			res: &pb.ExportResponse{},
			pts: []int{20, 40},
			sps: [][]int{{20}, {40}},
			stp: stp,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
			epo: errors.New("put object error"),
		},
		{
			req: &pb.ExportRequest{
				Namespace: 0,
				Project:   "enwiki",
				Language:  "en",
			},
			res: &pb.ExportResponse{Total: 4},
			pts: []int{20, 40},
			sps: [][]int{{20}, {40}},
			stp: stp,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace:      0,
				Project:        "enwiki",
				Language:       "en",
				EnableChunking: true,
			},
			res: &pb.ExportResponse{Total: 4},
			pts: []int{20, 40},
			sps: [][]int{{20}, {40}},
			stp: stp,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
		},
		{
			req: &pb.ExportRequest{
				Namespace:      0,
				Project:        "dewiki",
				Language:       "de",
				Type:           "article",
				EnableChunking: true,
			},
			res: &pb.ExportResponse{Total: 8},
			pts: []int{20, 40, 60, 80},
			sps: [][]int{{20}, {40}, {60}, {80}},
			stp: stp,
			tpc: "aws.structured-data.dewiki-articles-compacted.v1",
		},
	} {
		suite.Run(t, testCase)
	}
}
