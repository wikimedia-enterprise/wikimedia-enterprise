package schema

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var errHelperMock = errors.New("helper mock error")

type getterCreatorMock struct {
	mock.Mock
}

func (m *getterCreatorMock) GetByID(_ context.Context, id int) (*Schema, error) {
	args := m.Called(id)
	return args.Get(0).(*Schema), args.Error(1)
}

func (m *getterCreatorMock) GetBySubject(_ context.Context, name string, _ ...int) (*Schema, error) {
	args := m.Called(name)
	return args.Get(0).(*Schema), args.Error(1)
}

func (m *getterCreatorMock) CreateSubject(_ context.Context, name string, subject *Subject) (*Schema, error) {
	args := m.Called(name, *subject)
	return args.Get(0).(*Schema), args.Error(1)
}

type producerMock struct {
	mock.Mock
}

func (m *producerMock) ProduceChannel() chan *kafka.Message {
	return m.Called().Get(0).(chan *kafka.Message)
}

func (m *producerMock) Flush(timeMs int) int {
	return m.Called(timeMs).Int(0)
}

type helperTestSuite struct {
	suite.Suite
	prod   *producerMock
	reg    *getterCreatorMock
	shl    *Helper
	ctx    context.Context
	topic  string
	msg    *Message
	valCfg *Config
	keyCfg *Config
	key    interface{}
	val    interface{}
	valSub *Subject
	keySub *Subject
	valSch *Schema
	keySch *Schema
	tms    int
}

func (s *helperTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.tms = 100
	s.valSub = &Subject{
		Schema:     s.valCfg.Schema,
		SchemaType: SchemaTypeAVRO,
	}
	s.keySub = &Subject{
		Schema:     s.keyCfg.Schema,
		SchemaType: SchemaTypeAVRO,
	}
	ptn := int32(0)
	s.msg = &Message{
		Partition: &ptn,
		Topic:     s.topic,
		Key:       s.key,
		Value:     s.val,
		Config:    s.valCfg,
	}

	s.Assert().NoError(s.valSch.Parse())
	s.Assert().NoError(s.keySch.Parse())
}

func (s *helperTestSuite) SetupTest() {
	s.prod = new(producerMock)
	s.reg = new(getterCreatorMock)

	s.shl = new(Helper)
	s.shl.reg = s.reg
	s.shl.prod = s.prod
}

func (s *helperTestSuite) TestSync() {
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, nil)
	s.reg.On("GetBySubject", s.valSch.Subject).Return(s.valSch, nil)

	sch, err := s.shl.Sync(s.ctx, s.topic, s.valCfg)
	s.Assert().NoError(err)
	s.Assert().Equal(s.valSch, sch)

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 1)
}

func (s *helperTestSuite) TestSyncCreateErr() {
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, errHelperMock)

	sch, err := s.shl.Sync(s.ctx, s.topic, s.valCfg)
	s.Assert().Equal(errHelperMock, err)
	s.Assert().Nil(sch)

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
}

func (s *helperTestSuite) TestSyncGetErr() {
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, nil)
	s.reg.On("GetBySubject", s.valSch.Subject).Return(s.valSch, errHelperMock)

	sch, err := s.shl.Sync(s.ctx, s.topic, s.valCfg)
	s.Assert().Equal(errHelperMock, err)
	s.Assert().Nil(sch)

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 1)
}

func (s *helperTestSuite) TestGet() {
	s.reg.On("GetByID", s.valSch.ID).Return(s.valSch, nil)

	sch, err := s.shl.Get(s.ctx, s.valSch.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(s.valSch, sch)

	s.reg.AssertNumberOfCalls(s.T(), "GetByID", 1)
}

func (s *helperTestSuite) TestGetErr() {
	s.reg.On("GetByID", s.valSch.ID).Return(s.valSch, errHelperMock)

	sch, err := s.shl.Get(s.ctx, s.valSch.ID)
	s.Assert().Equal(errHelperMock, err)
	s.Assert().Nil(sch)

	s.reg.AssertNumberOfCalls(s.T(), "GetByID", 1)
}

func (s *helperTestSuite) TestMarshal() {
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, nil)
	s.reg.On("GetBySubject", s.valSch.Subject).Return(s.valSch, nil)

	hData, err := s.shl.Marshal(s.ctx, s.topic, s.valCfg, s.val)
	s.Assert().NoError(err)

	sData, err := s.valSch.Marshal(s.val)
	s.Assert().NoError(err)
	s.Assert().Equal(sData, hData)

	h2Data, err := s.shl.Marshal(s.ctx, s.topic, s.valCfg, s.val)
	s.Assert().NoError(err)
	s.Assert().Equal(sData, h2Data)

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 1)
}

func (s *helperTestSuite) TestMarshalGetBySubjectErr() {
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, nil)
	s.reg.On("GetBySubject", s.valSch.Subject).Return(s.valSch, errHelperMock)

	_, err := s.shl.Marshal(s.ctx, s.topic, s.valCfg, s.val)
	s.Assert().Error(err)
	s.Assert().Equal(errHelperMock, err)

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 1)
}

func (s *helperTestSuite) TestMarshalCreateSubjectErr() {
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, errHelperMock)

	_, err := s.shl.Marshal(s.ctx, s.topic, s.valCfg, s.val)
	s.Assert().Error(err)
	s.Assert().Equal(errHelperMock, err)

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
}

func (s *helperTestSuite) TestUnmarshal() {
	s.reg.On("GetByID", s.valSch.ID).Return(s.valSch, nil)

	sData, err := s.valSch.Marshal(s.val)
	s.Assert().NoError(err)

	val := reflect.New(reflect.ValueOf(s.val).Elem().Type()).Interface()
	s.Assert().NoError(s.shl.Unmarshal(s.ctx, sData, val))
	s.Assert().Equal(s.val, val)

	s.reg.AssertNumberOfCalls(s.T(), "GetByID", 1)
}

func (s *helperTestSuite) TestUnmarshalErr() {
	s.reg.On("GetByID", s.valSch.ID).Return(s.valSch, errHelperMock)

	sData, err := s.valSch.Marshal(s.val)
	s.Assert().NoError(err)

	val := reflect.New(reflect.ValueOf(s.val).Elem().Type()).Interface()
	s.Assert().Equal(errHelperMock, s.shl.Unmarshal(s.ctx, sData, val))
	s.Assert().NotEqual(s.val, val)

	s.reg.AssertNumberOfCalls(s.T(), "GetByID", 1)
}

func (s *helperTestSuite) TestProduce() {
	s.reg.On("CreateSubject", s.keySch.Subject, *s.keySub).Return(s.keySch, nil)
	s.reg.On("GetBySubject", s.keySch.Subject).Return(s.keySch, nil)
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, nil)
	s.reg.On("GetBySubject", s.valSch.Subject).Return(s.valSch, nil)

	kmsgs := make(chan *kafka.Message, 1)
	s.prod.On("ProduceChannel").Return(kmsgs)

	err := s.shl.Produce(s.ctx, s.msg)
	s.Assert().NoError(err)

	if err == nil {
		kmsg := <-kmsgs
		s.Assert().Equal(s.msg.Topic, *kmsg.TopicPartition.Topic)
		s.Assert().Equal(*s.msg.Partition, kmsg.TopicPartition.Partition)

		vData, err := s.valSch.Marshal(s.val)
		s.Assert().NoError(err)
		s.Assert().Equal(vData, kmsg.Value)

		kData, err := s.keySch.Marshal(s.key)
		s.Assert().NoError(err)
		s.Assert().Equal(kData, kmsg.Key)
	}

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 2)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 2)
}

func (s *helperTestSuite) TestProduceCreateSubjectKeyErr() {
	s.reg.On("CreateSubject", s.keySch.Subject, *s.keySub).Return(s.keySch, errHelperMock)

	s.Assert().Equal(errHelperMock, s.shl.Produce(s.ctx, s.msg))

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
}

func (s *helperTestSuite) TestProduceGetBySubjectKeyErr() {
	s.reg.On("CreateSubject", s.keySch.Subject, *s.keySub).Return(s.keySch, nil)
	s.reg.On("GetBySubject", s.keySch.Subject).Return(s.keySch, errHelperMock)

	s.Assert().Equal(errHelperMock, s.shl.Produce(s.ctx, s.msg))

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 1)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 1)
}

func (s *helperTestSuite) TestProduceCreateSubjectValueErr() {
	s.reg.On("CreateSubject", s.keySch.Subject, *s.keySub).Return(s.keySch, nil)
	s.reg.On("GetBySubject", s.keySch.Subject).Return(s.keySch, nil)
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, nil)
	s.reg.On("GetBySubject", s.valSch.Subject).Return(s.valSch, errHelperMock)

	s.Assert().Equal(errHelperMock, s.shl.Produce(s.ctx, s.msg))

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 2)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 2)
}

func (s *helperTestSuite) TestProduceGetBySubjectValueErr() {
	s.reg.On("CreateSubject", s.keySch.Subject, *s.keySub).Return(s.keySch, nil)
	s.reg.On("GetBySubject", s.keySch.Subject).Return(s.keySch, nil)
	s.reg.On("CreateSubject", s.valSch.Subject, *s.valSub).Return(s.valSch, errHelperMock)

	s.Assert().Equal(errHelperMock, s.shl.Produce(s.ctx, s.msg))

	s.reg.AssertNumberOfCalls(s.T(), "CreateSubject", 2)
	s.reg.AssertNumberOfCalls(s.T(), "GetBySubject", 1)
}

func (s *helperTestSuite) TestFlush() {
	s.prod.On("Flush", s.tms).Return(s.tms)
	s.Assert().Equal(s.tms, s.shl.Flush(s.tms))
}

func TestHelper(t *testing.T) {
	for _, testCase := range []*helperTestSuite{
		{
			topic:  "cool-topic-name",
			keyCfg: ConfigKey,
			keySch: &Schema{
				ID:      1,
				Subject: "cool-topic-name-key",
				Schema:  ConfigKey.Schema,
			},
			key: &Key{
				Identifier: "my-test-license-key",
				Type:       "test",
			},
			valCfg: ConfigLicense,
			valSch: &Schema{
				ID:         10,
				Subject:    "cool-topic-name-value",
				Version:    1,
				References: []*Reference{},
				Schema:     ConfigLicense.Schema,
			},
			val: &License{
				Identifier: "my-test-license",
				Name:       "Test License",
				URL:        "https://test-license.com",
			},
		},
		{
			topic:  "another-cool-topic-name",
			keyCfg: ConfigKey,
			keySch: &Schema{
				ID:      1,
				Subject: "another-cool-topic-name-key",
				Schema:  ConfigKey.Schema,
			},
			key: &Key{
				Identifier: "my-test-event-key",
				Type:       "test",
			},
			valCfg: ConfigEvent,
			valSch: &Schema{
				ID:         21,
				Subject:    "another-cool-topic-name-value",
				Version:    100,
				References: []*Reference{},
				Schema:     ConfigEvent.Schema,
			},
			val: &Event{
				Identifier: "event-id",
				Type:       "test",
				FailCount:  10,
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
