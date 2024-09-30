package tracing

import (
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/services/event-bridge/config/env"
)

func NewAPI(env *env.Environment) (tracing.Tracer, error) {
	//ctx := context.Background()

	opts := []tracing.ClientOption{tracing.WithGRPCHost(env.OTELCollectorAddr),
		tracing.WithSamplingRate(env.TracingSamplingRate),
		tracing.WithServiceName(env.ServiceName),
		//tracing.WithContext(ctx),
	}

	return tracing.NewAPI(opts...)
}
