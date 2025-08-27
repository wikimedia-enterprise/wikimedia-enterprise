package kafka_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	libkafka "wikimedia-enterprise/api/realtime/libraries/kafka"

	"github.com/avast/retry-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type consumerMock struct {
	mock.Mock
}

func (c *consumerMock) Assign(_ []kafka.TopicPartition) error {
	return c.Called().Error(0)
}

func (c *consumerMock) ReadMessage(_ time.Duration) (*kafka.Message, error) {
	ags := c.Called()

	return ags.Get(0).(*kafka.Message), ags.Error(1)
}

func (c *consumerMock) OffsetsForTimes(tps []kafka.TopicPartition, timeOutMS int) ([]kafka.TopicPartition, error) {
	ags := c.Called(tps, timeOutMS)

	return ags.Get(0).([]kafka.TopicPartition), ags.Error(1)
}

func (c *consumerMock) Close() error {
	return c.Called().Error(0)
}

func (c *consumerMock) Poll(_ int) kafka.Event {
	return c.Called().Get(0).(kafka.Event)
}

type readAllTestSuite struct {
	suite.Suite
	msg       any
	topic     string
	partition int32
	callback  func(msg *kafka.Message) error
	assignErr error
	cbkErr    error
}

func (s *readAllTestSuite) TestReadAll() {
	con := new(consumerMock)
	con.On("Assign", mock.Anything).Return(s.assignErr)

	con.On("Poll").Return(s.msg).Once()

	c := &libkafka.Consumer{Consumer: con}
	params := &libkafka.ReadParams{}

	err := c.ReadAll(context.Background(), params, s.topic, []int{int(s.partition)}, s.callback)
	if s.assignErr != nil {
		s.Assert().Contains(err.Error(), s.assignErr.Error())
	} else if s.cbkErr != nil {
		s.Assert().Equal(s.cbkErr.Error(), err.Error())
	}
}

func TestReadAll(t *testing.T) {
	testTopic := "local"

	for _, testCase := range []*readAllTestSuite{
		{
			assignErr: errors.New("assign error"),
			callback: func(m *kafka.Message) error {
				return nil
			},
			msg: &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic: &testTopic, Partition: 0, Offset: kafka.OffsetEnd,
				},
				Value: []byte("val"),
				Key:   []byte("key"),
			},
			topic:     "local",
			partition: 0,
		},
		{
			cbkErr: errors.New("callback error"),
			callback: func(m *kafka.Message) error {
				return errors.New("callback error")
			},
			msg: &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic: &testTopic, Partition: 0, Offset: kafka.OffsetEnd,
				},
				Value: []byte("val"),
				Key:   []byte("key"),
			},
			topic:     "local",
			partition: 0,
		},
	} {
		suite.Run(t, testCase)
	}
}

type setPartitionOffsetTestSuite struct {
	suite.Suite
	params         *libkafka.ReadParams
	topic          string
	partitions     []int
	topicPartition []kafka.TopicPartition
	offsetErr      error
}

func (s *setPartitionOffsetTestSuite) TestSetPartitionOffset() {
	retry.DefaultAttempts = 1
	retry.DefaultDelay = 0
	con := new(consumerMock)
	con.On("OffsetsForTimes", mock.Anything, mock.Anything).Return(s.topicPartition, s.offsetErr)

	c := &libkafka.Consumer{Consumer: con}

	out, err := c.SetPartitionOffsets(s.topic, s.partitions, s.params)

	if s.offsetErr != nil {
		s.Assert().Equal(err, s.offsetErr)
		s.Assert().Nil(out)
	} else {
		s.Assert().NoError(err)
		s.Assert().Equal(out, s.topicPartition)
	}
}

func TestSetPartitionOffset(t *testing.T) {
	testTopic := "local"

	for _, testCase := range []*setPartitionOffsetTestSuite{
		{
			offsetErr:      errors.New("offset error"),
			topicPartition: []kafka.TopicPartition{},
			topic:          "local",
			partitions:     []int{0},
			params: &libkafka.ReadParams{
				Since: time.Now(),
			},
		},
		{
			topicPartition: []kafka.TopicPartition{
				{
					Topic:     &testTopic,
					Partition: 0,
					Offset:    0,
				},
			},
			topic:      "local",
			partitions: []int{0},
			params: &libkafka.ReadParams{
				Since: time.Now(),
			},
		},
		{
			topicPartition: []kafka.TopicPartition{
				{
					Topic:     &testTopic,
					Partition: 0,
					Offset:    0,
				},
			},
			topic:      "local",
			partitions: []int{0},
			params: &libkafka.ReadParams{
				SincePerPartition: map[int]time.Time{0: time.Now()},
			},
		},
		{
			topicPartition: []kafka.TopicPartition{
				{
					Topic:     &testTopic,
					Partition: 0,
					Offset:    0,
				},
			},
			topic:      "local",
			partitions: []int{0},
			params: &libkafka.ReadParams{
				OffsetPerPartition: map[int]int64{0: 0},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type closeTestSuite struct {
	suite.Suite
	csr  *libkafka.Consumer
	cErr error
}

func (s *closeTestSuite) TestClose() {
	con := new(consumerMock)
	con.On("Close").Return(s.cErr)

	s.csr = &libkafka.Consumer{Consumer: con}

	err := s.csr.Close()

	s.Assert().Equal(s.cErr, err)
}

func TestClose(t *testing.T) {

	for _, testCase := range []*closeTestSuite{
		{
			cErr: errors.New("close error"),
		},
		{},
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
