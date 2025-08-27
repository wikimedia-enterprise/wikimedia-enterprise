package tracing

import (
	"context"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/submodules/tracing"
)

func NewAPI(env *env.Environment) (tracing.Tracer, error) {
	ctx := context.Background()

	opts := []tracing.ClientOption{tracing.WithGRPCHost(env.TracingGrpcHost),
		tracing.WithSamplingRate(env.TracingSamplingRate),
		tracing.WithServiceName(env.ServiceName),
		tracing.WithContext(ctx),
		tracing.WithGRPCHost(env.TracingGrpcHost),
	}

	return tracing.NewAPI(opts...)
}
