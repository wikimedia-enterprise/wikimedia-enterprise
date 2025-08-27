package schema

import (
	"testing"
	"time"

	"github.com/google/uuid"
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
	isInternal := false

	s.event = &Event{
		Identifier:    "ec369574-bf5e-4325-a720-c78d36a80cdb",
		Type:          "create",
		DateCreated:   &dateCreated,
		FailCount:     1,
		FailReason:    "internal server error",
		DatePublished: &dateCreated,
		Partition:     &ptn,
		Offset:        &oft,
		IsInternal:    &isInternal,
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
	s.Assert().Equal(
		s.event.DateCreated.Truncate(time.Microsecond).Format(time.RFC3339Nano),
		event.DateCreated.Truncate(time.Microsecond).Format(time.RFC3339Nano),
	)
	s.Assert().Equal(s.event.FailCount, event.FailCount)
	s.Assert().Equal(*s.event.IsInternal, *event.IsInternal)
}

func (s *eventTestSuite) TestSetters() {
	evt := NewEvent(EventTypeCreate)

	dt := time.Now().Add(-1 * time.Hour).UTC()
	evt.SetDatePublished(&dt)
	s.Assert().Equal(dt, *evt.DatePublished)

	evt.SetType(EventTypeDelete)
	s.Assert().Equal(evt.Type, EventTypeDelete)

	uuid := uuid.New().String()
	evt.SetIdentifier(uuid)
	s.Assert().Equal(evt.Identifier, uuid)
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
	s.Assert().NotNil(event.IsInternal)
	s.Assert().False(*event.IsInternal)
}

func TestEvent(t *testing.T) {
	suite.Run(t, new(eventTestSuite))
}
