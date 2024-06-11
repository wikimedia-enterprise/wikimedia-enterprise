package subscriber

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
	cfg *Config
	key string
}

func (s *configTestSuite) TestCreateWorker() {
	s.Assert().Equal(cap(s.cfg.CreateWorker()), s.cfg.MessagesChannelCap)
}

func (s *configTestSuite) TestGetWorkerID() {
	idr, err := s.cfg.GetWorkerID([]byte(s.key))
	s.Assert().NoError(err)
	s.Assert().IsType(99, idr)
}

func (s *configTestSuite) TestPushEvent() {
	msg := new(kafka.Message)
	err := errors.New("test")

	s.cfg.PushEvent(msg, err)
	res := <-s.cfg.Events

	s.Assert().Equal(err, res.Error)
	s.Assert().Equal(msg, res.Message)
}

func TestConfig(t *testing.T) {
	for _, testCase := range []*configTestSuite{
		{
			key: "test",
			cfg: &Config{
				MessagesChannelCap: 10,
				NumberOfWorkers:    10,
				Events:             make(chan *Event, 1),
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) SubscribeTopics(topics []string, handle kafka.RebalanceCb) error {
	args := m.Called(topics, handle)
	return args.Error(0)
}

func (m *MockConsumer) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	args := m.Called(timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*kafka.Message), nil
}

func (m *MockConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Flush(timeoutMs int) int {
	args := m.Called(timeoutMs)
	return args.Int(0)
}

func (m *MockProducer) Close() {
	m.Called()
}

type SubscriberTestSuite struct {
	suite.Suite
	mockConsumer *MockConsumer
	mockProducer *MockProducer
	cfg          *Config
	subscriber   *Subscriber
	handler      Handler
}

func (s *SubscriberTestSuite) SetupTest() {
	s.mockConsumer = new(MockConsumer)
	s.mockProducer = new(MockProducer)

	s.cfg = &Config{
		Topics:             []string{"test-topic"},
		Events:             make(chan *Event),
		NumberOfWorkers:    1,
		MessagesChannelCap: 1,
		ReadTimeout:        5000 * time.Millisecond,
		FlushTimeoutMs:     200,
	}

	s.handler = func(ctx context.Context, msg *kafka.Message) error {
		return nil
	}

	s.subscriber = &Subscriber{
		Consumer: s.mockConsumer,
		Producer: s.mockProducer,
	}

	s.mockConsumer.On("SubscribeTopics", s.cfg.Topics, mock.Anything).Return(nil)
}

func (s *SubscriberTestSuite) TearDownTest() {
	s.mockConsumer.ExpectedCalls = nil
	s.mockProducer.ExpectedCalls = nil
}

func (s *SubscriberTestSuite) TestNew() {
	s.Assert().NotNil(New(new(kafka.Consumer), new(kafka.Producer)))
	s.Assert().NotNil(New(new(kafka.Consumer), nil))
}

func (s *SubscriberTestSuite) TestSubscribeNoTracer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &s.cfg.Topics[0], Partition: 0},
		Key:            []byte("key"),
		Value:          []byte("value"),
	}, nil).Once()
	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafka.NewError(kafka.ErrTimedOut, "timeout", false)).Once()

	go func() {
		err := s.subscriber.Subscribe(ctx, s.handler, s.cfg)
		assert.NoError(s.T(), err)
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		process, _ := os.FindProcess(os.Getpid())
		err := process.Signal(syscall.SIGTERM)

		assert.NoError(s.T(), err)
	}()

	// Give some time for cleanup
	time.Sleep(200 * time.Millisecond)

	s.mockConsumer.AssertExpectations(s.T())
	s.mockProducer.AssertExpectations(s.T())
}

func (s *SubscriberTestSuite) TestSubscribeWithTracer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tracerCalled := false
	endCalled := false

	s.cfg.Tracer = func(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context) {
		tracerCalled = true
		return func(err error, msg string) {
			endCalled = true
			if err == nil {
				assert.Equal(s.T(), "message processed", msg)
			} else {
				assert.Equal(s.T(), "error processing message", msg)
			}
		}, ctx
	}

	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &s.cfg.Topics[0], Partition: 0},
		Key:            []byte("key"),
		Value:          []byte("value"),
	}, nil).Once()
	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafka.NewError(kafka.ErrTimedOut, "timeout", false)).Maybe()

	go func() {
		err := s.subscriber.Subscribe(ctx, s.handler, s.cfg)
		assert.NoError(s.T(), err)
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		process, _ := os.FindProcess(os.Getpid())
		err := process.Signal(syscall.SIGTERM)

		if err != nil {
			assert.NoError(s.T(), err)
		}
	}()

	// Give some time for cleanup
	time.Sleep(200 * time.Millisecond)
	assert.True(s.T(), tracerCalled)
	assert.True(s.T(), endCalled)

	s.mockConsumer.AssertExpectations(s.T())
	s.mockProducer.AssertExpectations(s.T())
}

func (s *SubscriberTestSuite) TestSubscribeTracerHandlerSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tracerCalled := false
	endCalled := false

	s.cfg.Tracer = func(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context) {
		tracerCalled = true
		return func(err error, msg string) {
			endCalled = true
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "message processed", msg)
		}, ctx
	}

	s.handler = func(ctx context.Context, msg *kafka.Message) error {
		return nil
	}

	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &s.cfg.Topics[0], Partition: 0},
		Key:            []byte("key"),
		Value:          []byte("value"),
	}, nil).Once()
	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafka.NewError(kafka.ErrTimedOut, "timeout", false)).Maybe()

	go func() {
		err := s.subscriber.Subscribe(ctx, s.handler, s.cfg)
		assert.NoError(s.T(), err)
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		process, _ := os.FindProcess(os.Getpid())
		err := process.Signal(syscall.SIGTERM)

		if err != nil {
			assert.NoError(s.T(), err)
		}
	}()

	// Give some time for cleanup
	time.Sleep(200 * time.Millisecond)

	assert.True(s.T(), tracerCalled)
	assert.True(s.T(), endCalled)

	s.mockConsumer.AssertExpectations(s.T())
	s.mockProducer.AssertExpectations(s.T())
}

func (s *SubscriberTestSuite) TestSubscribeTracerHandlerError() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tracerCalled := false
	endCalled := false

	s.cfg.Tracer = func(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context) {
		tracerCalled = true
		return func(err error, msg string) {
			endCalled = true
			assert.NotNil(s.T(), err)
			assert.Equal(s.T(), "error processing message", msg)
		}, ctx
	}

	s.handler = func(ctx context.Context, msg *kafka.Message) error {
		return errors.New("handler error")
	}

	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &s.cfg.Topics[0], Partition: 0},
		Key:            []byte("key"),
		Value:          []byte("value"),
	}, nil).Once()
	s.mockConsumer.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafka.NewError(kafka.ErrTimedOut, "timeout", false)).Maybe()

	go func() {
		err := s.subscriber.Subscribe(ctx, s.handler, s.cfg)
		assert.NoError(s.T(), err)
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		process, _ := os.FindProcess(os.Getpid())
		err := process.Signal(syscall.SIGTERM)

		if err != nil {
			assert.NoError(s.T(), err)
		}
	}()

	// Give some time for cleanup
	time.Sleep(200 * time.Millisecond)

	assert.True(s.T(), tracerCalled)
	assert.False(s.T(), endCalled)

	s.mockConsumer.AssertExpectations(s.T())
	s.mockProducer.AssertExpectations(s.T())
}

func TestSubscriberTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriberTestSuite))
}
