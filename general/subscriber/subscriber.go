package subscriber

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/twmb/murmur3"
)

// DefaultNumberOfWorkers defines the default number of workers to use.
const DefaultNumberOfWorkers = 1

// DefaultMessagesChannelCap defines the default channel capacity for messages.
const DefaultMessagesChannelCap = 5

// DefaultReadTimeout defines the default read timeout for consumer.
const DefaultReadTimeout = time.Second * 1

// DefaultFlushTimeoutMs defines the default flush timeout for messages in milliseconds.
const DefaultFlushTimeoutMs = 10000

// Event defines a struct that represents and error event.
// Technically Error should never bi nil, but the Message can be.
// If the Message is is nil that means something happened before we
// have received the message.
type Event struct {
	Error   error
	Message *kafka.Message
}

// Config defines a struct that holds configuration parameters for the subscriber.
type Config struct {
	// Topics represents the Kafka topics to subscribe to.
	Topics []string

	// Events represents the channel on which to send error events.
	Events chan *Event

	// NumberOfWorkers represents the number of worker goroutines to use.
	NumberOfWorkers int

	// MessagesChannelCap represents the channel capacity for messages for each worker.
	// Note that total amount of capacity would be MessagesChannelCap multiplied by
	// the NumberOfWorkers. As each worker has it's own channel.
	MessagesChannelCap int

	// ReadTimeout represents the read timeout for kafka consumer.
	ReadTimeout time.Duration

	// FlushTimeoutMs represents the flush timeout for messages in milliseconds.
	// This is specifically exit flush, meaning time to clear the messages.
	FlushTimeoutMs int

	// Tracer is a function that injects the tracer into a Client.
	Tracer func(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context)
}

// CreateWorker creates a worker goroutine and returns a channel for sending messages to it.
func (c *Config) CreateWorker() chan *kafka.Message {
	return make(chan *kafka.Message, c.MessagesChannelCap)
}

// GetWorkerID returns the ID of the worker that should process the given message key.
func (s *Config) GetWorkerID(key []byte) (int, error) {
	hsh, err := murmur3.New32().Write(key)

	if err != nil {
		return 0, err
	}

	return hsh % s.NumberOfWorkers, nil
}

// PushEvent sends an event to the events channel. If the channel
// is nil, no events will be sent.
func (c *Config) PushEvent(msg *kafka.Message, err error) {
	if c.Events != nil {
		c.Events <- &Event{
			Error:   err,
			Message: msg,
		}
	}
}

// Handler is a function that processes a Kafka message.
type Handler func(ctx context.Context, msg *kafka.Message) error

type ConsumerInterface interface {
	SubscribeTopics(topics []string, config kafka.RebalanceCb) error
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	Close() error
}

type ProducerInterface interface {
	Flush(timeoutMs int) int
	Close()
}

// Subscriber is a struct used to subscribe to Kafka topics.
// This struct uses dependency injection to find consumer and producer.
type SubscriberInterface interface {
	Subscribe(ctx context.Context, hdl Handler, cfg *Config) error
}

// New creates a new Subscriber instance.
func New(consumer *kafka.Consumer, producer *kafka.Producer) *Subscriber {
	return &Subscriber{
		Consumer: consumer,
		Producer: producer,
	}
}

type Subscriber struct {
	Consumer ConsumerInterface
	Producer ProducerInterface `optional:"true"`
}

// Subscribe subscribes to Kafka topics and handles incoming messages.
// Also it implements graceful shutdown to make sure there are no messages lost.
func (s *Subscriber) Subscribe(ctx context.Context, hdl Handler, cfg *Config) error {
	if cfg.NumberOfWorkers <= 0 {
		cfg.NumberOfWorkers = DefaultNumberOfWorkers
	}

	if cfg.MessagesChannelCap <= 0 {
		cfg.MessagesChannelCap = DefaultMessagesChannelCap
	}

	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = DefaultReadTimeout
	}

	if cfg.FlushTimeoutMs <= 0 {
		cfg.FlushTimeoutMs = DefaultFlushTimeoutMs
	}

	if err := s.Consumer.SubscribeTopics(cfg.Topics, nil); err != nil {
		return err
	}

	// Create a channel to receive OS signals.
	sgn := make(chan os.Signal, 1)
	signal.Notify(sgn, os.Interrupt, syscall.SIGTERM)

	// Create a map to store worker channels and a WaitGroup to synchronize them.
	wks := map[int]chan *kafka.Message{}
	swg := new(sync.WaitGroup)
	swg.Add(cfg.NumberOfWorkers)

	// Create and start worker goroutines.
	for i := 0; i < cfg.NumberOfWorkers; i++ {
		wks[i] = cfg.CreateWorker()

		go func(mgs chan *kafka.Message) {
			defer swg.Done()

			for msg := range mgs {
				// Tracer is enabled, so we need to create a new context.
				if cfg.Tracer != nil {
					end, trx := cfg.Tracer(ctx, nil)
					if err := hdl(trx, msg); err != nil {
						cfg.PushEvent(msg, err)
						end(err, "error processing message")
					} else {
						end(nil, "message processed")
					}

					continue
				}

				// Tracer is disabled
				if err := hdl(ctx, msg); err != nil {
					cfg.PushEvent(msg, err)

				}

			}
		}(wks[i])
	}

	// Loop indefinitely, waiting for signals or messages.
	for {
		select {
		case <-sgn: // If a signal is received, clean up and return.
			if err := s.Consumer.Close(); err != nil {
				return err
			}

			// Close the go routines for all the workers.
			for i := 0; i < cfg.NumberOfWorkers; i++ {
				close(wks[i])
			}

			swg.Wait() // Wait for all workers to complete before returning.

			if s.Producer != nil {
				defer s.Producer.Close()

				for {
					if s.Producer.Flush(cfg.FlushTimeoutMs) == 0 {
						return nil
					}
				}
			}

			return nil
		default: // If no signals are received, try to read a message from Kafka.
			msg, err := s.Consumer.ReadMessage(cfg.ReadTimeout)

			if err != nil {
				// If there's an error that's not a timeout, push it to the event channel.
				if err.Error() != kafka.ErrTimedOut.String() {
					cfg.PushEvent(nil, err)
				}
				continue
			}

			// Get the worker ID for the message key.
			wid, err := cfg.GetWorkerID(msg.Key)

			if err != nil {
				cfg.PushEvent(msg, err)
				continue
			}

			// Send the message to the appropriate worker channel
			wks[wid] <- msg
		}
	}
}
