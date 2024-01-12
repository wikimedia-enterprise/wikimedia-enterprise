package namespaces_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"wikimedia-enterprise/services/bulk-ingestion/handlers/namespaces"
)

type unmarshalProducerMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

func (m *unmarshalProducerMock) Produce(_ context.Context, msgs ...*schema.Message) error {
	msg := msgs[0]
	return m.Called(msg).Error(0)
}

func (m *unmarshalProducerMock) Flush(time int) int {
	return 0
}

type wmfAPI struct {
	wmf.API
	mock.Mock
}

func (m *wmfAPI) GetNamespaces(_ context.Context, dtb string, ops ...func(*url.Values)) ([]*wmf.Namespace, error) {
	ags := m.Called(dtb)
	return ags.Get(0).([]*wmf.Namespace), ags.Error(1)
}

type S3API struct {
	s3iface.S3API
	mock.Mock
}

func (m *S3API) PutObjectWithContext(_ context.Context, input *s3.PutObjectInput, ops ...request.Option) (*s3.PutObjectOutput, error) {
	ags := m.Called(input)
	return ags.Get(0).(*s3.PutObjectOutput), ags.Error(1)
}

type namespaceTestSuite struct {
	suite.Suite
	ctx           context.Context
	req           *pb.NamespacesRequest
	stream        *unmarshalProducerMock
	s3Err         error
	wmfError      error
	producerError error
	params        namespaces.Parameters
}

func (p *namespaceTestSuite) SetupSuite() {
	p.ctx = context.Background()
	str := &unmarshalProducerMock{}
	str.On("Produce", mock.Anything).Return(p.producerError)
	wmfInit := &wmfAPI{}
	wmfInit.On("GetNamespaces", mock.Anything).Return([]*wmf.Namespace{
		{
			ID:   0,
			Name: "Main",
		},
	}, p.wmfError)

	p.req = &pb.NamespacesRequest{}
	s3Init := &S3API{}
	s3Init.On("PutObjectWithContext", mock.Anything).Return(&s3.PutObjectOutput{}, p.s3Err)
	p.stream = str

	p.params = namespaces.Parameters{
		Stream: p.stream,
		API:    wmfInit,
		Env: &env.Environment{
			TopicNamespaces: "aws.bulk-ingestion.namespaces.v1",
			DefaultProject:  "enwiki",
		},
		S3: s3Init,
	}
}

func (p *namespaceTestSuite) TestNamespaces() {
	r, err := namespaces.Handler(p.ctx, &p.params, p.req)

	if p.producerError != nil {
		p.Assert().Nil(r)
		p.Assert().Equal(err, p.producerError)
		return
	} else if p.wmfError != nil {
		p.Assert().Nil(r)
		p.Assert().Equal(err, p.wmfError)
		return
	} else if p.s3Err != nil {
		p.Assert().Nil(r)
		p.Assert().Equal(err, p.s3Err)
		return
	} else {
		p.Assert().NotNil(r)
		p.Assert().NoError(err)
		p.stream.AssertNumberOfCalls(p.T(), "Produce", 1)
		p.Assert().Equal(r.Total, int32(1))
		p.Assert().Equal(int(r.Errors), 0)
	}
}

func Test(t *testing.T) {
	testCases := []*namespaceTestSuite{
		{
			producerError: nil,
			wmfError:      nil,
			s3Err:         nil,
		},
		{
			producerError: fmt.Errorf("Error in producing message"),
			wmfError:      nil,
			s3Err:         nil,
		},
		{
			producerError: nil,
			wmfError:      fmt.Errorf("Error in getting languages"),
			s3Err:         nil,
		},
		{
			producerError: nil,
			wmfError:      nil,
			s3Err:         fmt.Errorf("Error in putting object in S3"),
		},
	}

	for _, tc := range testCases {
		suite.Run(t, tc)
	}
}