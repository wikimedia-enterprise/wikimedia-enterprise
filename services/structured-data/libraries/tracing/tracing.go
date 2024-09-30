package tracing

import (
	"context"
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/services/structured-data/config/env"
)

// NewAPI creates a new tracing API.
func NewAPI(env *env.Environment) (tracing.Tracer, error) {
	ctx := context.Background()

	opts := []tracing.ClientOption{
		tracing.WithSamplingRate(env.TracingSamplingRate),
		tracing.WithServiceName(env.ServiceName),
		tracing.WithContext(ctx),
		tracing.WithGRPCHost(env.TracingGRPCHost),
	}

	return tracing.NewAPI(opts...)
}
