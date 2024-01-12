package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type eventTestSuite struct {
	suite.Suite
	event *Event
}

func (s *eventTestSuite) SetupTest() {
	dateCreated := time.Now().UTC()
	ptn := 0
	oft := int64(0)

	s.event = &Event{
		Identifier:    "ec369574-bf5e-4325-a720-c78d36a80cdb",
		Type:          "create",
		DateCreated:   &dateCreated,
		FailCount:     1,
		FailReason:    "internal server error",
		DatePublished: &dateCreated,
		Partition:     &ptn,
		Offset:        &oft,
	}
}

func (s *eventTestSuite) TestNewEventSchema() {
	sch, err := NewEventSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.event)
	s.Assert().NoError(err)

	event := new(Event)
	s.Assert().NoError(avro.Unmarshal(sch, data, event))
	s.Assert().Equal(s.event.Identifier, event.Identifier)
	s.Assert().Equal(s.event.Type, event.Type)
	s.Assert().Equal(s.event.DateCreated.Format(time.RFC1123), event.DateCreated.Format(time.RFC1123))
	s.Assert().Equal(s.event.FailCount, event.FailCount)
}

func (s *eventTestSuite) TestNewEvent() {
	event := NewEvent(s.event.Type)
	s.Assert().Equal(s.event.Type, event.Type)
	s.Assert().NotEmpty(event.Identifier)
	s.Assert().NotEmpty(event.DateCreated)
	s.Assert().Equal(0, event.FailCount)
	s.Assert().Nil(event.Partition)
	s.Assert().Nil(event.Offset)
	s.Assert().NotEmpty(event.DatePublished)
}

func TestEvent(t *testing.T) {
	suite.Run(t, new(eventTestSuite))
}
