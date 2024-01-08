// Package parser implements utility routines for traversing and retrieving elements from a page html.
//
// The parser packages adds a series of utilities and methods to retrieve elements from wikipedia html page api.
// ex: https://en.wikipedia.org/api/rest_v1/page/html/Josephine_Baker
package parser

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

var (
	// Regular expressions list for the package.
	spaces                         = regexp.MustCompile(`\s\s+`)
	formula                        = regexp.MustCompile(`(\d|[A-Z][a-z]?|\))<su[bp]>[0-9+-]{1,3}<\/su[bp]>`)
	mathExpr                       = regexp.MustCompile(`\(( |&nbsp;)*<i>([A-z]|[0-9+-]){1}<\/i>`)
	brace                          = regexp.MustCompile(`\(|\)`)
	nestedBraces                   = regexp.MustCompile(`([(（][^()（）]*)([(（][^()（）]+[)）])`)
	bracesWithSpace                = regexp.MustCompile(`( |&nbsp;)\([^)]+ [^)]*\)`)
	bracesWithSpaceColonComma      = regexp.MustCompile(`（[^）]+[ ：，][^）]+）`)
	emptyBraces                    = regexp.MustCompile(`\(\s*\)`)
	emptyBracket                   = regexp.MustCompile(`\[\s*\]`)
	spaceBeforePunctuationNonLatin = regexp.MustCompile(`(\s|&nbsp;)([，。])`)
	spaceBeforePunctuationLatin    = regexp.MustCompile(`(\s|&nbsp;)([,.!?])(\s|&nbsp;|<\/)`)
	spacesAroundParenthesis        = regexp.MustCompile(`(\s*\()\s+|\s+(\))`)
	allInOneEmpty                  = regexp.MustCompile(`\p{C}|\[[^\]]+\]|•`)
	allInOneSpace                  = regexp.MustCompile(`\s+|&#160;|&nbsp;|\p{Zs}| `)
	allInOneDelimiter              = regexp.MustCompile(`\s+([:;：；, ，、])`)

	// Specific nodes to remove for abstract
	rmNodes = []string{
		// References
		`[class*="mw-ref"]`, `.mw-ref`, `.mw-reflink-text`, `.reference`,
		// Pronunciations
		`span[data-mw*="target":{"wt":"IPA"]`, `figure-inline[data-mw*="target":{"wt":"IPA"]`, `[title*="pronunciation"]`,
		`[data-mw*="IPA"]`, `.IPA`, `.pronunciation`, `.respelling`,
		// Display styles
		`span[style="display:none"]`, `span.Z3988`, `#coordinates`, `table.navbox`,
		`.geo-nondefault`, `.geo-multi-punct`, `.hide-when-compact`, `div.infobox`, `div.magnify`, `.noprint`,
		`.noexcerpt`, `.nomobile`, `span[class*=error]`, `.sortkey`, `ul.gallery`,
		`[encoding="application/x-tex"]`,
		// Comments
		`.lang-comment`,
	}

	// Element tags to flatten for abstract
	fltElements = []string{"a", "span"}

	// Attributes to remove for abstract
	rmAttrs = []string{"about", "data-mw", "id", "typeof"}

	// Remove nodes for text extraction
	rmNodesText = []string{"script", "style", "#comment"}

	// HTML cleaning lookup for images and infobox parsing
	cleanHTMlLookup = [][]string{
		{"class", "noprint"},
		{"class", "noviewer"},
		{"class", "navigation-only"},
		{"style", "display:none"},
	}
)

// TemplatesGetter interface to expose method to get all templates in a page.
type TemplatesGetter interface {
	GetTemplates(sel *goquery.Selection) Templates
}

// CategoriesGetter interface to expose method to get all categories in a page.
type CategoriesGetter interface {
	GetCategories(sel *goquery.Selection) Categories
}

// AbstractGetter interface to expose method to get abstract of a page.
type AbstractGetter interface {
	GetAbstract(sel *goquery.Selection) (string, error)
}

// ImagesGetter interface to expose method to get images of a page.
type ImagesGetter interface {
	GetImages(sel *goquery.Selection) []*Image
}

// TextGetter interface to expose method to get text of a page.
type TextGetter interface {
	GetText(sel *goquery.Selection, ign ...string) string
}

// API interface for the whole parser.
type API interface {
	TemplatesGetter
	CategoriesGetter
	AbstractGetter
	ImagesGetter
	TextGetter
}

// Relation base structure for the templates and categories.
type Relation struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Relations is a slice of Relation pointers.
type Relations []*Relation

// GetNames method to get all the names of the templates or categories.
func (t Relations) GetNames() []string {
	nms := []string{}

	for _, tpl := range t {
		nms = append(nms, tpl.Name)
	}

	return nms
}

// HasSubstringInNames method returns true if any of the provided substrings are present in any of the relation (template/category) names.
func (t Relations) HasSubstringInNames(sbs []string) bool {
	for _, nme := range t.GetNames() {
		for _, sub := range sbs {
			if strings.Contains(strings.ToLower(nme), strings.ToLower(sub)) {
				return true
			}
		}
	}

	return false
}

// Template structure for the GetTemplates.
type Template = Relation

// Category structure for the GetCategories.
type Category = Relation

// Templates is a slice of Template pointers.
type Templates = Relations

// Categories is a slice of Category pointers.
type Categories = Relations

// Image represents an image that can be found on a Wikipedia page.
type Image struct {
	ContentUrl      string `json:"content_url,omitempty"`
	AlternativeText string `json:"alternative_text,omitempty"`
	Caption         string `json:"caption,omitempty"`
	Height          int    `json:"height,omitempty"`
	Width           int    `json:"width,omitempty"`
}

// New creates new instance of the parser.
func New() API {
	return &Parser{}
}

// Parser structure for defining the type.
type Parser struct{}

// getPageTitle method that returns the page title.
func (p *Parser) getTitle(lnk string) string {
	ttl := strings.ReplaceAll(strings.TrimPrefix(lnk, "./"), "_", " ")

	if etl, err := url.QueryUnescape(ttl); err == nil {
		return etl
	}

	return ttl
}

// getFullURL method that returns the page full url.
func (p *Parser) getFullURL(lnk string, pul string) string {
	ttl := fmt.Sprintf("%s%s", pul, strings.TrimPrefix(lnk, "./"))

	if ttl, err := url.QueryUnescape(ttl); err != nil {
		return ttl
	}

	return ttl
}

func (p *Parser) getBaseURL(sel *goquery.Selection) string {
	gul := func(bul string) string {
		return fmt.Sprintf("https:%s", bul)
	}

	if url := sel.ParentsFiltered("html").Find("base").AttrOr("href", ""); len(url) > 0 {
		return gul(url)
	}

	if url := sel.Find("base").AttrOr("href", ""); len(url) > 0 {
		return gul(url)
	}

	return ""
}

// GetTemplates method to get all templates in a page.
func (p *Parser) GetTemplates(sel *goquery.Selection) Templates {
	// As services pass the same sel *goquery.Selection to all parser methods
	// we need to clone selection for each method as the selection csm be
	// modified by any of parser methods
	slc := sel.Clone()
	rsp := map[string]*Template{}
	bul := p.getBaseURL(slc)

	slc.Find("*[data-mw]").Each(func(_ int, s *goquery.Selection) {
		atr, ok := s.Attr("data-mw")

		if ok && strings.Contains(atr, "parts") {
			res := gjson.Get(atr, "parts.#.template.target.href")

			for _, res := range res.Array() {
				lnk := res.String()
				ttl := p.getTitle(lnk)

				rsp[ttl] = &Template{
					Name: ttl,
					URL:  p.getFullURL(lnk, bul),
				}
			}
		}
	})

	tps := []*Template{}

	for _, tpl := range rsp {
		tps = append(tps, tpl)
	}

	return tps
}

// GetCategories method to get all categories in a page.
func (p *Parser) GetCategories(sel *goquery.Selection) Categories {
	// As services pass the same sel *goquery.Selection to all parser methods
	// we need to clone selection for each method as the selection csm be
	// modified by any of parser methods
	slc := sel.Clone()
	rsp := map[string]*Category{}
	url := p.getBaseURL(slc)

	slc.Find(`link[rel="mw:PageProp/Category"]`).Each(func(_ int, s *goquery.Selection) {
		lnk, ok := s.Attr("href")

		if ok {
			// Split by # to remove redundant data which doesn't influence the link
			sln := strings.Split(lnk, "#")[0]
			ttl := p.getTitle(sln)

			if len(ttl) != 0 {
				rsp[ttl] = &Category{
					Name: ttl,
					URL:  p.getFullURL(sln, url),
				}
			}
		}
	})

	cts := []*Category{}

	for _, ctg := range rsp {
		cts = append(cts, ctg)
	}

	return cts
}

// escapeParentAttr escapes parenthesis with random strings that may occur in the attribute values of any node in the selector.
// This is needed to safely apply regex for parenthesis to the content of the node.
func (p *Parser) escapeParentAttr(sel *goquery.Selection) *goquery.Selection {
	sel.Find("*").Each(func(i int, sc *goquery.Selection) {
		for _, n := range sc.Nodes {
			for _, a := range n.Attr {
				if brace.MatchString(a.Val) {
					nVal := a.Val
					nVal = brace.ReplaceAllStringFunc(nVal, func(mch string) string {
						if mch == "(" {
							return `substitue_for_open_paren_in_attr`
						}
						return `substitue_for_closed_paren_in_attr`
					})

					sc.FilterNodes(n).SetAttr(a.Key, nVal)
				}
			}
		}
	})

	return sel
}

// removeNestedBraces removes nested braces recursively.
func (p *Parser) removeNestedBraces(html string) string {
	ret := nestedBraces.ReplaceAllStringFunc(html, func(mch string) string {
		return nestedBraces.FindStringSubmatch(mch)[1]
	})

	if nestedBraces.MatchString(ret) {
		return p.removeNestedBraces(ret)
	}

	return ret
}

// GetAbstract method to get abstract on an article from html.
func (p *Parser) GetAbstract(sel *goquery.Selection) (string, error) {
	// As services pass the same sel *goquery.Selection to all parser methods
	// we need to clone selection for each method as the selection csm be
	// modified by any of parser methods
	slc := sel.Clone()

	// Remove specific nodes including subtitles which are templates
	for _, sl := range rmNodes {
		slc.Find(sl).Each(func(_ int, sc *goquery.Selection) {
			for _, n := range sc.Nodes {
				n.Parent.RemoveChild(n)
			}
		})
	}

	slc = slc.Find("body section").First().ChildrenFiltered("p,ul,ol")

	// Remove generic unnecessary nodes
	slc.Find(`table,div,input,script,style,link`).Each(func(_ int, sc *goquery.Selection) {
		for _, n := range sc.Nodes {
			n.Parent.RemoveChild(n)
		}
	})

	// Flatten the DOM - we replace a/span tags with plain text
	// in order to avoid noise later when applying regex to html.
	for _, s := range fltElements {
		slc.Find(s).Each(func(_ int, sc *goquery.Selection) {
			sc.ReplaceWithHtml(sc.Text())
		})
	}

	// Remove certain attributes to make it easier to apply regex later.
	slc.Find("*").Each(func(_ int, sc *goquery.Selection) {
		for _, atr := range rmAttrs {
			sc.RemoveAttr(atr)
		}
	})

	// Escape parentheses in node attributes
	slc.Each(func(_ int, s *goquery.Selection) {
		p.escapeParentAttr(s)
	})

	// Retrieve html from the processed selection
	hml := ""

	slc.Each(func(_ int, s *goquery.Selection) {
		shl, _ := s.Html()
		hml = fmt.Sprintf("%s %s", hml, shl)
	})

	if !formula.MatchString(hml) && !mathExpr.MatchString(hml) {
		// Remove any nested parentheticals
		hml = p.removeNestedBraces(hml)

		// [1] Replace any parentheticals which have at least one space inside
		// and at least one space or No-Break Space before.
		hml = bracesWithSpace.ReplaceAllString(hml, "")

		// Remove content inside any other non-Latin parentheticals. The behaviour is the
		// same as 1 but for languages that are not Latin based.
		// One difference to #1 is that in addition to a space the non-Latin colon or comma
		// could also trigger the removal of parentheticals.
		// Another difference is that a leading space is not required since in non-Latin
		// languages that rarely happens.
		hml = bracesWithSpaceColonComma.ReplaceAllString(hml, "")
	}

	// Remove any empty parentheticals due to earlier transformations, or pre-existing ones.
	hml = emptyBraces.ReplaceAllString(hml, "")
	hml = emptyBracket.ReplaceAllString(hml, "")

	// Replace any leading spaces or &nbsp; before punctuation for non-Latin
	// (which could be the result of earlier transformations)
	hml = spaceBeforePunctuationNonLatin.ReplaceAllStringFunc(hml, func(mch string) string {
		return spaceBeforePunctuationNonLatin.FindStringSubmatch(mch)[2]
	})

	// Replace any leading spaces or &nbsp; before punctuation for Latin
	// (which could be the result of earlier transformations), but only if whitespace
	// or a closing tag come afterwards.
	hml = spaceBeforePunctuationLatin.ReplaceAllStringFunc(hml, func(mch string) string {
		return spaceBeforePunctuationLatin.FindStringSubmatch(mch)[2] + spaceBeforePunctuationLatin.FindStringSubmatch(mch)[3]
	})

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hml))
	if err != nil {
		return "", err
	}

	abs := doc.Selection.Text()

	// Remove all multi-spaces
	abs = spaces.ReplaceAllString(abs, " ")

	return abs, nil
}

// cleanHTML removes specific nodes from the html.
func (*Parser) cleanHTML(sel *goquery.Selection, attrLookup [][]string, rmNodes []string) *goquery.Selection {
	slc := sel.Clone()
	lcf := func(slc *goquery.Selection) bool {
		for _, nde := range attrLookup {
			if strings.Contains(slc.AttrOr(nde[0], ""), nde[1]) {
				return true
			}
		}

		return false
	}

	rmf := func(_ int, rsl *goquery.Selection) bool {
		return slices.Contains(rmNodes, goquery.NodeName(rsl)) ||
			lcf(rsl)
	}

	slc.Find("*").FilterFunction(rmf).Remove()

	return slc
}

// GetImages method to get images from html element.
func (p *Parser) GetImages(sel *goquery.Selection) []*Image {
	ims := []*Image{}
	slc := p.cleanHTML(sel, cleanHTMlLookup, rmNodesText)

	slc.Find("img").Each(func(_ int, ssl *goquery.Selection) {
		img := new(Image)
		img.ContentUrl = fmt.Sprintf("https:%s", ssl.AttrOr("src", ""))
		img.AlternativeText = ssl.AttrOr("alt", "")
		img.Width, _ = strconv.Atoi(ssl.AttrOr("width", "0"))
		img.Height, _ = strconv.Atoi(ssl.AttrOr("width", "0"))
		img.Caption = ssl.ParentsFiltered(".infobox-image").Find(".infobox-caption").Text()

		if len(img.Caption) == 0 {
			img.Caption = p.GetText(ssl.ParentsFiltered("figure").Find("figcaption"))
		}

		ims = append(ims, img)
	})

	return ims
}

// GetText method to get text from html element, ign (ignore) is a list of HTML tags to skip when building the text string.
func (p *Parser) GetText(sel *goquery.Selection, ign ...string) string {
	slc := p.cleanHTML(sel, cleanHTMlLookup, rmNodesText)

	var fnc func(*html.Node)
	buf := new(bytes.Buffer)

	fnc = func(nde *html.Node) {
		if nde.Type == html.TextNode {
			nds := []*html.Node{}
			rec := nde

			// look for the first 3 parents, we need spaces between text to display correctly
			// this will add the spaces for siblings that are separated by a html elements
			for i := 0; i < 3; i++ {
				if rec == nil {
					break
				}

				nds = append(nds, rec)
				rec = rec.Parent
			}

			if nde.Parent != nil {
				nds = append(nds, nde.Parent.Parent)

				if nde.Parent.Parent != nil {
					nds = append(nds, nde.Parent.Parent.Parent)
				}
			}

			// if the previous sibling is an html element, add a space
			for _, tnd := range nds {
				if tnd != nil && tnd.PrevSibling != nil && tnd.PrevSibling.Type == html.ElementNode {
					if !strings.HasPrefix(tnd.Data, " ") {
						_, _ = buf.WriteString(" ")
					}
				}
			}

			_, _ = buf.WriteString(nde.Data)
		}

		if nde.FirstChild != nil {
			for sbg := nde.FirstChild; sbg != nil; sbg = sbg.NextSibling {
				// skip inner table rows
				if sbg.Type == html.ElementNode && slices.Contains(ign, sbg.Data) {
					continue
				}

				fnc(sbg)
			}
		}
	}

	for _, nde := range slc.Nodes {
		fnc(nde)
	}

	// clean up the text
	str := buf.String()
	str = allInOneEmpty.ReplaceAllString(str, "")
	str = allInOneDelimiter.ReplaceAllString(str, "$1")
	str = allInOneSpace.ReplaceAllString(str, " ")
	str = spaces.ReplaceAllString(str, " ")
	str = spacesAroundParenthesis.ReplaceAllString(str, "$1$2")

	return strings.TrimSpace(str)
}
