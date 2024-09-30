package tracing

import (
	"context"
	"fmt"

	"runtime"
	"strings"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	tracer "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"

	otr "go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

// Trace is an interface that defines methods to interact with the tracer
type Tracer interface {
	Trace(context context.Context, attributes map[string]string) (func(err error, msg string), context.Context)
	StartTrace(context context.Context, service string, attributes map[string]string) (func(err error, msg string), context.Context)
	Shutdowner
}

// ClientOption is a function interface that injects the client into a Client
type ClientOption func(*Client)

type Client struct {
	Provider     Provider
	ServiceName  string
	SamplingRate float64
	GRPCHost     string
	Context      context.Context
	exporter     *otlptrace.Exporter
}

type ForceFlusher interface {
	ForceFlush(ctx context.Context) error
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type SpanProcessor interface {
	RegisterSpanProcessor(sp tracer.SpanProcessor)
}

type UnregisterSpanProcessor interface {
	UnregisterSpanProcessor(sp tracer.SpanProcessor)
}

// Provider is an interface that defines methods to interact with the provider
type Provider interface {
	ForceFlusher
	Shutdowner
	SpanProcessor
	Tracer(name string, opts ...trace.TracerOption) trace.Tracer
	UnregisterSpanProcessor
}

// TracerProvider is a struct that implements the Provider interface
type TracerProvider struct {
	Provider
}

// NewAPI is a function that creates a new tracer
func NewAPI(opts ...ClientOption) (Tracer, error) {
	return NewClient(opts...)
}

// NewClient is a function that creates a new client
func NewClient(ops ...ClientOption) (*Client, error) {
	cl := &Client{}

	// for injecting service level attributes
	for _, op := range ops {
		op(cl)
	}

	eco, err := newExporterOption(cl.Context, cl.GRPCHost)
	if err != nil {
		return nil, fmt.Errorf("error creating exporter: %v", err)
	}

	eco(cl)

	prf, err := newProviderOption(cl.ServiceName, cl.SamplingRate, cl.exporter)
	if err != nil {
		return nil, fmt.Errorf("error creating provider: %v", err)
	}

	prf(cl)

	return cl, nil
}

func newProviderOption(serviceName string, samplingRate float64, exp *otlptrace.Exporter) (ClientOption, error) {
	// labels/tags/resources that are common to all traces.
	resource := otr.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	idg := xray.NewIDGenerator()

	provider := tracer.NewTracerProvider(
		tracer.WithBatcher(exp),
		tracer.WithResource(resource),
		tracer.WithSampler(tracer.ParentBased(tracer.TraceIDRatioBased(samplingRate))),
		tracer.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(xray.Propagator{})

	return func(c *Client) {
		c.SamplingRate = samplingRate
		c.Provider = provider
	}, nil
}

func newExporterOption(context context.Context, hostname string) (ClientOption, error) {
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(hostname, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating grpc client: %v", err)
	}

	exporter, err := otlptracegrpc.New(context, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("error creating exporter: %v", err)
	}

	return func(c *Client) {
		c.exporter = exporter
	}, nil
}

// Trace is a method that creates a new span and returns a function that ends the span.
func (cl *Client) Trace(context context.Context, attributes map[string]string) (func(err error, msg string), context.Context) {
	attrs := []attribute.KeyValue{}
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	var snm string
	pc, _, _, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)

		if fn == nil {
			snm = "unknown caller"
		} else {
			snm = fn.Name()
			snl := strings.Split(snm, "/")

			if len(snl) > 0 {
				snm = snl[len(snl)-1]
			}
		}
	}

	trx, span := cl.Provider.Tracer("services").Start(context, snm, trace.WithAttributes(attrs...))

	return func(err error, msg string) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, msg)
		}

		span.End()
	}, trx
}

// StartTrace is a method that creates a new span and returns a function that ends the span for event handlers.
func (cl *Client) StartTrace(context context.Context, service string, attributes map[string]string) (func(err error, msg string), context.Context) {
	attrs := []attribute.KeyValue{}
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	trx, span := cl.Provider.Tracer("services").Start(context, service, trace.WithAttributes(attrs...))

	return func(err error, msg string) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, msg)
		}

		span.End()
	}, trx
}

// Shutdown is a method that shuts down the client
func (c *Client) Shutdown(ctx context.Context) error {
	if err := c.Provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down trace provider: %v", err)
	}

	if err := c.exporter.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down exporter: %v", err)
	}

	return nil
}

// WithServiceName is a function that sets the service name for the client
func WithServiceName(serviceName string) ClientOption {
	return func(c *Client) {
		c.ServiceName = serviceName
	}
}

// WithSamplingRate is a function that sets the sampling rate for the client
func WithSamplingRate(samplingRate float64) ClientOption {
	return func(c *Client) {
		c.SamplingRate = samplingRate
	}
}

// WithGRPCHost is a function that sets the host for the client
func WithGRPCHost(host string) ClientOption {
	return func(c *Client) {
		c.GRPCHost = host
	}
}

// WithContext is a function that sets the context for the client
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.Context = ctx
	}
}

// TracerInjector is a function interface that injects the tracer into a Client
type TracerInjector func(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context)

// InjectTrace is a function that injects the tracer into a Client
func InjectTrace(clt Tracer) TracerInjector {
	return func(ctx context.Context, attributes map[string]string) (func(err error, msg string), context.Context) {
		return clt.Trace(ctx, attributes)
	}
}
