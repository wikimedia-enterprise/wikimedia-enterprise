package tracing

import (
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/submodules/tracing"
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
