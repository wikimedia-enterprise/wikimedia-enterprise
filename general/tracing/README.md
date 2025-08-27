# OpenTelemetry Tracing Library

This Go package provides an OpenTelemetry-based tracing library for distributed tracing.

## Key Components

- `Client`: Tracing client for creating and managing traces.
- `Provider`: Interface for the tracing provider.
- `Tracer`: Interface for creating and managing spans within a trace.

Uses OpenTelemetry's OTLP exporter for sending trace data.

## Usage

1. Import the `tracing` package.

2. Initialize the `Tracer` using dependency injection:

```go
var tracer tracing.Tracer
cxt = context.Background()

func TracerConstructor() {
    tracer, err = tracing.NewAPI(
        tracing.WithServiceName("your-service-name"),
        tracing.WithSamplingRate(1.0),
        tracing.WithGRPCHost("localhost"),
        tracing.WithGRPCPort("4317"),
        tracing.Withcontext(cxt)
    )
    // Handle error

    //return API
    return tracer
}
```

3. Use the `Trace` method to create spans:

```go
ctx := context.Background()
attributes := map[string]string{
    "title": "Eminem",
    "project": "enwiki",
}

endSpan, traceCtx := tracer.Trace(ctx, attributes)
defer endSpan(nil, "request complete")

// Your code here
```

4. Record errors and set span status:

```go
err := someOperation()
if err != nil {
    endSpan(err, "Error occurred")
    return
}
```

For more practical examples please see draft MRs in [] and []. 

## Configuration Options

- `WithServiceName`: Service name being traced.
- `WithSamplingRate`: Sampling rate for traces (0 to 1).
- `WithGRPCHost`: Hostname of the OTLP gRPC collector.
- `WithGRPCPort`: Port number of the OTLP gRPC collector.

## Instrumentation

1. Initialize the `Tracer` in dependency injection.
2. Use the `Trace` method to create spans around important operations.
3. Record errors and set span status as needed.


## Areas of Improvement
1. Implement Gracefull Shutdown
2. Opensource Tracing Library and use interfaces in other libraries
3. Add OTLP metrics 