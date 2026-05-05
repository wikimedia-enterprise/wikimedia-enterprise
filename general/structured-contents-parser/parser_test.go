package parser

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testFiles = []string{
	"(Don%27t_Go_Back_To)_Rockville",
	".NET_Foundation",
	"1976_Democratic_Party_presidential_primaries",
	"ABC News",
	"Al_Messila_(Doha)",
	"Albert_Einstein",
	"Aldi",
	"Alexandros_Ioannis_Ginnis",
	"Anarchisme",
	"Angela_Merkel",
	"Armonium",
	"Azerbaijan",
	"Benjamin_Franklin",
	"Berliner_Mauer",
	"Body mass index",
	"Brasilia",
	"Brian May",
	"btopa",
	"Cactus",
	"Caryophyllales",
	"Charles_Darwin",
	"China",
	"Ciro_Immobile",
	"Coset",
	"Cot_side",
	"Darth_Vader",
	"David_Cameron",
	"Distinctive_feature",
	"Eurovision Song Contest 2023",
	"FR_Josephine_Baker",
	"Federico_II_di_Prussia",
	"FetLife",
	"Financial_statement",
	"Foobar",
	"Francois_Lapous",
	"Free trade agreements of the European Union",
	"Gad_Elmaleh",
	"Ganges",
	"Google Chrome",
	"Grecia",
	"Grimes",
	"Gwen Stefani",
	"ISO_IEC_2022",
	"ISO_IEC_8859-2",
	"International_Music_Score_Library_Project",
	"January-February_2019_North_American_cold_wave",
	"Jose_Eduardo_Dos_Santos",
	"Josephine_Baker",
	"Kane (noble family)",
	"Kaori_Icho",
	"knetka",
	"Kyoto",
	"Lionel_Messi",
	"London",
	"Main_page",
	"Malta",
	"Matrioska",
	"Matrjoschka",
	"Matryoshka_doll",
	"Mike_Pence",
	"Mike_Pompeo",
	"Mino_Raiola",
	"Moores_law",
	"Mumbai",
	"Narendra_Modi",
	"Nephrite",
	"O",
	"Organic_acid_anhydride",
	"Otto_Warmbier",
	"Ottoman_Empire",
	"Proceedings_of_the_Royal_Irish_Academy",
	"Roger_Federer",
	"Rougequeue_de_Hodgson",
	"S.L._Benfica",
	"Scythian_languages",
	"Shakira",
	"Slogan",
	"Spain",
	"Switzerland",
	"The Futurist",
	"Thesaurus",
	"Toyota",
	"United_States",
	"Vereinigte_Stahlwerke",
	"Vitis",
	"William_Shakespeare",
	"World_War_II",
	"Yoshua_Bengio",
	"Zillow",
	"Espa%C3%B1a",
	"Zypern",
	"Faixa_de_Gaza",
	"Roma_Garden_Open_2025",
	"Associazione_Sportiva_Pizzighettone",
	"Glossary_of_chess",
	"Outline_of_chess",
	"List_of_lists_of_lists",
	"USS_Nevada_(BB-36)",
	"Josephine_Baker",
	"Cultural_depictions_of_George_Washington",
	"Garden_Open_2025",
}

type parserTestSuite struct {
	suite.Suite
	ttl string
	prs API
	sel *goquery.Selection
	cup *cupaloy.Config
}

const ConfidenceThreshold = 0.5

func (s *parserTestSuite) SetupSuite() {
	fle, err := os.Open(fmt.Sprintf("./testdata/%s.html", s.ttl))
	s.Require().NoError(err)

	doc, err := goquery.NewDocumentFromReader(fle)
	s.Require().NoError(err)

	s.sel = doc.Selection

	s.prs = NewWithFlag(ConfidenceThreshold)

	subdir := "zSnapshot"

	// Make sure directory exists
	_ = os.MkdirAll(subdir, os.ModePerm)

	// Configure cupaloy
	s.cup = cupaloy.New(cupaloy.SnapshotSubdirectory(subdir))
}

func (s *parserTestSuite) TestGetLists() {
	scs, _, _, _ := s.prs.GetArticleParts(s.sel)

	var lists []*Part

	var walk func(parts []*Part)
	walk = func(parts []*Part) {
		for _, p := range parts {
			if p.Type == PartTypeList || p.Type == PartTypeOrderedList {
				lists = append(lists, p)
			}
			walk(p.HasParts)
		}
	}

	walk(scs)

	s.cup.SnapshotT(s.T(), lists)
}

func (s *parserTestSuite) TestGetImages() {
	ims := s.prs.GetImages(s.sel)
	s.cup.SnapshotT(s.T(), ims)
}

func TestIsFileSpan(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`
		<div>
			<span typeof="mw:File"><img src="x"/></span>
			<span typeof="mw:File/Frameless"><img src="y"/></span>
			<span typeof="mw:File/Thumb"><img src="z"/></span>
			<span typeof="mw:File/Frame"><img src="w"/></span>
			<span typeof="mw:Extension/ref">ref</span>
		</div>
	`))
	expected := map[string]bool{
		"mw:File":           true,
		"mw:File/Frameless": true,
		"mw:File/Thumb":     false,
		"mw:File/Frame":     false,
		"mw:Extension/ref":  false,
	}
	spans := doc.Find("span")
	spans.Each(func(i int, sel *goquery.Selection) {
		typeof := sel.AttrOr("typeof", "")
		got := isFileSpan(sel)
		want, ok := expected[typeof]
		assert.True(t, ok, "unexpected typeof %q for span %d", typeof, i)
		assert.Equal(t, want, got, "span %d typeof=%q", i, typeof)
	})
}

func TestCaptionFromDataMw(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(
		`<span data-mw="{&quot;caption&quot;:&quot;George Washington's signature&quot;}">x</span>` +
			`<span data-mw="{}">y</span>`,
	))
	withCaption := doc.Find("span").First()
	if got := captionFromDataMw(withCaption); got != "George Washington's signature" {
		t.Errorf("captionFromDataMw with caption: got %q", got)
	}
	empty := doc.Find("span").Last()
	if got := captionFromDataMw(empty); got != "" {
		t.Errorf("captionFromDataMw empty: got %q", got)
	}
}

func (s *parserTestSuite) TestGetText() {
	hsh := md5.New()
	_, err := hsh.Write([]byte(s.prs.GetText(s.sel)))
	s.Assert().NoError(err)

	s.cup.SnapshotT(s.T(), hsh.Sum(nil))
}

func (s *parserTestSuite) TestGetLinks() {
	lks := s.prs.GetLinks(s.sel, "")
	s.cup.SnapshotT(s.T(), lks)
}

func (s *parserTestSuite) TestGetInfoBoxes() {
	ibs := s.prs.GetInfoBoxes(s.sel)
	s.cup.SnapshotT(s.T(), ibs)
}

func (s *parserTestSuite) TestGetArticleParts() {
	scs, tbls, _, _ := s.prs.GetArticleParts(s.sel)

	// Round confidence scores to 4 decimal places to avoid snapshot diffs
	for _, tbl := range tbls {
		tbl.Table.ConfidenceScore = roundFloat(tbl.Table.ConfidenceScore, 4)
	}

	s.cup.SnapshotT(s.T(), map[string]interface{}{
		"sections": scs,
		"tables":   tbls,
	})
}

func (s *parserTestSuite) TestListPartsOnly() {
	scs, _, _, _ := s.prs.GetArticleParts(s.sel)

	var lists []*Part

	var collect func(parts []*Part)
	collect = func(parts []*Part) {
		for _, p := range parts {
			if p.Type == PartTypeList || p.Type == PartTypeOrderedList || p.Type == PartTypeDefinitionList {
				lists = append(lists, p)
			}
			collect(p.HasParts)
		}
	}

	collect(scs)

	s.cup.SnapshotT(s.T(), lists)
}

// Helper function to round a float to `precision` digits
func roundFloat(val float64, precision int) float64 {
	format := fmt.Sprintf("%%.%df", precision)
	str := fmt.Sprintf(format, val)
	rounded, _ := strconv.ParseFloat(str, 64)
	return rounded
}

func (s *parserTestSuite) TestGetReferences() {
	scs := s.prs.GetReferences(s.sel)
	s.cup.SnapshotT(s.T(), scs)
}

func (s *parserTestSuite) TestGetTablesOnly() {
	_, tbls, _, _ := s.prs.GetArticleParts(s.sel)

	// Round confidence scores to 4 decimal places
	for _, tbl := range tbls {
		tbl.Table.ConfidenceScore = roundFloat(tbl.Table.ConfidenceScore, 4)
	}

	s.cup.SnapshotT(s.T(), tbls)
}

func (s *parserTestSuite) TestGetTablesOnlyJSON() {
	_, tbls, _, _ := s.prs.GetArticleParts(s.sel)

	// Round confidence scores to 4 decimal places
	for _, tbl := range tbls {
		tbl.Table.ConfidenceScore = roundFloat(tbl.Table.ConfidenceScore, 4)
	}

	out, err := json.MarshalIndent(tbls, "", "  ")
	s.NoError(err)

	outStr := string(out)
	outStr = s.ttl + "\n" + outStr
	s.cup.SnapshotT(s.T(), outStr)
}

func (s *parserTestSuite) TestTableIDsInParts() {
	scs, _, _, _ := s.prs.GetArticleParts(s.sel)

	var invalid []string

	var check func(parts []*Part)
	check = func(parts []*Part) {
		for _, p := range parts {
			if p.Type == PartTypeTable && len(p.TableReferences) == 0 {
				invalid = append(invalid, fmt.Sprintf("Missing TableID in part: %+v", p))
			}
			check(p.HasParts)
		}
	}
	check(scs)

	s.Require().Empty(invalid, "Some tables have empty TableID")
}

func (s *parserTestSuite) TestAllTableIDsAreUnique() {
	_, tables, _, _ := s.prs.GetArticleParts(s.sel)

	seen := make(map[string]bool)
	dupes := []string{}

	for _, tbl := range tables {
		id := tbl.ID
		if seen[id] {
			dupes = append(dupes, id)
		}
		seen[id] = true
	}

	s.Require().Empty(dupes, "Duplicate table IDs found: %v", dupes)
}

func (s *parserTestSuite) TestEachTableHasHeadersOrRows() {
	_, tables, _, _ := s.prs.GetArticleParts(s.sel)

	for _, tbl := range tables {
		hasHeaders := len(tbl.Table.Headers) > 0
		hasRows := len(tbl.Table.Rows) > 0

		s.Assert().True(hasHeaders || hasRows, "Table %s has neither headers nor rows", tbl.ID)
	}
}

func (s *parserTestSuite) TestAddCellsForState_Pizzighettone() {
	// Only run this for Associazione_Sportiva_Pizzighettone
	if s.ttl != "Associazione_Sportiva_Pizzighettone" {
		s.T().Skip("Skipping: only relevant for Associazione_Sportiva_Pizzighettone")
	}

	_, tbls, _, _ := s.prs.GetArticleParts(s.sel)

	var found bool
	for _, tbl := range tbls {
		if strings.Contains(tbl.ID, "campionati_nazionali_table1") {
			found = true

			s.Require().Equal(5, len(tbl.Table.Rows),
				"Expected 5 rows in campionati_nazionali_table1, got %d", len(tbl.Table.Rows))

			// Snapshot for future regressions
			//s.cup.SnapshotT(s.T(), tbl.Table)
		}
	}
	s.Require().True(found, "campionati_nazionali_table1 not found in parsed tables")
}

func (s *parserTestSuite) TestNoEmptyCellsInFirstRow() {
	_, tables, _, _ := s.prs.GetArticleParts(s.sel)

	for _, tbl := range tables {
		if len(tbl.Table.Rows) == 0 {
			continue
		}
		first := tbl.Table.Rows[0]
		nonEmpty := false
		for _, cell := range first {
			if strings.TrimSpace(cell.Value) != "" || len(cell.Inner) > 0 {
				nonEmpty = true
				break
			}
		}
		s.Assert().True(nonEmpty, "First row of table %s is entirely empty", tbl.ID)
	}
}

func (s *parserTestSuite) TestTableIDsExistInTablesList() {
	scs, tbls, _, _ := s.prs.GetArticleParts(s.sel)

	allTableIDs := make(map[string]struct{}, len(tbls))
	for _, tbl := range tbls {
		allTableIDs[tbl.ID] = struct{}{}
	}

	var missing []string
	var check func(parts []*Part)
	check = func(parts []*Part) {
		for _, p := range parts {
			for _, ref := range p.TableReferences {
				//Replace 0.5 with conf score
				if _, ok := allTableIDs[ref.Identifier]; !ok && ref.ConfidenceScore > ConfidenceThreshold {
					missing = append(missing, fmt.Sprintf("Part: %s, Missing ID: %s", p.Name, ref.Identifier))
				}
			}
			check(p.HasParts)
		}
	}
	check(scs)

	s.Require().Empty(missing, "Some parts refer to missing table IDs: %v", missing)
}

func TestParser(t *testing.T) {
	for _, title := range testFiles {
		suite.Run(t, &parserTestSuite{
			ttl: title,
		})
	}
}

var benchRefsSink any // prevents DCE from messing with benchmark

func BenchmarkReferences(b *testing.B) {
	RunBenchmark(b, func(parser API, sel *goquery.Selection) {
		benchRefsSink = parser.GetReferences(sel)
	})
}

func BenchmarkParts(b *testing.B) {
	RunBenchmark(b, func(parser API, sel *goquery.Selection) {
		benchRefsSink, benchRefsSink, benchRefsSink, _ = parser.GetArticleParts(sel)
	})
}

func RunBenchmark(b *testing.B, benchmark func(API, *goquery.Selection)) {
	b.ReportAllocs()

	prs := NewWithFlag(ConfidenceThreshold)
	selections := make([]*goquery.Selection, 0, len(testFiles))
	for _, ttl := range testFiles {
		path := fmt.Sprintf("./testdata/%s.html", ttl)

		data, err := os.ReadFile(path)
		if err != nil {
			b.Fatalf("read %s: %v", path, err)
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("parse %s: %v", path, err)
		}

		selections = append(selections, doc.Selection)
	}
	if len(selections) == 0 {
		b.Fatalf("no test files")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sel := selections[i%len(selections)]
		benchmark(prs, sel)
	}
}

func (s *parserTestSuite) TestIsEmptyMarkdownList() {
	parser := s.prs.(*Parser)
	var b bytes.Buffer
	results := []map[string]interface{}{}

	s.sel.Find("ul, ol, dl").Each(func(i int, sel *goquery.Selection) {
		if len(sel.Nodes) == 0 {
			return
		}
		b.Reset()
		parser.getList(&b, sel.Nodes[0], 1)
		markdown := b.String()

		results = append(results, map[string]interface{}{
			"index":   i,
			"isEmpty": isEmptyMarkdownList(markdown),
			"preview": func() string {
				if len(markdown) > 100 {
					return markdown[:100] + "..."
				}
				return markdown
			}(),
		})
	})
	assert.True(s.T(), isEmptyMarkdownList("  - \n  - "))
	assert.False(s.T(), isEmptyMarkdownList("  - Item"))
}

func TestHasTextContent(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{"empty div", "<div></div>", false},
		{"non-text content only", "<li><img src=\"x\"/></li>", false},
		{"div with text", "<div>text</div>", true},
		{"nested text", "<li><span>hello</span></li>", true},
		{"whitespace only", "<li>   </li>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			node := doc.Find("body").Children().First()
			if !assert.NotEmpty(t, node.Nodes, "no node parsed for input: %s", tt.html) {
				return
			}
			assert.Equal(t, tt.expected, parser.hasTextContent(node.Nodes[0]))
		})
	}
}
