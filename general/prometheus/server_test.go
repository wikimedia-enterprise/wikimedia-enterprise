package prometheus

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type prometheusTestSuite struct {
	suite.Suite
	promPort int
}

func (p *prometheusTestSuite) SetupSuite() {
	p.promPort = 12411
}

func (p *prometheusTestSuite) TestNewServer() {
	// Using NotFoundHandler as a stub.
	srv := NewServer(p.promPort, http.NotFoundHandler())
	p.Assert().NotEmpty(srv)
}

func TestNew(t *testing.T) {
	suite.Run(t, new(prometheusTestSuite))
}
