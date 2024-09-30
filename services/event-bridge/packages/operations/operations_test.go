package operations_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/common"
	"wikimedia-enterprise/services/event-bridge/packages/operations"
	"wikimedia-enterprise/services/event-bridge/packages/transformer"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

type operationsProducerMock struct {
	mock.Mock
	schema.UnmarshalProducer
}

func (p *operationsProducerMock) Produce(_ context.Context, msgs ...*schema.Message) error {
	for _, msg := range msgs {
		switch m := msg.Value.(type) {
		case *schema.Article:
			if m.Event == nil {
				return fmt.Errorf("empty event field")
			}

			m.Event = nil
		}
	}

	return p.Called(msgs[0]).Error(0)
}

type operationsTestSuite struct {
	suite.Suite
	trs transformer.TransformPageChangeToArticle
	ops *operations.Operations
	prm *operations.OperationParameters
	op  string
	evt *eventstream.PageChange
	env *env.Environment
	ctx context.Context
}

func (s *operationsTestSuite) SetupTest() {
	s.trs = transformer.New()
	s.ops = operations.New(s.trs)

	kafkaBootstrapServersKey := "KAFKA_BOOTSTRAP_SERVERS"
	redisAddrKey := "REDIS_ADDR"
	kafkaBootstrapServers := "localhost:8001"
	redisAddr := "localhost:2021"
	kafkaCredsKey := "KAFKA_CREDS"
	kafkaCreds := `{"username":"admin","password":"123456"}`
	redisPasswordKey := "REDIS_PASSWORD"
	redisPassword := "12345"
	schemaRegistryURLKey := "SCHEMA_REGISTRY_URL"
	schemaRegistryURL := "localhost:300"
	schemaRegistryCredsKey := "SCHEMA_REGISTRY_CREDS"
	schemaRegistryCreds := `{"username":"reg_admin","password":"654321"}`
	OTELCollectorAddrKey := "OTEL_COLLECTOR_ADDR"
	OTELCollectorAddr := "collector:4317"
	TracingSamplingRateKey := "TRACING_SAMPLING_RATE"
	TracingSamplingRate := 0.1
	ServiceNameKey := "SERVICE_NAME"
	ServiceName := "event-bridge.service"
	outputTopics := "{\"version\":[\"v1\"],\"service_name\":\"event-bridge\",\"location\":\"aws\"}"
	outputTopicsKey := "OUTPUT_TOPICS"

	os.Setenv(kafkaBootstrapServersKey, kafkaBootstrapServers)
	os.Setenv(kafkaCredsKey, kafkaCreds)
	os.Setenv(redisAddrKey, redisAddr)
	os.Setenv(redisPasswordKey, redisPassword)
	os.Setenv(schemaRegistryURLKey, schemaRegistryURL)
	os.Setenv(schemaRegistryCredsKey, schemaRegistryCreds)
	os.Setenv(OTELCollectorAddrKey, OTELCollectorAddr)
	os.Setenv(TracingSamplingRateKey, fmt.Sprintf("%f", TracingSamplingRate))
	os.Setenv(ServiceNameKey, ServiceName)
	os.Setenv(outputTopicsKey, outputTopics)

	var err error

	s.env, err = env.New()

	if err != nil {
		log.Fatal(err)
	}

	s.prm = &operations.OperationParameters{
		Producer:   &operationsProducerMock{},
		Env:        s.env,
		Dictionary: &langid.Dictionary{},
		Tracer:     nil,
	}

	s.evt = &eventstream.PageChange{}
	s.evt.Data.Database = "commonswiki"

	s.ctx = context.WithValue(context.Background(), common.EventIdentifierContextKey, "123")

}

func (s *operationsTestSuite) TestOperations() {

	s.evt.Data.PageChangeKind = s.op

	tpcs, art, err := s.ops.Execute(s.ctx, s.prm, s.evt)

	s.Assert().NoError(err, "Execute returned error")

	s.Assert().Equal(1, len(tpcs), "failed to assert length of topics")

	var op string

	switch s.op {
	case operations.Create:
		op = schema.EventTypeCreate
	case operations.UnDelete:
		op = schema.EventTypeCreate
	case operations.Update:
		op = schema.EventTypeUpdate
	case operations.Delete:
		op = schema.EventTypeDelete
	case operations.Move:
		op = schema.EventTypeMove
	case operations.VisibilityChange:
		op = schema.EventTypeVisibilityChange
	}

	s.Assert().Equal(fmt.Sprintf("aws.event-bridge.commons-%s.v1", op), tpcs[0], "Failed to assert topics")

	s.Assert().Equal(s.ctx.Value(common.EventIdentifierContextKey).(string), art.Event.Identifier, "failed to assert event identifier")
}

func TestOperations(t *testing.T) {
	for _, testCase := range []*operationsTestSuite{
		{op: operations.Create},
		{op: operations.Delete},
		{op: operations.Move},
		{op: operations.UnDelete},
		{op: operations.Update},
		{op: operations.VisibilityChange},
	} {
		suite.Run(t, testCase)
	}
}
