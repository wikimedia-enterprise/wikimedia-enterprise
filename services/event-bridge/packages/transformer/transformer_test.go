package transformer_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/transformer"

	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

type transformerTestSuite struct {
	suite.Suite
	evt *eventstream.PageChange
	dt  time.Time
}

// SetupTest initializes the article builder and seeds a random number generator.
func (s *transformerTestSuite) SetupTest() {

	s.evt = &eventstream.PageChange{}

	s.dt = time.Now()

	s.evt.Data.Performer.UserID = 1
	s.evt.Data.Performer.UserText = "usertext"
	s.evt.Data.Performer.UserEditCount = 1
	s.evt.Data.Performer.UserGroups = []string{"group1", "group2"}
	s.evt.Data.Performer.UserIsBot = true
	s.evt.Data.Performer.UserRegistrationDt = s.dt

	s.evt.Data.Revision.RevID = 1
	s.evt.Data.Revision.Comment = "comment"
	s.evt.Data.Revision.IsMinorEdit = true
	s.evt.Data.Revision.RevSize = 1000

	s.evt.Data.Database = "commonswiki"

	s.evt.Data.Meta.URI = "http://commonswiki.wikipedia.org"

	s.evt.Data.Meta.Domain = "commonswiki.wikipedia.org"

	s.evt.Data.Page.PageID = 12
	s.evt.Data.Page.PageTitle = "title"
	s.evt.Data.Revision.RevDt = s.dt

	s.evt.Data.Page.PageNamespace = 6

	s.evt.Data.Revision.IsContentVisible = true
	s.evt.Data.Revision.IsEditorVisible = true
	s.evt.Data.Revision.IsCommentVisible = true

}

func (s *transformerTestSuite) TestTrasformerPageChangeToArticle() {
	t := transformer.New()

	art, err := t.PageChangeToArticle(context.Background(), &langid.Dictionary{}, s.evt)

	s.Assert().NoError(err)

	s.Assert().Equal(s.evt.Data.Performer.UserID, art.Version.Editor.Identifier)
	s.Assert().Equal(s.evt.Data.Performer.UserText, art.Version.Editor.Name)
	s.Assert().Equal(s.evt.Data.Performer.UserEditCount, art.Version.Editor.EditCount)
	s.Assert().Equal(s.evt.Data.Performer.UserGroups, art.Version.Editor.Groups)
	s.Assert().Equal(s.evt.Data.Performer.UserIsBot, art.Version.Editor.IsBot)
	s.Assert().Equal(&s.evt.Data.Performer.UserRegistrationDt, art.Version.Editor.DateStarted)

	s.Assert().Equal(s.evt.Data.Revision.RevID, art.Version.Identifier)
	s.Assert().Equal(s.evt.Data.Revision.Comment, art.Version.Comment)
	s.Assert().Equal(s.evt.Data.Revision.IsMinorEdit, art.Version.IsMinorEdit)
	s.Assert().Equal(float64(s.evt.Data.Revision.RevSize), art.Version.Size.Value)

	s.Assert().Equal(s.evt.Data.Page.PageID, art.Identifier)
	s.Assert().Equal(s.evt.Data.Page.PageTitle, art.Name)
	s.Assert().Equal(&s.evt.Data.Revision.RevDt, art.DateCreated)
	s.Assert().Equal(&s.evt.Data.Revision.RevDt, art.DateModified)
	s.Assert().Equal(s.evt.Data.Database, art.IsPartOf.Identifier)
	s.Assert().Equal(fmt.Sprintf("https://%s", s.evt.Data.Meta.Domain), art.IsPartOf.URL)
	s.Assert().Equal(s.evt.Data.Page.PageNamespace, art.Namespace.Identifier)
	s.Assert().Equal("en", art.InLanguage.Identifier)
	s.Assert().Equal(s.evt.Data.Revision.IsCommentVisible, art.Visibility.Comment)
	s.Assert().Equal(s.evt.Data.Revision.IsEditorVisible, art.Visibility.Editor)
	s.Assert().Equal(s.evt.Data.Revision.IsContentVisible, art.Visibility.Text)

	url, err := url.QueryUnescape(s.evt.Data.Meta.URI)
	s.Assert().NoError(err)
	s.Assert().Equal(url, art.URL)

}

func TestArticleBuilder(t *testing.T) {
	suite.Run(t, new(transformerTestSuite))
}
