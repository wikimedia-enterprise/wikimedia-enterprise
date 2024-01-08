package aggregate_test

import (
	"errors"
	"testing"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"

	"github.com/stretchr/testify/suite"
)

type isNonFatalErrTestSuite struct {
	suite.Suite
}

func (s *isNonFatalErrTestSuite) TestErrPageNotFound() {
	s.True(
		aggregate.IsNonFatalErr(wmf.ErrPageNotFound),
		"Expected ErrPageNotFound to be a non-fatal error")
}

func (s *isNonFatalErrTestSuite) TestNotFound() {
	s.True(
		aggregate.IsNonFatalErr(errors.New("404 Not Found")),
		"Expected 404 Not Found to be a non-fatal error")
}

func (s *isNonFatalErrTestSuite) TestForbidden() {
	s.True(
		aggregate.IsNonFatalErr(errors.New("403 Forbidden")),
		"Expected 403 Forbidden to be a non-fatal error")
}

func (s *isNonFatalErrTestSuite) TestUnknownError() {
	s.False(
		aggregate.IsNonFatalErr(errors.New("Unknown Error")),
		"Expected Unknown Error to be a fatal error")
}

func (s *isNonFatalErrTestSuite) TestEmptyError() {
	s.False(
		aggregate.IsNonFatalErr(errors.New("")),
		"Expected empty error to be a fatal error")
}

func (s *isNonFatalErrTestSuite) TestNilError() {
	s.False(
		aggregate.IsNonFatalErr(nil),
		"Expected nil error to be a fatal error")
}

func TestIsNonFatalErr(t *testing.T) {
	suite.Run(t, new(isNonFatalErrTestSuite))
}
