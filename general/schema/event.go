package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/hamba/avro/v2"
)

// Type of events supported by the system.
const (
	EventTypeUpdate           = "update"
	EventTypeCreate           = "create"
	EventTypeDelete           = "delete"
	EventTypeMove             = "move"
	EventTypeVisibilityChange = "visibility-change"
)

// ConfigEvent schema configuration for Event.
var ConfigEvent = &Config{
	Type: ConfigTypeValue,
	Name: "Event",
	Schema: `{
		"type": "record",
		"name": "Event",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "string"
			},
			{
				"name": "type",
				"type": "string"
			},
			{
				"name": "date_created",
				"type": [
					"null",
					{
						"type": "long",
						"logicalType": "timestamp-micros"
					}
				]
			},
			{
				"name": "fail_count",
				"type": "int"
			},
			{
				"name": "fail_reason",
				"type": "string",
				"default": ""
			},
			{
				"name": "date_published",
				"type": [
					"null",
					{
						"type": "long",
						"logicalType": "timestamp-micros"
					}
				],
				"default": null
			},
			{
				"name": "is_internal_message",
				"type": ["null", "boolean"],
				"default": null
			}
		]
	}`,
	Reflection: Event{},
}

// NewEventSchema create new event avro schema.
func NewEventSchema() (avro.Schema, error) {
	return New(ConfigEvent)
}

// Event meta data for every event that happens in the system.
type Event struct {
	Identifier    string     `json:"identifier,omitempty" avro:"identifier"`
	Type          string     `json:"type,omitempty" avro:"type"`
	DateCreated   *time.Time `json:"date_created,omitempty" avro:"date_created"`
	FailCount     int        `json:"fail_count,omitempty" avro:"fail_count"`
	FailReason    string     `json:"fail_reason,omitempty" avro:"fail_reason"`
	DatePublished *time.Time `json:"date_published,omitempty" avro:"date_published"`
	Partition     *int       `json:"partition,omitempty"`
	Offset        *int64     `json:"offset,omitempty"`
	IsInternal    *bool      `json:"-" avro:"is_internal_message"`
}

// NewEvent returns an instance of Event struct with the passed eventType and other default values.
func NewEvent(eventType string) *Event {
	dtn := time.Now().UTC()
	isInternal := false

	return &Event{
		Identifier:    uuid.New().String(),
		Type:          eventType,
		DateCreated:   &dtn,
		DatePublished: &dtn,
		FailCount:     0,
		IsInternal:    &isInternal,
	}
}

func (e *Event) SetDatePublished(dtp *time.Time) {
	e.DatePublished = dtp
}

func (e *Event) SetType(typ string) {
	e.Type = typ
}

func (e *Event) SetIdentifier(id string) {
	e.Identifier = id
}
