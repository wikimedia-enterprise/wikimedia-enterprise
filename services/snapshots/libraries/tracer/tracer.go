package tracer

import (
	"context"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/submodules/tracing"
)

func NewAPI(env *env.Environment) (tracing.Tracer, error) {
	ctx := context.Background()

	opts := []tracing.ClientOption{tracing.WithGRPCHost(env.OTELCollectorAddr),
		tracing.WithSamplingRate(env.TracingSamplingRate),
		tracing.WithServiceName(env.ServiceName),
		tracing.WithContext(ctx),
	}

	return tracing.NewAPI(opts...)
}
