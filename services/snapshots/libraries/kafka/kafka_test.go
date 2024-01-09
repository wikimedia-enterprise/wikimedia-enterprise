package kafka_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/services/snapshots/config/env"
	libkafka "wikimedia-enterprise/services/snapshots/libraries/kafka"

	"github.com/avast/retry-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type consumerMock struct {
	mock.Mock
}

func (c *consumerMock) Assign(tps []kafka.TopicPartition) error {
	return c.Called(tps).Error(0)
}

func (c *consumerMock) ReadMessage(_ time.Duration) (*kafka.Message, error) {
	ags := c.Called()

	return ags.Get(0).(*kafka.Message), ags.Error(1)
}

func (c *consumerMock) GetMetadata(tpc *string, _ bool, _ int) (*kafka.Metadata, error) {
	ags := c.Called(*tpc)

	return ags.Get(0).(*kafka.Metadata), ags.Error(1)
}

func (c *consumerMock) GetWatermarkOffsets(tpc string, ptn int32) (int64, int64, error) {
	ags := c.Called(tpc, int(ptn))

	return int64(ags.Int(0)), int64(ags.Int(1)), ags.Error(2)
}

func (c *consumerMock) OffsetsForTimes(tps []kafka.TopicPartition, _ int) ([]kafka.TopicPartition, error) {
	ags := c.Called(tps)

	return tps, ags.Error(0)
}

func (c *consumerMock) Close() error {
	return c.Called().Error(0)
}

type consumerTestSuite struct {
	suite.Suite
	ctx context.Context
	csr *libkafka.Consumer
	msg *kafka.Message
	tpc string
	ptn int
	hof int
	lof int
	snc int
	pts []int
	eof error
	emd error
	aer error
	oer error
	rer error
	cer error
}

func (s *consumerTestSuite) SetupSuite() {
	tps := []kafka.TopicPartition{}

	for _, ptn := range s.pts {
		tpc := kafka.TopicPartition{
			Topic:     &s.tpc,
			Partition: int32(ptn),
			Offset:    kafka.OffsetBeginning,
		}

		if s.snc > 0 {
			s.Assert().NoError(
				tpc.Offset.Set(s.snc),
			)
		}

		tps = append(tps, tpc)
	}

	kmd := kafka.Metadata{
		Topics: map[string]kafka.TopicMetadata{
			s.tpc: {},
		},
	}
	cmk := new(consumerMock)
	cmk.On("Assign", tps).Return(s.aer)
	cmk.On("OffsetsForTimes", tps).Return(s.oer)
	cmk.On("ReadMessage").Return(s.msg, s.rer)
	cmk.On("GetMetadata", s.tpc).Return(&kmd, s.emd)
	cmk.On("GetWatermarkOffsets", s.tpc, s.ptn).Return(s.lof, s.hof, s.eof)
	cmk.On("Close").Return(s.cer)

	s.ctx = context.Background()
	s.csr = &libkafka.Consumer{
		Consumer: cmk,
	}
}

func (s *consumerTestSuite) TestReadAll() {
	err := s.csr.ReadAll(s.ctx, s.snc, s.tpc, s.pts, nil)

	if s.aer != nil {
		s.Assert().Equal(s.aer, err)
	} else if s.oer != nil {
		s.Assert().Equal(s.oer, err)
	} else if s.rer != nil && s.rer.Error() != kafka.ErrTimedOut.String() {
		s.Assert().Equal(s.rer, err)
	} else {
		s.Assert().NoError(err)
	}
}

func (s *consumerTestSuite) TestGetMetadata() {
	mtd, err := s.csr.GetMetadata(s.tpc)

	if s.emd != nil {
		s.Assert().Equal(s.emd, err)
		s.Assert().Nil(mtd)
	} else {
		s.Assert().NoError(err)
		s.Assert().NotNil(mtd)
	}
}

func (s *consumerTestSuite) TestQueryWatermarkOffsets() {
	lof, hof, err := s.csr.GetWatermarkOffsets(s.tpc, s.ptn)

	s.Assert().Equal(s.lof, lof)
	s.Assert().Equal(s.hof, hof)
	s.Assert().Equal(s.eof, err)
}

func (s *consumerTestSuite) TestClose() {
	err := s.csr.Close()

	s.Assert().Equal(s.cer, err)
}

func TestConsumer(t *testing.T) {
	retry.DefaultAttempts = 1

	for _, testCase := range []*consumerTestSuite{
		{
			tpc: "test.local",
			pts: []int{0, 1, 2, 3, 4, 5},
			msg: &kafka.Message{},
			lof: 5,
			hof: 10,
			rer: errors.New(kafka.ErrTimedOut.String()),
		},
		{
			tpc: "test.local",
			pts: []int{0, 1, 2, 3, 4, 5},
			msg: &kafka.Message{},
			rer: errors.New(kafka.ErrTimedOut.String()),
			eof: errors.New("offset error"),
		},
		{
			tpc: "test.local",
			pts: []int{0, 1, 2, 3, 4, 5},
			msg: &kafka.Message{},
			emd: errors.New("test"),
			rer: errors.New(kafka.ErrTimedOut.String()),
		},
		{
			tpc: "test.local",
			pts: []int{0, 1, 2, 3, 4, 5},
			msg: &kafka.Message{},
			rer: errors.New(kafka.ErrTimedOut.String()),
			aer: errors.New("test error"),
		},
		{
			tpc: "test.local",
			pts: []int{0, 1, 2, 3, 4, 5},
			msg: &kafka.Message{},
			snc: 10,
			rer: errors.New(kafka.ErrTimedOut.String()),
			oer: errors.New("test error"),
		},
		{
			tpc: "test.local",
			pts: []int{0, 1, 2, 3, 4, 5},
			msg: &kafka.Message{},
			rer: errors.New("test error"),
			cer: errors.New("test error"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type poolTestSuite struct {
	suite.Suite
	pol *libkafka.Pool
	gid string
	cfg *kafka.ConfigMap
}

func (s *poolTestSuite) SetupSuite() {
	s.pol = &libkafka.Pool{
		Config: s.cfg,
	}
}

func (s *poolTestSuite) TestGetConsumer() {
	csr, err := s.pol.GetConsumer(s.gid)

	s.Assert().NoError(err)
	s.Assert().NotNil(csr)
}

func TestPool(t *testing.T) {
	for _, testCase := range []*poolTestSuite{
		{
			gid: "unique",
			cfg: &kafka.ConfigMap{},
		},
	} {
		suite.Run(t, testCase)
	}
}

type newPoolTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *newPoolTestSuite) TestNewPool() {
	cgr := libkafka.NewPool(s.env)

	s.Assert().NotNil(cgr)
}

func TestNewPool(t *testing.T) {
	for _, testCase := range []*newPoolTestSuite{
		{
			env: &env.Environment{},
		},
		{
			env: &env.Environment{
				KafkaCreds: &env.Credentials{
					Username: "user",
					Password: "password",
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
