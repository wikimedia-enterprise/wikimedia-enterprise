package export_test

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"testing"
	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers/export"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	libkafka "wikimedia-enterprise/services/snapshots/libraries/kafka"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
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

func (m *consumerMock) ReadAll(_ context.Context, snc int, tpc string, pts []int, _ func(msg *kafka.Message) error) error {
	ags := m.Called(snc, tpc, pts)
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

	_, _ = io.ReadAll(uin.Body)

	return nil, ags.Error(0)
}

type s3Mock struct {
	mock.Mock
	s3iface.S3API
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
	ecr error
	eur error
	era error
	eho error
	epo error
}

func (s *handlerTestSuite) SetupSuite() {
	cfm := new(configMock)
	cfm.On("GetPartitions", s.req.GetProject(), int(s.req.GetNamespace())).Return(s.pts)

	csr := new(consumerMock)
	csr.On("Close").Return(nil)

	for _, pts := range s.sps {
		csr.On("ReadAll", s.snc, s.tpc, pts).Return(s.era)
	}

	idr := fmt.Sprintf("%s_namespace_%d", s.req.Project, s.req.Namespace)
	pol := new(poolMock)
	pol.On("GetConsumer", fmt.Sprintf("_%s", idr)).
		Return(csr, s.ecr)

	upr := new(uploaderMock)
	upr.On("Upload", fmt.Sprintf("/%s.tar.gz", idr)).
		Return(s.eur)

	s3m := new(s3Mock)
	s3m.On("HeadObjectWithContext", fmt.Sprintf("/%s.tar.gz", idr)).
		Return(s.eho)
	s3m.On("PutObjectWithContext", fmt.Sprintf("/%s.json", idr)).
		Return(s.epo)

	tps := new(schema.Topics)
	s.Assert().NoError(
		tps.UnmarshalEnvironmentValue(""),
	)

	s.hdr = &export.Handler{
		Env: &env.Environment{
			Topics:         tps,
			PipeBufferSize: 1024,
		},
		Cfg:      cfm,
		Uploader: upr,
		Pool:     pol,
		S3:       s3m,
	}
}

func (s *handlerTestSuite) TestExport() {
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
	for _, testCase := range []*handlerTestSuite{
		{
			req: &pb.ExportRequest{
				Project:  "enwiki",
				Language: "en",
			},
			pts: []int{20, 40},
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
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
			epo: errors.New("put object error"),
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
		},
	} {
		suite.Run(t, testCase)
	}
}
