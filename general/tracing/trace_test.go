package tracing

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/codes"

	trc "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TraceTestSuite tests the Trace function.
type TraceTestSuite struct {
	suite.Suite
	client *Client
	err   error
	msg  string
	attributes map[string]string
}

// TracerProviderMock is a mock implementation of the TracerProvider interface.
type TracerProviderMock struct {
	mock.Mock
}

// TestTracer is a mock implementation of the Tracer interface.
type TestTracer trace.Tracer

// TestTracerObject is a mock implementation of the Tracer interface.
type TestTracerObject struct {
	trace.Tracer
}

// TestSpan is a mock implementation of the Span interface.
type TestSpan struct {
	trace.Span
}

type spanProcessorState struct {
	sps trc.SpanProcessor
	state sync.Once
}

type spanProcessorStates []*spanProcessorState

// End ends the span.
func (s TestSpan) End(options ...trace.SpanEndOption) {
	fmt.Print("End")
}

// RecordError records an error on the span.
func (s TestSpan) RecordError(err error, opts ...trace.EventOption) {
}

// SetStatus sets the status of the span (ie success, error, throttled etc).
func (s TestSpan) SetStatus(code codes.Code, description string) {
}

// Start starts a new span.
func (tto * TestTracerObject) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	spn := TestSpan {
	}

	wct := trace.ContextWithSpan(ctx, spn)
	return wct, spn
}

// ForceFlush flushes the TracerProviderMock.
func (p *TracerProviderMock) ForceFlush(ctx context.Context) error {
	return nil
}

// GetTracerProvider returns the TracerProviderMock.
func (p *TracerProviderMock) RegisterSpanProcessor(sp trc.SpanProcessor) {
}

// Shutdown shuts down the TracerProviderMock.
func (p *TracerProviderMock) Shutdown(ctx context.Context) error {
	return nil
}

func (p *TracerProviderMock) getSpanProcessors() spanProcessorStates {
	sps := new(spanProcessorState)
	sps.sps = nil
	sps.state = sync.Once{}
	sss := spanProcessorStates{sps}
	fmt.Print(sss)
	return nil
}

// Tracer creates a new Tracer.
func (p *TracerProviderMock) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	tracer := TestTracerObject{}
	
	cll := p.getSpanProcessors()
	fmt.Print(cll)

	return &tracer
}

// UnregisterSpanProcessor deregisters a SpanProcessor with the TracerProviderMock.
func (p *TracerProviderMock) UnregisterSpanProcessor(sp trc.SpanProcessor) {
}

// TestProviderTracer is a mock interface for the TracerProvider.
type TestProviderTracer interface{
	Trace(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context)
}

// SetupSuite initializes the Client.
func (s *TraceTestSuite) SetupSuite() {
	s.client = &Client{
		Provider: &TracerProviderMock{},
	}
}

// TearDownSuite clears the Client.
func (s *TraceTestSuite) TearDownSuite() {
	s.client = nil
}

// TestTrace tests that the Trace function throws NoError
func (s *TraceTestSuite) TestTrace() {
	s.err = nil
	s.msg = "Test"
	s.attributes = map[string]string{
		"test": "test",
	}

	_ , ctx := s.client.Trace(context.Background(), s.attributes)

	s.Assert().NoError(ctx.Err())
}
// TestTraces tests that the Trace function throws an error
func TestTraces(t *testing.T) {
	suite.Run(t, new(TraceTestSuite))
}


