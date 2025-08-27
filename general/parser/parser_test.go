package parser

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"fmt"
	"sort"

	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/suite"
)

type templatesTestSuite struct {
	suite.Suite
	tps Templates
	nms []string
	sbs []string
	hss bool
}

func (s *templatesTestSuite) TestGetNames() {
	nms := s.tps.GetNames()
	s.Assert().Equal(s.nms, nms)
}

func (s *templatesTestSuite) TestHasSubstringInNames() {
	s.Assert().Equal(s.hss, s.tps.HasSubstringInNames(s.sbs))
}

func TestTemplates(t *testing.T) {
	for _, testCase := range []*templatesTestSuite{
		{
			tps: Templates{
				&Template{Name: "Template 1"},
				&Template{Name: "Template 2"},
				&Template{Name: "Template 3"},
				&Template{Name: "Template:NOINDEX"},
			},
			nms: []string{"Template 1", "Template 2", "Template 3", "Template:NOINDEX"},
			sbs: []string{"noindex", "pas index"},
			hss: true,
		},
		{
			tps: Templates{},
			nms: []string{},
		},
	} {
		suite.Run(t, testCase)
	}
}

type categoriesTestSuite struct {
	suite.Suite
	cts Categories
	nms []string
	sbs []string
	hss bool
}

func (s *categoriesTestSuite) TestGetNames() {
	nms := s.cts.GetNames()
	s.Assert().Equal(s.nms, nms)
}

func (s *categoriesTestSuite) TestHasSubstringInNames() {
	s.Assert().Equal(s.hss, s.cts.HasSubstringInNames(s.sbs))
}

func TestCategories(t *testing.T) {
	for _, testCase := range []*categoriesTestSuite{
		{
			cts: Categories{
				&Category{Name: "Category 1"},
				&Category{Name: "Category 2"},
				&Category{Name: "Category 3"},
				&Category{Name: "Category:Noindexed articles"},
			},
			nms: []string{"Category 1", "Category 2", "Category 3", "Category:Noindexed articles"},
			sbs: []string{"noindex"},
			hss: true,
		},
		{
			cts: Categories{},
			nms: []string{},
		},
	} {
		suite.Run(t, testCase)
	}
}

type parserTestSuite struct {
	suite.Suite
	ttl string
	prs *Parser
	sel *goquery.Selection
	cup *cupaloy.Config
}

func (s *parserTestSuite) SetupSuite() {
	fle, err := os.Open(fmt.Sprintf("./testdata/%s.html", s.ttl))
	s.Assert().NoError(err)

	doc, err := goquery.NewDocumentFromReader(fle)
	s.Assert().NoError(err)

	s.sel = doc.Selection
	s.prs = &Parser{}

	s.cup = cupaloy.New(cupaloy.SnapshotSubdirectory("zSnapshot"))
}

func (s *parserTestSuite) sortRelations(rls Relations) {
	sort.Slice(rls, func(i, j int) bool {
		return rls[i].Name < rls[j].Name
	})
}

func (s *parserTestSuite) TestGetTemplates() {
	tps := s.prs.GetTemplates(s.sel)

	s.sortRelations(tps)

	s.cup.SnapshotT(s.T(), tps)
}

func (s *parserTestSuite) TestGetCategories() {
	cts := s.prs.GetCategories(s.sel)

	s.sortRelations(cts)

	s.cup.SnapshotT(s.T(), cts)
}

func (s *parserTestSuite) TestGetAbstract() {
	abs, err := s.prs.GetAbstract(s.sel)
	s.Assert().NoError(err)

	s.cup.SnapshotT(s.T(), abs)
}
func (s *parserTestSuite) TestGetImages() {
	ims := s.prs.GetImages(s.sel)

	s.cup.SnapshotT(s.T(), ims)
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

func (s *parserTestSuite) TestGetSections() {
	scs := s.prs.GetSections(s.sel)

	s.cup.SnapshotT(s.T(), scs)
}

func (s *parserTestSuite) TestReplaceLists() {
	var b bytes.Buffer
	before := s.sel.Clone()

	// Replicating the 'replace' code and testing getList
	s.prs.findFirstLevel(before, "ul, ol, dl").Each(func(_ int, sli *goquery.Selection) {
		b.Reset()
		b.WriteString("\n")
		s.prs.getList(&b, sli.Nodes[0], 1)
		md := b.String()
		sli.ReplaceWithHtml(md)
	})

	beforeHTML, _ := before.Html()

	s.prs.ReplaceLists(s.sel)
	afterHTML, _ := s.sel.Html()

	s.Assert().Equal(beforeHTML, afterHTML)
}

func TestParser(t *testing.T) {
	for _, testCase := range []*parserTestSuite{
		{
			ttl: "Body mass index",
		},
		{
			ttl: "Toyota",
		},
		{
			ttl: "Gwen Stefani",
		},
		{
			ttl: "Kane (noble family)",
		},
		{
			ttl: "Free trade agreements of the European Union",
		},
		{
			ttl: "The Futurist",
		},
		{
			ttl: "Zillow",
		},
		{
			ttl: "Google Chrome",
		},
		{
			ttl: "_i_Monday Night Football__i_",
		},
		{
			ttl: "Grimes",
		},
		{
			ttl: "Thesaurus",
		},
		{
			ttl: "Aldi",
		},

		{
			ttl: "FetLife",
		},

		{
			ttl: "Slogan",
		},
		{
			ttl: "ABC News",
		},
		{
			ttl: "Eurovision Song Contest 2023",
		},
		{
			ttl: "O",
		},
		{
			ttl: "Ciro_Immobile",
		},
		{
			ttl: "Gad_Elmaleh",
		},
		{
			ttl: ".NET_Foundation",
		},
		{
			ttl: "(Dont_Go_Back_To)_Rockville",
		},
		{
			ttl: "1976_Democratic_Party_presidential_primaries",
		},
		{
			ttl: "Albert_Einstein",
		},
		{
			ttl: "Alexandros_Ioannis_Ginnis",
		},
		{
			ttl: "Anarchisme",
		},
		{
			ttl: "Angela_Merkel",
		},
		{
			ttl: "Armadni_general",
		},
		{
			ttl: "Azerbaijan",
		},
		{
			ttl: "Benjamin_Franklin",
		},
		{
			ttl: "Berliner_Mauer",
		},
		{
			ttl: "Brasilia",
		},
		{
			ttl: "Cactus",
		},
		{
			ttl: "Caryophyllales",
		},
		{
			ttl: "Charles_Darwin",
		},
		{
			ttl: "China",
		},
		{
			ttl: "Coset",
		},
		{
			ttl: "Cot_side",
		},
		{
			ttl: "Darth_Vader",
		},
		{
			ttl: "David_Cameron",
		},
		{
			ttl: "Distinctive_feature",
		},
		{
			ttl: "(Dont_Go_Back_To)_Rockville",
		},
		{
			ttl: "Financial_statement",
		},
		{
			ttl: "Foobar",
		},
		{
			ttl: "Ganges",
		},
		{
			ttl: "Grecia",
		},
		{
			ttl: "International_Music_Score_Library_Project",
		},
		{
			ttl: "ISO_IEC_2022",
		},
		{
			ttl: "ISO_IEC_8895_2",
		},
		{
			ttl: "January-February_2019_North_American_cold_wave",
		},
		{
			ttl: "Jose_Eduardo_Dos_Santos",
		},
		{
			ttl: "Josephine_Baker",
		},
		{
			ttl: "Kaori_Icho",
		},
		{
			ttl: "Kyoto",
		},
		{
			ttl: "Lionel_Messi",
		},
		{
			ttl: "London",
		},
		{
			ttl: "Main_page",
		},
		{
			ttl: "Malta",
		},
		{
			ttl: "Matrjoschka",
		},
		{
			ttl: "Matryoshka_doll",
		},
		{
			ttl: "Mike_Pence",
		},
		{
			ttl: "Mike_Pompeo",
		},
		{
			ttl: "Mino_Raiola",
		},
		{
			ttl: "Moores_law",
		},
		{
			ttl: "Mumbai",
		},
		{
			ttl: "Narendra_Modi",
		},
		{
			ttl: "Nephrite",
		},
		{
			ttl: "Organic_acid_anhydride",
		},
		{
			ttl: "Ottoman_Empire",
		},
		{
			ttl: "Otto_Warmbier",
		},
		{
			ttl: "Roger_Federer",
		},
		{
			ttl: "Rougequeue_de_Hodgson",
		},
		{
			ttl: "Sablon:Infokutija_filozof",
		},
		{
			ttl: "Scythian_languages",
		},
		{
			ttl: "Shakira",
		},
		{
			ttl: "S.L._Benfica",
		},
		{
			ttl: "Spain",
		},
		{
			ttl: "Switzerland",
		},
		{
			ttl: "United_States",
		},
		{
			ttl: "Vereinigte_Stahlwerke",
		},
		{
			ttl: "Vitis",
		},
		{
			ttl: "William_Shakespeare",
		},
		{
			ttl: "Yoshua_Bengio",
		},
		{
			ttl: "Btopa",
		},
		{
			ttl: "Knetka",
		},
		{
			ttl: "FR_Josephine_Baker",
		},
		{
			ttl: "Francois_Lapous",
		},
		{
			ttl: "Poetry",
		},
	} {
		suite.Run(t, testCase)
	}
}
