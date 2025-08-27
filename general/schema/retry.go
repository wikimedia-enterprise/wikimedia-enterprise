package schema

import (
	"context"
	"errors"
	"reflect"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
)

// Set of errors for retry method inside retryer.
var (
	ErrEmptyEventField       = errors.New("empty event field")
	ErrCantSetFailCount      = errors.New("can't set FailCount")
	ErrCantSetFailReason     = errors.New("can't set FailReason")
	ErrConfigNotSet          = errors.New("config not set")
	ErrTopicErrorNotSet      = errors.New("topic for error messages not set")
	ErrTopicDeadLetterNotSet = errors.New("topic for dead letter messages not set")
	ErrErrNotSet             = errors.New("error message not set")
)

// RetryMessage is a message that will be sent to retry topic.
type RetryMessage struct {
	Config          *Config
	TopicError      string
	TopicDeadLetter string
	MaxFailCount    int
	Message         *kafka.Message
	Error           error
}

// Retryer interface to represent retry method that will put the message into retry queue.
type Retryer interface {
	Retry(ctx context.Context, rms *RetryMessage) error
}

// NewRetry creates new instance fo retry struct.
func NewRetry(str UnmarshalProducer) Retryer {
	return &Retry{
		Stream: str,
	}
}

// Retry struct that helps with enqueueing messages for retry.
type Retry struct {
	dig.In
	Stream UnmarshalProducer
}

// Retry sends a message to retry queue.
func (r *Retry) Retry(ctx context.Context, rms *RetryMessage) error {
	if rms.Config == nil {
		return ErrConfigNotSet
	}

	if len(rms.TopicError) == 0 {
		return ErrTopicErrorNotSet
	}

	if len(rms.TopicDeadLetter) == 0 {
		return ErrTopicDeadLetterNotSet
	}

	if rms.Error == nil {
		return ErrErrNotSet
	}

	key := new(Key)

	if err := r.Stream.Unmarshal(ctx, rms.Message.Key, key); err != nil {
		return err
	}

	val := reflect.
		New(reflect.TypeOf(rms.Config.Reflection))

	if err := r.Stream.Unmarshal(ctx, rms.Message.Value, val.Interface()); err != nil {
		return err
	}

	evt := val.Elem().FieldByName("Event")

	if evt.IsNil() {
		return ErrEmptyEventField
	}

	fct := evt.Elem().FieldByName("FailCount")

	if !fct.CanSet() {
		return ErrCantSetFailCount
	}

	fcr := evt.Elem().FieldByName("FailReason")

	if !fct.CanSet() {
		return ErrCantSetFailReason
	}

	fct.SetInt(fct.Int() + 1)
	fcr.SetString(rms.Error.Error())

	tpc := rms.TopicError

	if fct.Int() > int64(rms.MaxFailCount) {
		tpc = rms.TopicDeadLetter
	}

	msg := &Message{
		Config: rms.Config,
		Topic:  tpc,
		Value:  val.Interface(),
		Key:    key,
	}

	return r.Stream.Produce(ctx, msg)
}
