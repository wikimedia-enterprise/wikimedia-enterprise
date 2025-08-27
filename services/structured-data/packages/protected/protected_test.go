package protected_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/protected"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type protectedTestSuite struct {
	suite.Suite
	wmk *protectedAPIMock
	prt *protected.Protected
	dtb string
	ttl string
	res bool
}

type protectedAPIMock struct {
	mock.Mock
	wmf.API
}

func (m *protectedAPIMock) GetPage(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*wmf.Page, error) {
	ags := m.Called(dtb, ttl)

	pge := ags.Get(0)
	if pge == nil {
		return nil, ags.Error(1)
	} else {
		return pge.(*wmf.Page), ags.Error(1)
	}
}

func (t *protectedTestSuite) SetupTest() {
	t.wmk = new(protectedAPIMock)
	t.prt = protected.New(t.wmk)
}

func (t *protectedTestSuite) TestFailsToLoadPagesIsProtected() {
	t.wmk.On("GetPage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("Failed to get page")).Times(3)

	t.Assertions.False(t.prt.IsProtectedPage(t.dtb, t.ttl))
}

func (t *protectedTestSuite) TestIsProtected() {
	rev := &wmf.Revision{Slots: &wmf.Slots{Main: &wmf.Main{Content: "BOYL"}}}
	t.wmk.On("GetPage", mock.Anything, mock.Anything, mock.Anything).Return(&wmf.Page{Revisions: []*wmf.Revision{rev}}, nil).Once()

	t.Assertions.Equal(t.res, t.prt.IsProtectedPage(t.dtb, t.ttl))
}

func TestAggregate(t *testing.T) {
	for _, testCase := range []*protectedTestSuite{
		{
			dtb: "ruwiki",
			ttl: "BOYL",
			res: true,
		},
		{
			dtb: "ruwiki",
			ttl: "NotProtected",
			res: false,
		},
		{
			dtb: "enwiki",
			ttl: "BOYL",
			res: false,
		},
	} {
		suite.Run(t, testCase)
	}
}
