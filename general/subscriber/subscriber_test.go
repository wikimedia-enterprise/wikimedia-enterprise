package subscriber

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
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

type subscriberTestSuite struct {
	suite.Suite
	tps []string
	prr *kafka.Producer
	cns *kafka.Consumer
	ctx context.Context
	hdl Handler
	evs chan *Event
}

func (s *subscriberTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.tps = []string{"test"}
	s.evs = make(chan *Event, 100000)

	prr, err := kafka.NewProducer(&kafka.ConfigMap{})
	s.Assert().NoError(err)
	s.prr = prr

	cns, err := kafka.NewConsumer(&kafka.ConfigMap{
		"group.id": "test",
	})
	s.Assert().NoError(err)
	s.cns = cns

	s.hdl = func(ctx context.Context, msg *kafka.Message) error {
		return nil
	}
}

func (s *subscriberTestSuite) TestSubscribe() {
	sbr := Subscriber{
		Producer: s.prr,
		Consumer: s.cns,
	}

	go func() {
		time.Sleep(time.Second * 1)
		prc, err := os.FindProcess(os.Getpid())
		s.Assert().NoError(err)
		s.Assert().NoError(prc.Signal(os.Interrupt))
	}()

	pms := &Config{
		Topics: s.tps,
		Events: s.evs,
	}

	s.Assert().NoError(sbr.Subscribe(s.ctx, s.hdl, pms))
	s.Zero(len(s.evs))
}

func TestSubscriber(t *testing.T) {
	suite.Run(t, new(subscriberTestSuite))
}
