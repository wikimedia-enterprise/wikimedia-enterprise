package schema

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hamba/avro/v2"
)

// Error messages from the helper.
var (
	ErrNoMessageConfig = errors.New("no config was provided to the message")
	ErrEmptyTopic      = errors.New("empty topic name was provied")
	ErrEmptyKey        = errors.New("key cannot be nil")
	ErrEmptyValue      = errors.New("value cannot be nil")

	SchemaParsingMutex sync.RWMutex
)

// Syncher is an interface that wraps default Sync method for unit testing.
type Syncher interface {
	Sync(ctx context.Context, topic string, cfg *Config, names ...string) (*Schema, error)
}

// Getter is an interface that wraps default Getter method for unit testing.
type Getter interface {
	Get(ctx context.Context, id int) (*Schema, error)
}

// SyncGetter is an interface that wraps default Sync and Get methods for unit testing.
type SyncGetter interface {
	Syncher
	Getter
}

// Marshaler is an interface that wraps default Marshal method for unit testing.
type Marshaler interface {
	Marshal(ctx context.Context, topic string, cfg *Config, v interface{}) ([]byte, error)
}

// Unmarshaler is an interface that wraps default Unmarshal method for unit testing.
type Unmarshaler interface {
	Unmarshal(ctx context.Context, data []byte, v interface{}) error
	UnmarshalNoCache(ctx context.Context, data []byte, v interface{}) error
}

// MarshalUnmarshaler is an interface that wraps default Unmarshal and Marshal methods for unit testing.
type MarshalUnmarshaler interface {
	Marshaler
	Unmarshaler
}

// Producer is an interface that wraps default Produce method for unit testing.
type Producer interface {
	Produce(ctx context.Context, msgs ...*Message) error
	Flush(int) int
}

// UnmarshalProducer is an interface that wraps default Unmarshal and Produce methods for unit testing.
type UnmarshalProducer interface {
	Unmarshaler
	Producer
}

// KafkaProducer is an interface to wrap confluent kafka producer for unit testing.
type KafkaProducer interface {
	ProduceChannel() chan *kafka.Message
	Flush(int) int
}

// Message custom wrapper ont top of kafka message. Intended to simplify schema encoding.
type Message struct {
	Subject   string
	Topic     string
	Partition *int32
	Key       interface{}
	Value     interface{}
	KeyConfig *Config
	Config    *Config
	Headers   []kafka.Header
}

// NewHelper creates new instance of the schema helper.
func NewHelper(reg GetterCreator, prod KafkaProducer) *Helper {
	return &Helper{
		reg:  reg,
		prod: prod,
	}
}

// Helper intended to help synchronize and retrieve schemas from registry.
type Helper struct {
	reg     GetterCreator
	prod    KafkaProducer
	configs sync.Map
	schemas sync.Map
}

// Sync creates schema registry subjects and references by automatically
// resolving dependencies inside the configuration (recursively resolves schema dependencies).
func (h *Helper) Sync(ctx context.Context, topic string, cfg *Config, names ...string) (*Schema, error) {
	name := fmt.Sprintf("%s-%s", topic, cfg.Type)

	if len(names) > 0 {
		name = fmt.Sprintf("%s-%s-%s", topic, strings.ToLower(names[0]), cfg.Type)
	}

	sub := Subject{
		Schema:     cfg.Schema,
		SchemaType: SchemaTypeAVRO,
	}

	for _, ref := range cfg.References {
		sch, err := h.Sync(ctx, topic, ref, ref.Name)

		if err != nil {
			return nil, err
		}

		sub.References = append(sub.References, &Reference{
			Name:    ref.Name,
			Subject: sch.Subject,
			Version: sch.Version,
		})
	}

	if _, err := h.reg.CreateSubject(ctx, name, &sub); err != nil {
		return nil, err
	}

	sch, err := h.reg.GetStoredSubject(ctx, name, &sub)
	if err != nil {
		return nil, err
	}

	if err := sch.Parse(); err != nil {
		return nil, err
	}

	h.schemas.Store(sch.ID, sch)

	return sch, nil
}

// Get return schemas by id. Recursively resolves
// dependencies and caches the result.
// When a new schema is being resolved we prevent parallel execution, because
// calls to `avro.Parse()` from different versions of the schema may
// be in conflict.
func (h *Helper) Get(ctx context.Context, id int) (*Schema, error) {
	SchemaParsingMutex.RLock()
	cached, ok := h.schemas.Load(id)
	SchemaParsingMutex.RUnlock()
	if ok {
		return cached.(*Schema), nil
	}

	SchemaParsingMutex.Lock()
	defer SchemaParsingMutex.Unlock()

	// Check again, a previous holder of the write lock might have created it.
	if cached, ok = h.schemas.Load(id); ok {
		return cached.(*Schema), nil
	}

	sch, err := h.reg.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	for _, ref := range sch.References {
		sch, err := h.reg.GetBySubject(ctx, ref.Subject, ref.Version)

		if err != nil {
			return nil, err
		}

		if err := h.Resolve(ctx, sch.ID); err != nil {
			return nil, err
		}
	}

	if err := sch.Parse(); err != nil {
		return nil, err
	}

	h.schemas.Store(id, sch)

	return sch, nil
}

// Resolve fetches the requested schema if needed and resolves its dependencies.
// The actual outcome of this is calling `avro.Parse()`, which makes the dependencies available for parent schemas.
func (h *Helper) Resolve(ctx context.Context, id int) error {
	if sch, ok := h.schemas.Load(id); ok {
		return sch.(*Schema).Parse()
	}

	sch, err := h.reg.GetByID(ctx, id)

	if err != nil {
		return err
	}

	for _, ref := range sch.References {
		sch, err := h.reg.GetBySubject(ctx, ref.Subject, ref.Version)

		if err != nil {
			return err
		}

		if err := h.Resolve(ctx, sch.ID); err != nil {
			return err
		}
	}

	if err := sch.Parse(); err != nil {
		return err
	}

	h.schemas.Store(id, sch)

	return nil
}

// Marshal syncs schema with schema registry and encodes message into AVRO.
func (h *Helper) Marshal(ctx context.Context, topic string, cfg *Config, v interface{}) ([]byte, error) {
	key := fmt.Sprintf("%s-%s", topic, cfg.Type)

	SchemaParsingMutex.RLock()
	sch, ok := h.schemas.Load(key)
	SchemaParsingMutex.RUnlock()

	if ok {
		return sch.(*Schema).Marshal(v)
	}

	SchemaParsingMutex.Lock()
	defer SchemaParsingMutex.Unlock()

	// Check again, a previous holder of the write lock might have created it.
	if sch, ok := h.schemas.Load(key); ok {
		return sch.(*Schema).Marshal(v)
	}

	sch, err := h.Sync(ctx, topic, cfg)

	if err != nil {
		return nil, err
	}

	h.schemas.Store(key, sch)

	return sch.(*Schema).Marshal(v)
}

// Unmarshal gets schema by id and decodes it into struct.
func (h *Helper) Unmarshal(ctx context.Context, data []byte, v interface{}) error {
	return h.unmarshal(ctx, data, v, false)
}

// Unmarshal gets schema by id and decodes it into struct, with avro cache disabled.
func (h *Helper) UnmarshalNoCache(ctx context.Context, data []byte, v interface{}) error {
	return h.unmarshal(ctx, data, v, true)
}

// Unmarshal gets schema by id and decodes it into struct, allowing to enable or disable avro caching.
func (h *Helper) unmarshal(ctx context.Context, data []byte, v interface{}, dch bool) error {
	sch, err := h.Get(ctx, GetID(data))

	if err != nil {
		return err
	}

	var api avro.API

	if cfg, exs := h.configs.Load(sch.ID); exs {
		api = cfg.(avro.API) // using cached api
	} else {
		cfg := avro.Config{
			// Modify max slice size to 20MB.
			// This is needed to decode large messages.
			MaxByteSliceSize: 1024 * 1024 * 20,

			// Disabling avro caching, this is needed for snaphshots.
			// Due to multiple writes overtime and different avro schemas.
			DisableCaching: dch,
		}
		api = cfg.Freeze() // getting a new instance of the api

		h.configs.Store(sch.ID, api) // caching api
	}

	return sch.Unmarshal(data, v, api)
}

// Produce wraps kafka producer to work with our custom messages.
// Receives list of messages encode and publish them.
func (h *Helper) Produce(ctx context.Context, msgs ...*Message) error {
	kmsgs := []*kafka.Message{}

	for _, msg := range msgs {
		if msg.KeyConfig == nil {
			msg.KeyConfig = ConfigKey
		}

		if msg.Config == nil {
			return ErrNoMessageConfig
		}

		if len(msg.Topic) <= 0 {
			return ErrEmptyTopic
		}

		if msg.Key == nil {
			return ErrEmptyKey
		}

		if msg.Value == nil {
			return ErrEmptyValue
		}

		sbj := msg.Topic

		if len(msg.Subject) > 0 {
			sbj = msg.Subject
		}

		key, err := h.Marshal(ctx, sbj, msg.KeyConfig, msg.Key)

		if err != nil {
			return err
		}

		val, err := h.Marshal(ctx, sbj, msg.Config, msg.Value)

		if err != nil {
			return err
		}

		if msg.Partition == nil {
			ptn := kafka.PartitionAny
			msg.Partition = &ptn
		}

		kmsgs = append(kmsgs, &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &msg.Topic,
				Partition: *msg.Partition,
			},
			Key:     key,
			Value:   val,
			Headers: msg.Headers,
		})
	}

	for _, msg := range kmsgs {
		h.prod.ProduceChannel() <- msg
	}

	return nil
}

// Flush publishes messages from the producer queue.
func (h *Helper) Flush(timeMs int) int {
	return h.prod.Flush(timeMs)
}
