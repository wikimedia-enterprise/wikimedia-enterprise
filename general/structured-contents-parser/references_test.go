package parser

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/suite"
)

type referencesTestSuite struct {
	suite.Suite
}

func (s *referencesTestSuite) TestBogusCOinSMetadata() {
	// If it doesn't look like actual metadata, make sure we don't fail or falsely assert it's COinS.
	s.Assert().Empty(ParseCOinSMetadata(""))
	s.Assert().Empty(ParseCOinSMetadata("word"))
	s.Assert().Empty(ParseCOinSMetadata("=word"))
	s.Assert().Empty(ParseCOinSMetadata("&"))
	s.Assert().Empty(ParseCOinSMetadata("="))
	s.Assert().Empty(ParseCOinSMetadata("=&&===&=&&&"))
}

func (s *referencesTestSuite) TestCOinSExtraction() {

	type testCase struct {
		markup           string
		expectedType     string
		expectedMetadata map[string]string
		expectedIsCite   bool
	}
	cases := []testCase{
		{
			markup: `<div><cite class="bogus citation"></cite><b>no class after citation</b></div>`,
		},
		{
			markup:         `<div><cite class="bogus citation cs1"></cite><span>no title</span></div>`,
			expectedType:   "cs1",
			expectedIsCite: true,
		},
		{
			markup:         `<div><cite class="bogus citation cs1"></cite>no span after cite</div>`,
			expectedType:   "cs1",
			expectedIsCite: true,
		},
		{
			markup:         `<div><cite class="bogus citation cs1"></cite><span title="nometadata"></span></div>`,
			expectedType:   "cs1",
			expectedIsCite: true,
		},
		{
			markup:       `<div><cite class="bogus citation cs1"></cite><span title="key=value"></span></div>`,
			expectedType: "cs1",
			expectedMetadata: map[string]string{
				"key":      "value",
				"is_coins": "true",
			},
			expectedIsCite: true,
		},
	}

	for _, testCase := range cases {
		// For realism, wrap in the usual markup we expect.
		markup := fmt.Sprintf(`<ol class="references"><li id="refid"><span class="reference-text">%s</span></li></ol>`, testCase.markup)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
		if err != nil {
			log.Panicln(err)
		}

		refType, metadata, isCite := ExtractCOinSTypeAndMetadata(doc.Find("span"))
		s.Assert().Equal(testCase.expectedType, refType)
		if testCase.expectedMetadata != nil {
			s.Assert().Equal(testCase.expectedMetadata, metadata)
		} else {
			s.Assert().Empty(metadata)
		}
		s.Assert().Equal(testCase.expectedIsCite, isCite)
	}
}

func (s *referencesTestSuite) TestParentReferenceDetectedFromParsoidDataMw() {
	// Simulates Parsoid HTML where cite_note-3 is a child of cite_note-2
	markup := `
<head>
  <base href="https://de.wikipedia.org/">
</head>
<body>
<p>
  <sup about="#mwt1" class="mw-ref reference" data-mw="{&quot;name&quot;:&quot;ref&quot;,&quot;attrs&quot;:{},&quot;body&quot;:{},&quot;isSubRefWithMainBody&quot;:1,&quot;mainBody&quot;:&quot;mw-reference-text-cite_note-2&quot;}">
    <a href="#cite_note-3"><span class="mw-linkback-text">↑</span></a>
  </sup>
</p>
<ol class="references">
  <li id="cite_note-2">
    <span id="mw-reference-text-cite_note-2" class="mw-reference-text">
      <cite class="citation book cs1">Michael Walter: Enge Oligopole und Wettbewerbspolitik: Gesetzmäßigkeiten des Videospiel- und des Mobilfunkmarktes. Springer, 2013, ISBN 978-3-322-81047-2.</cite>
      <span title="rft.genre=book"></span>
      <span class="cs1-format">S. 51–57.</span>
    </span>
  </li>
  <li id="cite_note-3">
    <span class="mw-reference-text">S. 53.</span>
  </li>
</ol>
</body>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
	s.Require().NoError(err)

	var p Parser
	refs := p.GetReferences(doc.Selection)
	var r2, r3 *Reference
	for _, ref := range refs {
		switch ref.Identifier {
		case "cite_note-2":
			r2 = ref
		case "cite_note-3":
			r3 = ref
		}
	}
	s.Require().NotNil(r2, "cite_note-2 should be present")
	s.Require().NotNil(r3, "cite_note-3 should be present")

	s.Equal("book", r2.Type)
	if s.Assert().NotNil(r2.Text) {
		s.Contains(r2.Text.Value, "S. 51–57.")
	}

	// Parsoid data-mw tells us r3 is a sub-reference of r2
	s.Equal("cite_note-2", r3.ParentReference)
	if s.Assert().NotNil(r3.Text) {
		s.Equal("S. 53.", strings.TrimSpace(r3.Text.Value))
	}
}

func TestRecovery(t *testing.T) {
	suite.Run(t, new(referencesTestSuite))
}
