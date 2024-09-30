package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HeadersTestSuite tests the HeadersCarrier.
type HeadersTestSuite struct {
	suite.Suite
	headers     *HeadersCarrier
	spanContext trace.SpanContext
}

// SetupSuite initializes the HeadersCarrier and SpanContext.
func (h *HeadersTestSuite) SetupSuite() {
	h.headers = &HeadersCarrier{headers: map[string]string{}}
	spanContext := trace.SpanContext{}
	spanContext = spanContext.WithTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	spanContext = spanContext.WithSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	spanContext = spanContext.WithTraceFlags(trace.FlagsSampled)
	h.spanContext = spanContext
}

// TestNewHeadersCarrier tests the NewHeadersCarrier function.
func (h *HeadersTestSuite) TestExtractContext() {
	ctx := context.Background()
	propagators := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	ctx = trace.ContextWithSpanContext(ctx, h.spanContext)
	var hds HeaderCarrier = &HeadersCarrier{headers: map[string]string{}}
	extractedCtx := propagators.Extract(ctx, hds)
	extractedSpanContext := trace.SpanContextFromContext(extractedCtx)

	h.Assert().Equal(h.spanContext.TraceID(), extractedSpanContext.TraceID())
	h.Assert().Equal(h.spanContext.TraceFlags(), extractedSpanContext.TraceFlags())
	h.Assert().Equal(h.spanContext.SpanID(), extractedSpanContext.SpanID())
}

func (h *HeadersTestSuite) TestInjectContext() {
	ctx := context.Background()
	propagators := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	ctx = trace.ContextWithSpanContext(ctx, h.spanContext)
	hds := &HeadersCarrier{headers: map[string]string{}}
	kafkaHeaders := hds.InjectContext(ctx)
	h.Assert().NotNil(kafkaHeaders)
	extractedCtx := propagators.Extract(context.Background(), hds)
	extractedSpanContext := trace.SpanContextFromContext(extractedCtx)

	h.Assert().Equal(h.spanContext.TraceID(), extractedSpanContext.TraceID())
	h.Assert().Equal(h.spanContext.TraceFlags(), extractedSpanContext.TraceFlags())
}

// TearDownSuite clears the HeadersCarrier and SpanContext.
func (h *HeadersTestSuite) TearDownSuite() {
	h.headers = nil
	h.spanContext = trace.SpanContext{}
}

// TestHeaders runs the HeadersTestSuite.
func TestHeaders(t *testing.T) {
	suite.Run(t, new(HeadersTestSuite))
}
