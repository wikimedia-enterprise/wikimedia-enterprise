package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type namespaceTestSuite struct {
	suite.Suite
	namespace *Namespace
}

func (s *namespaceTestSuite) SetupTest() {
	s.namespace = &Namespace{
		Name:          "File",
		AlternateName: "File",
		Identifier:    6,
		Description:   "This is a file.",
		Event:         NewEvent(EventTypeCreate),
	}
}

func (s *namespaceTestSuite) TestNewNamespaceSchema() {
	sch, err := NewNamespaceSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.namespace)
	s.Assert().NoError(err)

	namespace := new(Namespace)
	s.Assert().NoError(avro.Unmarshal(sch, data, namespace))
	s.Assert().Equal(s.namespace.Identifier, namespace.Identifier)
	s.Assert().Equal(s.namespace.Name, namespace.Name)
	s.Assert().Equal(s.namespace.Description, namespace.Description)
	s.Assert().Equal(s.namespace.AlternateName, namespace.AlternateName)
	s.Assert().Equal(s.namespace.Event.Identifier, namespace.Event.Identifier)
	s.Assert().Equal(s.namespace.Event.Type, namespace.Event.Type)
	s.Assert().Equal(s.namespace.Event.DateCreated.Format(time.RFC1123), namespace.Event.DateCreated.Format(time.RFC1123))
}

func TestNamespace(t *testing.T) {
	suite.Run(t, new(namespaceTestSuite))
}
