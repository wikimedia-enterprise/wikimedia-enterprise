package schema

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type newMockTestSuite struct {
	suite.Suite
}

func (s *newMockTestSuite) TestNew() {
	mck, err := NewMock()
	s.Assert().NoError(err)
	s.Assert().NotNil(mck)
}

func TestNewMock(t *testing.T) {
	suite.Run(t, new(newMockTestSuite))
}

type prodMock struct {
	mock.Mock
}

func (m *prodMock) Produce(_ context.Context, mgs ...*Message) error {
	return m.Called(mgs).Error(0)
}

func (m *prodMock) Flush(tms int) int {
	return m.Called(tms).Int(0)
}

type mockTestSuite struct {
	suite.Suite
	ctx context.Context
	prd *prodMock
	mgs []*Message
	mck *Mock
	tps []*MockTopic
	err error
}

func (s *mockTestSuite) SetupSuite() {
	s.ctx = context.Background()

	s.prd = new(prodMock)
	s.prd.On("Produce", s.mgs).Return(s.err)
	s.prd.On("Flush", 1000).Return(0)

	s.mck = &Mock{
		Producer: s.prd,
	}
}

func (s *mockTestSuite) TestRun() {
	err := s.mck.Run(s.ctx, s.tps...)
	s.Assert().Equal(s.err, err)
}

func TestMock(t *testing.T) {
	pdl := "{\"identifier\":\"/enwiki/Earth\", \"type\":\"articles\"}\n{\"name\": \"Earth\"}"
	rdr := strings.NewReader(pdl)
	tps := []*MockTopic{
		{
			Topic:  "articles",
			Config: ConfigArticle,
			Type:   Article{},
			Reader: rdr,
		},
	}

	tpp := []*MockTopic{
		{
			Topic:      "articles",
			Partitions: []int32{0, 1},
			Config:     ConfigArticle,
			Type:       Article{},
			Reader:     rdr,
		},
	}

	ptn := kafka.PartitionAny

	mgs := []*Message{
		{
			Topic:     "articles",
			Partition: &ptn,
			Key: &Key{
				Identifier: "/enwiki/Earth",
				Type:       "articles",
			},
			Value: &Article{
				Name: "Earth",
			},
			Config: ConfigArticle,
		},
	}

	pt0 := int32(0)
	pt1 := int32(1)

	mgp := []*Message{
		{
			Topic:     "articles",
			Partition: &pt0,
			Key: &Key{
				Identifier: "/enwiki/Earth",
				Type:       "articles",
			},
			Value: &Article{
				Name: "Earth",
			},
			Config: ConfigArticle,
		},
		{
			Topic:     "articles",
			Partition: &pt1,
			Key: &Key{
				Identifier: "/enwiki/Earth",
				Type:       "articles",
			},
			Value: &Article{
				Name: "Earth",
			},
			Config: ConfigArticle,
		},
	}

	for _, testCase := range []*mockTestSuite{
		{
			tps: tps,
			mgs: mgs,
			err: nil,
		},
		{
			tps: tpp,
			mgs: mgp,
			err: nil,
		},
		{
			tps: tps,
			mgs: mgs,
			err: errors.New("produce error"),
		},
		{
			tps: tpp,
			mgs: mgp,
			err: errors.New("produce error"),
		},
	} {
		suite.Run(t, testCase)
		rdr.Reset(pdl)
	}
}
