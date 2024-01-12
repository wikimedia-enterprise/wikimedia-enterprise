package schema

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"reflect"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// MockTopic representation of mock stream with configuration.
type MockTopic struct {
	Topic      string
	Partitions []int32
	Config     *Config
	Type       interface{}
	Reader     io.Reader
}

// NewMock creates new mocker instance to read topics and produce local mocks.
func NewMock() (*Mock, error) {
	prd, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"compression.type":       "gzip",
		"go.batch.producer":      true,
		"queue.buffering.max.ms": 10,
	})

	if err != nil {
		return nil, err
	}

	reg := NewRegistry(os.Getenv("SCHEMA_REGISTRY_URL"))
	mck := &Mock{
		Producer: NewHelper(reg, prd),
	}

	return mck, nil
}

// Mock helper that provides the ability to easily produce mock data.
type Mock struct {
	Producer Producer
}

// Run method runs through the list of given topics.
// Format of the file should be, key then value then line break.
// It tries to read from the reader line by line and publish message in the topic.
// Please look into the tests for example.
func (m *Mock) Run(ctx context.Context, tps ...*MockTopic) error {
	for _, tpc := range tps {
		scn := bufio.NewScanner(tpc.Reader)
		var key *Key

		for scn.Scan() {
			if len(scn.Text()) == 0 {
				key = nil
				continue
			}

			if key != nil && len(key.Type) > 0 {
				val := reflect.
					New(reflect.TypeOf(tpc.Type)).
					Interface()

				if err := json.Unmarshal(scn.Bytes(), val); err != nil {
					return err
				}

				mss := []*Message{}

				if len(tpc.Partitions) == 0 {
					tpc.Partitions = []int32{kafka.PartitionAny}
				}

				for _, ptn := range tpc.Partitions {
					ptc := ptn
					msg := &Message{
						Config:    tpc.Config,
						Topic:     tpc.Topic,
						Partition: &ptc,
						Key:       key,
						Value:     val,
					}
					mss = append(mss, msg)
				}

				if err := m.Producer.Produce(ctx, mss...); err != nil {
					return err
				}
			}

			if key == nil {
				key = new(Key)

				if err := json.Unmarshal(scn.Bytes(), key); err != nil {
					return err
				}
			}
		}
	}

	_ = m.Producer.Flush(1000)

	return nil
}
