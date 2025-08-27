package log

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type logTestSuite struct {
	suite.Suite
}

func (s *logTestSuite) TestNew() {
	lgr, err := New()

	s.Assert().NotNil(lgr)
	s.Assert().NoError(err)
}

func (s *logTestSuite) TestLogger() {
	s.Assert().NotNil(logger)
}

func (s *logTestSuite) TestGetZap() {
	s.Assert().NotNil(GetZap())
}
func captureOutput(s *logTestSuite, f func()) (string, string) {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Redirect stdout and stderr
	os.Stdout = wOut
	os.Stderr = wErr

	// Copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, rOut)
		s.Nil(err)
		outC <- buf.String()
		close(outC)
	}()
	errC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, rErr)
		s.Nil(err)
		errC <- buf.String()
		close(errC)
	}()

	// Run the function
	f()

	// Restore stdout and stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	out := <-outC
	err := <-errC
	return out, err
}

func (s *logTestSuite) TestDefaultStderrLogger() {
	os.Setenv("DISABLE_SPLIT_LOGGING", "1")
	os.Setenv("LOG_LEVEL", "info")

	stdout, stderr := captureOutput(s, func() {
		lgr, err := New()
		s.Nil(err)
		lgr.Info("info logging")
		lgr.Error("error logging")
	})
	s.Empty(stdout)
	s.Contains(stderr, "info logging")
	s.Contains(stderr, "error logging")
}

func (s *logTestSuite) TestSplitStreamLogger() {
	os.Setenv("DISABLE_SPLIT_LOGGING", "")
	os.Setenv("LOG_LEVEL", "info")

	stdout, stderr := captureOutput(s, func() {
		lgr, err := New()
		s.Nil(err)
		lgr.Info("info logging")
		lgr.Error("error logging")
	})
	s.Contains(stdout, "info logging")
	s.Contains(stderr, "error logging")
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(logTestSuite))
}
