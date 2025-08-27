package wmf

import (
	"context"
	"testing"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/submodules/tracing"

	"github.com/stretchr/testify/suite"
)

type wmfTestSuite struct {
	suite.Suite
	env    *env.Environment
	tracer tracing.Tracer
}

type mockTracer struct{}

func (m *mockTracer) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockTracer) StartTrace(ctx context.Context, _ string, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (m *mockTracer) Trace(context.Context, map[string]string) (func(error, string), context.Context) {
	return func(error, string) {}, context.Background()
}

func (s *wmfTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.tracer = &mockTracer{}
}

func (s *wmfTestSuite) TestNew() {
	clt := NewAPI(s.env, s.tracer)
	s.Assert().NotNil(clt)
}

func TestWMF(t *testing.T) {
	suite.Run(t, new(wmfTestSuite))
}
