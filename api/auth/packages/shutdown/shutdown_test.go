package shutdown_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/api/auth/packages/shutdown"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type serverMock struct {
	mock.Mock
}

// Close method mocks http.Server Close() behaviour.
func (srv *serverMock) Close() error {
	return srv.Called().Error(0)
}

type shutdownTestSuite struct {
	suite.Suite
	cancel context.CancelFunc
	helper *shutdown.Helper
	srv    *serverMock
}

// SetupTest initializes server mock and shutdown.Helper.
func (s *shutdownTestSuite) SetupTest() {
	s.helper = shutdown.NewHelper(context.Background())
	s.srv = new(serverMock)
}

// TearDownSuite eventually cancels the helper's context if not cancelled.
func (s *shutdownTestSuite) TearDownSuite() {
	if s.helper.Ctx().Err() != context.Canceled {
		s.cancel()
	}
}

// TestWaitSuccess verifies that the http.Server closes gracefully and helper's context is cancelled, when we receive an interrupt.
func (s *shutdownTestSuite) TestWaitSuccess() {
	s.srv.On("Close").Return(nil)

	interruptTime := time.Now()
	s.helper.Kill()

	s.Assert().NoError(s.helper.Wait(s.srv))
	s.srv.AssertNumberOfCalls(s.T(), "Close", 1)

	// Verify the sleep duration between interrupt time and server close time.
	s.Assert().True(time.Now().After(interruptTime))
	s.Assert().Equal(context.Canceled, s.helper.Ctx().Err())
}

// TestWaitFail verifies error case of shutdown.Wait() when server failed to close, when we receive an interrupt. Helper's context will still be cancelled.
func (s *shutdownTestSuite) TestWaitFail() {
	err := errors.New("Failed to close server")
	s.srv.On("Close").Return(err)

	interruptTime := time.Now()
	s.helper.Kill()

	s.Assert().Equal(err, s.helper.Wait(s.srv))
	s.srv.AssertNumberOfCalls(s.T(), "Close", 1)

	// Verify the sleep duration between interrupt time and server close time.
	s.Assert().True(time.Now().After(interruptTime))
	s.Assert().Equal(context.Canceled, s.helper.Ctx().Err())
}

// TestShutdown runs the test suite for shutdown package.
func TestShutdown(t *testing.T) {
	suite.Run(t, new(shutdownTestSuite))
}
