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
	"github.com/stretchr/testify/assert"
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

func (c *consumerMock) QueryWatermarkOffsets(tpc string, ptn int32, timeoutMs int) (int64, int64, error) {
	ags := c.Called(tpc, int(ptn))

	return int64(ags.Int(0)), int64(ags.Int(1)), ags.Error(2)
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
	ctx        context.Context
	csr        *libkafka.Consumer
	msg        *kafka.Message
	tpc        string
	ptn        int
	hof        int
	lof        int
	snc        int
	partition  int
	eof        error
	emd        error
	assignErr  error
	offsetsErr error
	readMsgErr error
	cer        error
}

func (s *consumerTestSuite) SetupSuite() {
	tps := []kafka.TopicPartition{}

	tpc := kafka.TopicPartition{
		Topic:     &s.tpc,
		Partition: int32(s.partition),
		Offset:    kafka.OffsetBeginning,
	}

	if s.snc > 0 {
		s.Assert().NoError(
			tpc.Offset.Set(s.snc),
		)
	}

	tps = append(tps, tpc)

	kmd := kafka.Metadata{
		Topics: map[string]kafka.TopicMetadata{
			s.tpc: {},
		},
	}
	cmk := new(consumerMock)
	cmk.On("Assign", tps).Return(s.assignErr)
	cmk.On("OffsetsForTimes", tps).Return(s.offsetsErr)
	cmk.On("ReadMessage").Return(s.msg, s.readMsgErr)
	cmk.On("GetMetadata", s.tpc).Return(&kmd, s.emd)
	cmk.On("GetWatermarkOffsets", s.tpc, s.ptn).Return(s.lof, s.hof, s.eof)
	cmk.On("Close").Return(s.cer)

	s.ctx = context.Background()
	s.csr = &libkafka.Consumer{
		Consumer: cmk,
	}
}

func (s *consumerTestSuite) TestReadAll() {
	err := s.csr.ReadAll(s.ctx, s.snc, s.tpc, s.partition, "idr", nil)

	if s.assignErr != nil {
		s.Assert().Equal(s.assignErr, err)
	} else if s.offsetsErr != nil {
		s.Assert().Equal(s.offsetsErr, err)
	} else if s.readMsgErr != nil && s.readMsgErr.Error() != kafka.ErrTimedOut.String() {
		s.Assert().Equal(s.readMsgErr, err)
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
			tpc:        "test.local",
			partition:  0,
			msg:        &kafka.Message{},
			lof:        5,
			hof:        10,
			readMsgErr: errors.New(kafka.ErrTimedOut.String()),
		},
		{
			tpc:        "test.local",
			partition:  1,
			msg:        &kafka.Message{},
			readMsgErr: errors.New(kafka.ErrTimedOut.String()),
			eof:        errors.New("offset error"),
		},
		{
			tpc:        "test.local",
			partition:  2,
			msg:        &kafka.Message{},
			emd:        errors.New("test"),
			readMsgErr: errors.New(kafka.ErrTimedOut.String()),
		},
		{
			tpc:        "test.local",
			partition:  3,
			msg:        &kafka.Message{},
			readMsgErr: errors.New(kafka.ErrTimedOut.String()),
			assignErr:  errors.New("test error"),
		},
		{
			tpc:        "test.local",
			partition:  4,
			msg:        &kafka.Message{},
			snc:        10,
			readMsgErr: errors.New(kafka.ErrTimedOut.String()),
			offsetsErr: errors.New("test error"),
		},
		{
			tpc:        "test.local",
			partition:  5,
			msg:        &kafka.Message{},
			readMsgErr: errors.New("test error"),
			cer:        errors.New("test error"),
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

type consumerOutput struct {
	msg *kafka.Message
	err error
}

type fakeConsumerPartition struct {
	outputs []consumerOutput

	// Consumer state
	nextIdx int
}

func (c *fakeConsumerPartition) ReadMessage() (*kafka.Message, error) {
	if c.nextIdx >= len(c.outputs) {
		return nil, errors.New(kafka.ErrTimedOut.String())
	}

	next := c.outputs[c.nextIdx]
	c.nextIdx++
	return next.msg, next.err
}

// fakeConsumer implements a fake Kafka round-robin consumer.
type fakeConsumer struct {
	partitions            []*fakeConsumerPartition
	lastPartitionConsumed int
}

func ptr(in string) *string {
	return &in
}

func (c *fakeConsumer) AddMsg(offset int) {
	c.AddMsgToPartition(0, offset)
}

func (c *fakeConsumer) AddMsgToPartition(partition int32, offset int) {
	for int(partition) >= len(c.partitions) {
		c.partitions = append(c.partitions, &fakeConsumerPartition{})
	}
	prt := c.partitions[partition]

	prt.outputs = append(prt.outputs, consumerOutput{msg: &kafka.Message{TopicPartition: kafka.TopicPartition{
		Topic:     ptr("test topic"),
		Partition: partition,
		Offset:    kafka.Offset(offset),
	}}})
}

func (c *fakeConsumer) PopLast(partition int32) {
	if int(partition) >= len(c.partitions) {
		return
	}
	prt := c.partitions[partition]

	prt.outputs = prt.outputs[:len(prt.outputs)-1]
}

func (c *fakeConsumer) AddError(err error) {
	c.AddErrorToPartition(0, err)
}

func (c *fakeConsumer) AddErrorToPartition(partition int32, err error) {
	for int(partition) >= len(c.partitions) {
		c.partitions = append(c.partitions, &fakeConsumerPartition{})
	}
	prt := c.partitions[partition]
	prt.outputs = append(prt.outputs, consumerOutput{err: err})

}

func (c *fakeConsumer) Assign(tps []kafka.TopicPartition) error {
	return nil
}

func (c *fakeConsumer) ReadMessage(_ time.Duration) (*kafka.Message, error) {
	if len(c.partitions) == 0 {
		return nil, errors.New(kafka.ErrTimedOut.String())
	}

	res, err := c.partitions[c.lastPartitionConsumed].ReadMessage()
	c.lastPartitionConsumed++
	if c.lastPartitionConsumed >= len(c.partitions) {
		c.lastPartitionConsumed = 0
	}

	return res, err
}

func (c *fakeConsumer) GetMetadata(tpc *string, _ bool, _ int) (*kafka.Metadata, error) {
	// Unimplemented.
	return nil, nil
}

func (c *fakeConsumer) QueryWatermarkOffsets(_ string, partition int32, _ int) (int64, int64, error) {
	if int(partition) >= len(c.partitions) {
		return int64(kafka.OffsetInvalid), 0, nil
	}

	prt := c.partitions[partition]
	inv := int64(kafka.OffsetInvalid)
	low := int64(kafka.OffsetInvalid)
	high := int64(kafka.OffsetInvalid)

	for _, msg := range prt.outputs {
		if msg.err != nil {
			continue
		}

		offset := msg.msg.TopicPartition.Offset
		if low == inv {
			low = int64(msg.msg.TopicPartition.Offset)
		}
		high = int64(offset) + 1
	}

	return low, high, nil
}

func (c *fakeConsumer) GetWatermarkOffsets(tpc string, ptn int32) (int64, int64, error) {
	// Unimplemented.
	return -1, -1, nil
}

func (c *fakeConsumer) OffsetsForTimes(tps []kafka.TopicPartition, time int) ([]kafka.TopicPartition, error) {
	// From librdkafka documentation:
	// The returned offset for each partition is the earliest offset whose timestamp is greater than or
	// equal to the given timestamp in the corresponding partition.
	res := []kafka.TopicPartition{}
	for _, tp := range tps {
		offset := kafka.OffsetEnd
		if int(tp.Partition) >= len(c.partitions) {
			offset = kafka.OffsetInvalid
		} else {
			prt := c.partitions[tp.Partition]
			for _, o := range prt.outputs {
				if o.msg != nil && int(o.msg.Timestamp.UnixMilli()) >= time {
					offset = o.msg.TopicPartition.Offset
					break
				}
			}
		}

		tp.Offset = offset
		res = append(res, tp)
	}
	return res, nil
}

func (c *fakeConsumer) Close() error {
	return nil
}

func TestTimeout(t *testing.T) {
	fake := fakeConsumer{}

	fake.AddMsg(0)
	fake.AddMsg(1)
	fake.AddMsg(2)
	fake.AddError(errors.New(kafka.ErrTimedOut.String()))
	fake.AddMsg(5)

	received := []int{}
	cbk := func(msg *kafka.Message) error {
		received = append(received, int(msg.TopicPartition.Offset))
		return nil
	}

	readCloser := libkafka.Consumer{
		Consumer: &fake,
	}

	err := readCloser.ReadAll(context.Background(), 0, "topic", 0, "enwiki_namespace_0", cbk)
	assert.NoError(t, err)

	// Should stop on timeout.
	assert.Equal(t, []int{0, 1, 2}, received)
}

func TestWatermark(t *testing.T) {
	fake := fakeConsumer{}

	fake.AddMsg(0)
	fake.AddMsg(1)
	fake.AddMsg(2)
	fake.AddError(errors.New(kafka.ErrTimedOut.String()))
	fake.AddMsg(5)
	fake.AddMsg(6)
	fake.AddMsg(7)

	received := []int{}
	cbk := func(msg *kafka.Message) error {
		received = append(received, int(msg.TopicPartition.Offset))
		return nil
	}

	readCloser := libkafka.Consumer{
		Consumer:     &fake,
		UseWatermark: true,
	}

	err := readCloser.ReadAll(context.Background(), 0, "topic", 0, "enwiki_namespace_0", cbk)
	assert.NoError(t, err)

	// Should retry the timeout and read to the end.
	assert.Equal(t, []int{0, 1, 2, 5, 6, 7}, received)
}

func TestPastWatermark(t *testing.T) {
	fake := fakeConsumer{}

	fake.AddMsg(0)
	fake.AddMsg(1)
	fake.AddMsg(2)
	fake.AddError(errors.New(kafka.ErrTimedOut.String()))
	fake.AddMsg(5)
	fake.AddMsg(6)
	fake.AddMsg(7)

	received := []int{}
	first := true
	cbk := func(msg *kafka.Message) error {
		if first {
			// After querying watermarks, new messages arrive and compaction deletes the previous last message.
			fake.PopLast(0)
			fake.AddMsg(8)
			fake.AddMsg(9)

			first = false
		}

		received = append(received, int(msg.TopicPartition.Offset))
		return nil
	}

	readCloser := libkafka.Consumer{
		Consumer:     &fake,
		UseWatermark: true,
	}

	err := readCloser.ReadAll(context.Background(), 0, "topic", 0, "enwiki_namespace_0", cbk)
	assert.NoError(t, err)

	// Should retry the timeout and not read past the previous watermark.
	assert.Equal(t, []int{0, 1, 2, 5, 6}, received)
}

func TestEmptyWithSince(t *testing.T) {
	fake := fakeConsumer{}

	fake.AddMsg(0)
	fake.AddMsg(1)
	fake.AddMsg(2)
	fake.AddError(errors.New(kafka.ErrTimedOut.String()))
	fake.AddMsg(5)
	fake.AddMsg(6)
	fake.AddMsg(7)

	called := false
	cbk := func(msg *kafka.Message) error {
		called = true
		return nil
	}

	readCloser := libkafka.Consumer{
		Consumer:     &fake,
		UseWatermark: true,
	}

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	err := readCloser.ReadAll(context.Background(), int(since), "topic", 0, "enwiki_namespace_0", cbk)
	assert.NoError(t, err)
	assert.False(t, called)
}
