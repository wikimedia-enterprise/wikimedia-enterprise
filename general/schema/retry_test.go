package schema

import (
	"context"
	"errors"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type retryUnmarshalProducerMock struct {
	UnmarshalProducer
	mock.Mock
}

func (m *retryUnmarshalProducerMock) Unmarshal(_ context.Context, dta []byte, val interface{}) error {
	ags := m.Called(dta)

	switch val := val.(type) {
	case *Key:
		*val = *ags.Get(0).(*Key)
	case *Article:
		*val = *ags.Get(0).(*Article)
	}

	return ags.Error(1)
}

func (m *retryUnmarshalProducerMock) Produce(_ context.Context, mgs ...*Message) error {
	return m.Called(len(mgs)).Error(0)
}

type retryTestSuite struct {
	suite.Suite
	ctx context.Context
	rms *RetryMessage
	rtr *Retry
	key *Key
	val *Article
	euk error
	euv error
	epr error
	ert error
	fct int
}

func (s *retryTestSuite) SetupSuite() {
	str := new(retryUnmarshalProducerMock)
	str.On("Unmarshal", s.rms.Message.Key).Return(s.key, s.euk)
	str.On("Unmarshal", s.rms.Message.Value).Return(s.val, s.euv)
	str.On("Produce", 1).Return(s.epr)

	s.rtr = &Retry{
		Stream: str,
	}
}

func (s *retryTestSuite) TestRetry() {
	err := s.rtr.Retry(s.ctx, s.rms)

	if s.ert != nil {
		s.Assert().Equal(s.ert, err)
	} else if s.euk != nil {
		s.Assert().Equal(s.euk, err)
	} else if s.euv != nil {
		s.Assert().Equal(s.euv, err)
	} else if s.epr != nil {
		s.Assert().Equal(s.epr, err)
	} else {
		s.Assert().NoError(err)
		s.Assert().Equal(s.fct, s.val.Event.FailCount)
		s.Assert().Equal(s.rms.Error.Error(), s.val.Event.FailReason)
	}
}

func TestRetry(t *testing.T) {
	for _, testCase := range []*retryTestSuite{
		{
			rms: &RetryMessage{
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			ert: ErrConfigNotSet,
		},
		{
			rms: &RetryMessage{
				Config: ConfigArticle,
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			ert: ErrTopicErrorNotSet,
		},
		{
			rms: &RetryMessage{
				Config:     ConfigArticle,
				TopicError: "local.error.v1",
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			ert: ErrTopicDeadLetterNotSet,
		},
		{
			rms: &RetryMessage{
				Config:          ConfigArticle,
				TopicError:      "local.error.v1",
				TopicDeadLetter: "local.dead-letter.v1",
				MaxFailCount:    1,
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			ert: ErrErrNotSet,
		},
		{
			key: &Key{
				Identifier: "/enwiki/Earth",
			},
			rms: &RetryMessage{
				Config:          ConfigArticle,
				TopicError:      "local.error.v1",
				TopicDeadLetter: "local.dead-letter.v1",
				MaxFailCount:    1,
				Error:           errors.New("test error"),
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			euk: errors.New("can't unmarshal key"),
		},
		{
			key: &Key{
				Identifier: "/enwiki/Earth",
			},
			val: &Article{
				Name:  "Earth",
				Event: &Event{},
			},
			rms: &RetryMessage{
				Config:          ConfigArticle,
				TopicError:      "local.error.v1",
				TopicDeadLetter: "local.dead-letter.v1",
				MaxFailCount:    1,
				Error:           errors.New("test error"),
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			euv: errors.New("can't unmarshal val"),
		},
		{
			key: &Key{
				Identifier: "/enwiki/Earth",
			},
			val: &Article{
				Name:  "Earth",
				Event: &Event{},
			},
			rms: &RetryMessage{
				Config:          ConfigArticle,
				TopicError:      "local.error.v1",
				TopicDeadLetter: "local.dead-letter.v1",
				MaxFailCount:    1,
				Error:           errors.New("test error"),
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			epr: errors.New("can't produce msg"),
		},
		{
			key: &Key{
				Identifier: "/enwiki/Earth",
			},
			val: &Article{
				Name:  "Earth",
				Event: &Event{},
			},
			rms: &RetryMessage{
				Config:          ConfigArticle,
				TopicError:      "local.error.v1",
				TopicDeadLetter: "local.dead-letter.v1",
				MaxFailCount:    1,
				Error:           errors.New("test error"),
				Message: &kafka.Message{
					Key:   []byte("key"),
					Value: []byte("val"),
				},
			},
			fct: 1,
		},
	} {
		suite.Run(t, testCase)
	}
}
