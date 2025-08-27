package shutdown_test

import (
	"context"
	"testing"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/packages/shutdown"

	"github.com/stretchr/testify/suite"
)

type shutdownTestSuite struct {
	suite.Suite
}

// TestNewHelper ensures the shutdown helper is created ok
func (s *shutdownTestSuite) TestNewHelper() {
	ctx := context.Background()
	h := shutdown.NewHelper(ctx)

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.Assert().Equal(h.Ctx(), cancelCtx)
	s.Assert().NotNil(h.WG())
}

// TestWait ensures the workload on goroutines is finished before shutting down
func (s *shutdownTestSuite) TestWait() {
	finished := false
	sh := shutdown.NewHelper(context.Background())
	wg := sh.WG()
	control := make(chan bool, 1)

	// Simulate a handler's workload by waiting 500ms
	wg.Add(1)
	go func() {
		defer wg.Done()
		control <- true
		time.Sleep(500 * time.Millisecond)
		finished = true
	}()

	go func() {
		<-control
		sh.Kill()
	}()

	// Wait until the handler is done
	sh.Wait(nil)
	s.Assert().True(finished)
}

// TestShutdown runs the test suite for shutdown package.
func TestShutdown(t *testing.T) {
	suite.Run(t, new(shutdownTestSuite))
}
