package tracing

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opentelemetry.io/otel/propagation"
)

// HeaderCarrier is an interface for storing and retrieving key-value pairs.
type HeaderCarrier interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

// HeadersCarrier stores key-value pairs in a map.
type HeadersCarrier struct {
	headers map[string]string
}

// NewHeadersCarrier initializes a new HeadersCarrier.
func NewHeadersCarrier() *HeadersCarrier {
	return &HeadersCarrier{
		headers: make(map[string]string),
	}
}

// Get retrieves the value for the key in the HeadersCarrier.
func (c *HeadersCarrier) Get(key string) string {
	if val, ok := c.headers[key]; ok {
		return val
	}

	return c.headers[key]
}

// Set stores the key-value pair in the HeadersCarrier.
func (c *HeadersCarrier) Set(key string, value string) {
	c.headers[key] = value
}


// Keys returns the keys stored in the HeadersCarrier.
func (c *HeadersCarrier) Keys() []string {
	kys := make([]string, 0, len(c.headers))
	for key := range c.headers {
		kys = append(kys, key)
	}
	
        return kys
}

// ToKafkaHeaders converts the HeadersCarrier to a slice of kafka.Header.
func (c *HeadersCarrier) ToKafkaHeaders() []kafka.Header {
	hds := make([]kafka.Header, 0, len(c.headers))
	for key, val := range c.headers {
		hds = append(hds, kafka.Header{Key: key, Value: []byte(val)})
	}
	
       return hds
}

// FromKafkaHeaders initializes the HeadersCarrier from a slice of kafka.Header.
func (c *HeadersCarrier) FromKafkaHeaders(kafkaHeaders []kafka.Header) {
	for _, h := range kafkaHeaders {
		c.headers[h.Key] = string(h.Value)
	}
}

// ExtractContext extracts the context from Kafka headers.
func (c *HeadersCarrier) ExtractContext(ctx context.Context) context.Context {
	pps := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	return pps.Extract(ctx, c)
}

// InjectContext injects the context into Kafka headers.
func (c *HeadersCarrier) InjectContext(ctx context.Context) []kafka.Header {
	pps := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	pps.Inject(ctx, c)
	return c.ToKafkaHeaders()
}

