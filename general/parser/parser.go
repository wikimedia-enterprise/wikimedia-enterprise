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

// Parts type for the broken down page.
const (
	PartTypeInfoBox   = "infobox"
	PartTypeField     = "field"
	PartTypeImage     = "image"
	PartTypeSection   = "section"
	PartTypeList      = "list"
	PartTypeListItem  = "list_item"
	PartTypeParagraph = "paragraph"
)

var (
	// Regular expressions list for the package.
	spaces                         = regexp.MustCompile(`\s\s+`)
	spacesOrHyphen                 = regexp.MustCompile(`(\s{2,}-)|\s{2,}`) // Regex to match multiple spaces possibly following a hyphen and a space
	blankLines                     = regexp.MustCompile(`(?m)^\s*$\n?`)
	formula                        = regexp.MustCompile(`(\d|[A-Z][a-z]?|\))<su[bp]>[0-9+-]{1,3}<\/su[bp]>`)
	mathExpr                       = regexp.MustCompile(`\(( |&nbsp;)*<i>([A-z]|[0-9+-]){1}<\/i>`)
	brace                          = regexp.MustCompile(`\(|\)`)
	nestedBraces                   = regexp.MustCompile(`([(（][^()（）]*)([(（][^()（）]+[)）])`)
	bracesWithSpace                = regexp.MustCompile(`( |&nbsp;)\([^)]+ [^)]*\)`)
	bracesWithSpaceColonComma      = regexp.MustCompile(`（[^）]+[ ：，][^）]+）`)
	emptyBraces                    = regexp.MustCompile(`\(\s*\)`)
	emptyBracket                   = regexp.MustCompile(`\[\s*\]`)
	spaceBeforePunctuationNonLatin = regexp.MustCompile(`(\s|&nbsp;)([，。])`)
	spaceAroundPunctuationLatin    = regexp.MustCompile(`(\s|&nbsp;)([,.!?])(\s|&nbsp;|<\/)`)
	spaceBeforePunctuation         = regexp.MustCompile(`(\s|&nbsp;)([,.!?])`)
	spacesAroundParenthesis        = regexp.MustCompile(`(\s*\()\s+|\s+(\))`)
	allInOneEmpty                  = regexp.MustCompile(`\p{C}|\[[^\]]+\]|•`)
	allInOneSpace                  = regexp.MustCompile(`\s+|&#160;|&nbsp;|\p{Zs}| `)
	allInOneDelimiter              = regexp.MustCompile(`\s+([:;：；, ，、])`)

	parenthesisSemicolon = regexp.MustCompile(`\(\s*;\s*`)
	isRef                = regexp.MustCompile(`\[[0-9]+\]`)
	editText             = regexp.MustCompile(`\(\s*edit\s*\)`)
	numeric              = regexp.MustCompile(`^\d+$`)
	delimiterAtEnd       = regexp.MustCompile(`(?:[:;；,，、]+)$`)
	multiLine            = regexp.MustCompile(`\n+`)

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
		`.lang-comment`, `.rt-commentedText`, `[data-mw*="respell"]`, `[title*="Help:Pronunciation"]`,
	}

	infoboxSelector = `.infobox, .infobox_v3, .infobox_v2, .sinottico`

	rmSelector = `
	figure, footer, input, link, nav, noscript, script, style, sub, sup, table,
	.book, .catlinks, .citation, .gallery, .gallerybox, .hatnote, .listaref, .metadata,
	.mw-authority-control, .mw-editsection-bracket, .mw-editsection-divider,
	.mw-editsection-like, .mw-editsection-visualeditor, .mw-editsection,
	.mw-footer-container, .mw-gallery-packed, .mw-magiclink-isbn, .mw-mf-linked-projects, .mw-redirectedfrom,
	.NavFrame, .navigation-not-searchable, .noprint,
	.normdaten, .portal-bar, .printfooter,
	.refbegin-columns, .refbegin, .references, .reflist,
	.side-box, .sister-box, .sistersitebox,
	.vector-body-before-content, .vector-dropdown, .vector-header-container, .vector-page-toolbar, .vector-settings,
	*[role="navigation"]
	`

	rmHeaderParent = `#References, #Explanatory_notes, #Further_reading, #See_also, #External_links, #Notes, #Notable_people
	#Primary_sources, #Secondary_sources, #Tertiary_sources, #Citations, #General_and_cited_sources, #Bibliography,
	#Referencias, #Véase_también, #Bibliografía, #Enlaces_externos, #Ciudades_hermanas, #Referencias_y_notas,
	#Literatur, #Weblinks, #Einzelnachweise,
	#Références, #Voir_aussi, #Bibliographie, #Liens_externes, #Notes_et_références, #Notes, #Références,
	#Galerie_photos, #Liens_externes, #Notes_et_références, #Notes, #Références, #Bibliographie, #Voir_aussi,
	#Referencias, #Véase_también, #Bibliografía, #Enlaces_externos, #Ciudades_hermanas, #Referencias_y_notas,
	#Referências, #Ver_também, #Bibliografia, #Ligações_externas, #Notas, #Referências,
	#Annexes, #Bibliographie, #Liens_externes, #Notes_et_références, #Notes, #Références, #Article_connexe
	#Bibliografia, #Voci_correlate, #Altri_progetti, #Collegamenti_esterni, #Note_e_riferimenti, #Note, #Riferimenti`

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

// LinksGetter interface to expose method to get all the links in a page.
type LinksGetter interface {
	GetLinks(sel *goquery.Selection, bul string) []*Link
}

// InfoBoxesGetter interface to expose method to get all the infoboxes in a page.
type InfoBoxesGetter interface {
	GetInfoBoxes(sel *goquery.Selection) []*Part
}

// SectionsGetter interface to expose method to get the sections of a page.
type SectionsGetter interface {
	GetSections(sel *goquery.Selection) []*Part
}

// ParagraphsGetter interface to expose method to get the paragraphs of a part of the page.
type ParagraphsGetter interface {
	GetParagraphs(sel *goquery.Selection) []*Part
}

// ListItemsGetter interface to expose method to get the list items of a part of the page.
type ListItemsGetter interface {
	GetListItems(sel *goquery.Selection) []*Part
}

// ReferencesGetter interface to expose method to get the references of a page.
type ReferencesGetter interface {
	GetReferences(s *goquery.Selection) []*Reference
}

// ListsReplacer interface to expose method to replaces the HTML lists of a page with plain text lists.
type ListsReplacer interface {
	ReplaceLists(s *goquery.Selection)
}

// API interface for the whole parser.
type API interface {
	TemplatesGetter
	CategoriesGetter
	AbstractGetter
	ImagesGetter
	TextGetter
	LinksGetter
	InfoBoxesGetter
	SectionsGetter
	ParagraphsGetter
	ListItemsGetter
	ReferencesGetter
	ListsReplacer
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

// Part represents a part of the parsed page.
type Part struct {
	Name     string   `json:"name,omitempty"`
	Type     string   `json:"type,omitempty"`
	Value    string   `json:"value,omitempty"`
	Values   []string `json:"values,omitempty"`
	HasParts []*Part  `json:"has_parts,omitempty"`
	Images   []*Image `json:"images,omitempty"`
	Links    []*Link  `json:"links,omitempty"`
}

// Image represents an image that can be found on a Wikipedia page.
type Image struct {
	ContentUrl      string `json:"content_url,omitempty"`
	AlternativeText string `json:"alternative_text,omitempty"`
	Caption         string `json:"caption,omitempty"`
	Height          int    `json:"height,omitempty"`
	Width           int    `json:"width,omitempty"`
}

// Link represents a link that can be found on a Wikipedia page.
type Link struct {
	URL    string   `json:"url,omitempty"`
	Text   string   `json:"text,omitempty"`
	Images []*Image `json:"images,omitempty"`
}

// Reference json representation of reference HTML structure.
type Reference struct {
	ID    string  `json:"id,omitempty"`
	Index int     `json:"index,omitempty"`
	Text  string  `json:"text,omitempty"`
	Links []*Link `json:"links,omitempty"`
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

	// Get the ipa text for a Nihongo template
	nih := slc.Find(`[data-mw*="Nihongo"]`).First().Text()

	// Remove specific nodes including subtitles which are templates
	for _, sl := range rmNodes {
		slc.Find(sl).Each(func(_ int, sc *goquery.Selection) {
			for _, n := range sc.Nodes {
				n.Parent.RemoveChild(n)
			}
		})
	}

	slc = slc.Find("body section").First() //.ChildrenFiltered("p,ul,ol,blockquote")

	// Remove generic unnecessary nodes
	slc.Find(`table,div,input,script,style,link,figure,figcaption,caption`).Each(func(_ int, sc *goquery.Selection) {
		for _, n := range sc.Nodes {
			n.Parent.RemoveChild(n)
		}
	})

	//convert superscript numbers to unicode
	p.convertSuperscriptNumberToUnicode(slc)

	// Convert OL, UL, DL lists into simple text, with bullet points with tabs and newlines
	p.ReplaceLists(slc)

	// Flatten the DOM - we replace a/span tags with plain text
	// in order to avoid noise later when applying regex to html.

	slc.Find(strings.Join(fltElements, ",")).Each(func(_ int, sc *goquery.Selection) {
		sc.ReplaceWithHtml(sc.Text())
	})

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
	hml = spaceAroundPunctuationLatin.ReplaceAllStringFunc(hml, func(mch string) string {
		return spaceAroundPunctuationLatin.FindStringSubmatch(mch)[2] + spaceAroundPunctuationLatin.FindStringSubmatch(mch)[3]
	})

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hml))
	if err != nil {
		return "", err
	}

	// Get the abstract text,
	abs := doc.Selection.Text()

	// if abstract is does not have the pronunciation words at the start then add it
	if !strings.HasPrefix(abs, nih) {
		abs = fmt.Sprintf("%s %s", nih, abs)
	}

	// Remove multi whitespaces except if it has hyphen at end, which are used for markdown bullet lists
	abs = spacesOrHyphen.ReplaceAllStringFunc(abs, func(m string) string {
		if strings.HasSuffix(m, "-") {
			return m
		}

		return " "
	})

	// Clean up spaces that appear beside punctuations, parentheses, and blank lines
	abs = spaceBeforePunctuation.ReplaceAllString(abs, "$2")
	abs = parenthesisSemicolon.ReplaceAllString(abs, "(")
	abs = blankLines.ReplaceAllString(abs, "")
	abs = strings.TrimSpace(abs)

	return abs, nil
}

// convertSuperscriptNumberToUnicode converts superscript numbers to unicode.
func (p *Parser) convertSuperscriptNumberToUnicode(sel *goquery.Selection) {
	sel.Find("sup").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if numeric.MatchString(text) {
			var replacement string
			for _, r := range text {
				replacement += string(rune(0x2070 + r - '0'))
			}
			s.ReplaceWithHtml(replacement)
		}
	})
}

// cleanHTML removes specific nodes from the html.
func (p *Parser) cleanHTML(sel *goquery.Selection, attrLookup [][]string, rmNodes []string) *goquery.Selection {
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
	p.RemoveWordBreakOpportunity(slc)

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

			_, _ = buf.WriteString(nde.Data + " ")
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
	str = spaceBeforePunctuation.ReplaceAllString(str, "$2")
	return strings.TrimSpace(str)
}

// GetLinks parses list of links based on provided element collection.
func (p *Parser) GetLinks(sel *goquery.Selection, bul string) []*Link {
	if len(bul) == 0 {
		bul = p.getBaseURL(sel)
	}

	lks := []*Link{}

	sel.Find("a").Each(func(_ int, ssl *goquery.Selection) {
		href := ssl.AttrOr("href", "")

		if strings.HasPrefix(href, "./File:") {
			return
		}

		lnk := new(Link)
		lnk.Text = p.GetText(ssl)
		lnk.Images = p.GetImages(ssl)

		if len(lnk.Text) == 0 {
			lnk.Text = ssl.AttrOr("title", "")
		}

		switch {
		case strings.HasPrefix(href, "./"):
			lnk.URL = fmt.Sprintf("%s%s", bul, strings.TrimPrefix(href, "./"))
		case strings.HasPrefix(href, "//"):
			lnk.URL = fmt.Sprintf("https:%s", href)
		case !strings.HasPrefix(href, "http"):
			lnk.URL = fmt.Sprintf("%s%s", bul, href)
		default:
			lnk.URL = href
		}

		lks = append(lks, lnk)
	})

	return lks
}

func (p *Parser) getInfoBoxPart(sel *goquery.Selection, base string) *Part {
	var b bytes.Buffer
	var lst string

	gtx := func(sel *goquery.Selection) string {
		str := p.GetText(sel, "tr")

		// remove any trailing delimiters from list item
		str = delimiterAtEnd.ReplaceAllString(str, "")
		return str
	}

	sel.Find(".infobox-data ul, .infobox-data ol, .infobox-data dl").Each(func(_ int, sli *goquery.Selection) {
		b.Reset()
		p.getList(&b, sli.Nodes[0], 1)
		lst = multiLine.ReplaceAllString(b.String(), "\n")

		// Remove the list from the infobox, so the `Value` field doesn't contain the list
		sli.ReplaceWithHtml("")
	})

	lbl := gtx(sel.Find(".infobox-label").First())
	if len(lbl) == 0 {
		lbl = gtx(sel.Find(".infobox-label tr"))
	}

	dta := ""
	if len(lst) == 0 {
		dta = gtx(sel.Find(".infobox-data").First())

		if len(dta) == 0 {
			dta = gtx(sel.Find(".infobox-data tr"))
		}
	}

	if len(dta) == 0 && len(lst) == 0 {
		dta = gtx(sel.Find(".infobox-full-data"))

		if len(dta) == 0 {
			dta = gtx(sel.Find(".infobox-full-data tr"))
		}
	}

	if len(dta) == 0 && len(lst) == 0 {
		ths := sel.ChildrenFiltered("th")
		tds := sel.ChildrenFiltered("td")

		if tds.Length() == 2 && ths.Length() == 0 {
			lbl = gtx(tds.First())
			dta = gtx(tds.Last())
		} else if tds.Length() == 1 && ths.Length() == 1 {
			lbl = gtx(ths.First())
			dta = gtx(tds.First())
		}

		if len(dta) == 0 && len(lst) == 0 {
			dta = gtx(sel)
		}
	}

	// will need to add a specific field types for this
	ims := p.GetImages(sel)

	if len(lbl) == 0 && len(lst) == 0 && len(dta) == 0 && len(ims) == 0 {
		return nil
	}

	typ := PartTypeField

	if len(dta) == 0 && len(ims) > 0 || sel.Find(".infobox-image").Length() > 0 {
		typ = PartTypeImage
	}

	if len(lst) > 0 {
		typ = PartTypeList
		// Get the text within .infobox-data (excluding UL, OL, and DL)
		sel.Find(".infobox-data").Each(func(i int, s *goquery.Selection) {
			dta = gtx(s.Contents().Not("ul, ol, dl"))
		})
	}

	return &Part{
		Type:   typ,
		Name:   lbl,
		Value:  dta,
		Images: ims,
		Links:  p.GetLinks(sel, base),
		Values: strings.Split(lst, "\n"),
	}
}

func (p *Parser) getInfoBox(sel *goquery.Selection) *Part {
	ibx := new(Part)
	ibx.Type = PartTypeInfoBox

	// get a template name
	if abt := sel.AttrOr("about", ""); len(abt) > 0 {
		sel.ParentsFiltered("body").Find(fmt.Sprintf(`[about="%s"]`, abt)).Each(func(_ int, ssl *goquery.Selection) {
			if dta := ssl.AttrOr("data-mw", ""); len(dta) > 0 && strings.Contains(dta, "parts") {
				ttl := gjson.Parse(dta).Get("parts.#.template.target.wt").Array()

				if len(ttl) > 0 {
					ibx.Name = strings.TrimSpace(ttl[0].String())
				}
			}
		})
	}

	var scn *Part
	bse := p.getBaseURL(sel)
	slc := p.cleanHTML(sel, cleanHTMlLookup, rmNodesText)
	itl := p.GetText(slc.Find(".infobox-title"))

	// specific to French wikipedia and infobox_v3 template
	// we need to try to get the title from the entete div
	if sel.HasClass("infobox_v3") && len(itl) == 0 {
		itl = p.GetText(slc.Find("div.entete"))
	}

	slc.Find("tr").Each(func(i int, s *goquery.Selection) {
		scs := []*goquery.Selection{
			s.Find(".infobox-above"),
			s.Find(".infobox-below"),
			s.Find(".infobox-header"),
			s.Find(".infobox-subheader"),
			s.Find(".sinottico_divisione"),
			s.Find(".sinottico_testata"),
			s.Find(".sinottico_piede"),
			s.Find(".sinottico_piede2"),
		}

		for _, s := range scs {
			if nme := p.GetText(s); len(nme) > 0 {
				scn = &Part{
					Name: nme,
					Type: PartTypeSection,
				}
				ibx.HasParts = append(ibx.HasParts, scn)
				return
			}
		}

		if scn == nil {
			scn = &Part{
				Name: itl,
				Type: PartTypeSection,
			}

			// Needed for french wikipedia infoboxes with main image not in a table row
			if sel.HasClass("infobox_v3") {
				scn.Images = p.GetImages(sel.Find(".images"))
				if len(scn.Images) > 0 {
					scn.Images[0].Caption = p.GetText(sel.Find(".legend").First())
				}
			}

			ibx.HasParts = append(ibx.HasParts, scn)
		}

		prt := p.getInfoBoxPart(s, bse)

		if prt == nil {
			return
		}

		scn.HasParts = append(scn.HasParts, prt)
	})

	return ibx
}

// GetInfoBoxes function extracts all the infoboxes from the HTML.
func (p *Parser) GetInfoBoxes(sel *goquery.Selection) []*Part {
	ibs := make([]*Part, 0)
	slc := sel.Clone()
	p.RemoveWordBreakOpportunity(slc)

	slc.Find(infoboxSelector).Each(func(_ int, scn *goquery.Selection) {
		ibx := p.getInfoBox(scn)
		ibs = append(ibs, ibx)
	})

	return ibs
}

// getSections parses list of sections based provided element collection.
func (p *Parser) getSections(sel *goquery.Selection, psc *Part) []*Part {
	scs := []*Part{}

	sel.Children().Each(func(_ int, scn *goquery.Selection) {
		switch tag := goquery.NodeName(scn); tag {
		case "section":
			csn := &Part{
				Type: PartTypeSection,
				Name: isRef.ReplaceAllString(
					p.GetText(scn.
						Find("h2,h3,h4,h5,h6,h7").First()), " "),
			}

			if len(csn.Name) == 0 && scn.AttrOr("data-mw-section-id", "") == "0" {
				csn.Name = "Abstract"
			}

			csn.HasParts = append(csn.HasParts, p.getSections(scn, csn)...)
			scs = append(scs, csn)
		case "p", "ul", "ol":
			// we're checking if the parent sections exists
			// technically this should never happen
			// but we are adding it just in case, to avoid panics
			if psc != nil {
				psc.HasParts = append(psc.HasParts, p.GetParagraphs(scn)...)
			}
		default:
			scs = append(scs, p.getSections(scn, psc)...)
		}
	})

	return scs
}

// GetSections parses list of sections based provided element collection.
func (p *Parser) GetSections(sel *goquery.Selection) []*Part {
	slc := sel.Clone()

	// remove infoboxes
	slc.Find(infoboxSelector).Each(func(_ int, scn *goquery.Selection) {
		scn.Remove()
	})

	// remove references and external links
	slc.Find(`section > div.reflist, section > *[id="External_links"]`).Each(func(_ int, scn *goquery.Selection) {
		scn.Parent().Remove() // remove the whole `section`
	})

	// remove generic unnecessary nodes
	slc.Find(rmSelector).Each(func(_ int, scn *goquery.Selection) {
		for _, nde := range scn.Nodes {
			nde.Parent.RemoveChild(nde)
		}
	})

	// Remove parent elements by IDs (e.g. References, See also, External links, etc.)
	slc.Find(rmHeaderParent).Each(func(i int, s *goquery.Selection) {
		s.Parent().Remove()
	})

	// Remove headers that have no sibling content
	slc.Find("section").Each(func(i int, s *goquery.Selection) {
		// Check if the section has only one child
		if s.Children().Length() == 1 {
			// Check if the first child is an <h1-h6> tag
			firstChild := s.Children().First()
			if firstChild.Is("h1") || firstChild.Is("h2") || firstChild.Is("h3") || firstChild.Is("h4") || firstChild.Is("h5") || firstChild.Is("h6") {
				s.Remove()
			}
		}
	})

	return p.getSections(slc.Find("body"), nil)
}

// removeInlineCitation removes citation numbers (example `[12]`) from the text while preserving the appropriate spaces, and removes extea space before any fullstop
func (p *Parser) removeInlineCitation(sel *goquery.Selection) string {
	txt := editText.ReplaceAllString(isRef.ReplaceAllString(p.GetText(sel), " "), " ")
	return strings.TrimSpace(strings.ReplaceAll(txt, " .", "."))
}

// GetParagraphs parses list of paragraphs based on provided element collection.
func (p *Parser) GetParagraphs(sel *goquery.Selection) []*Part {
	pgs := []*Part{}
	bul := p.getBaseURL(sel)

	sel.Each(func(_ int, scn *goquery.Selection) {
		pgf := new(Part)

		switch tag := goquery.NodeName(scn); tag {
		case "p":
			pgf.Type = PartTypeParagraph
			pgf.Value = p.removeInlineCitation(scn)
			pgf.Links = p.GetLinks(scn, bul)
		case "ol", "ul":
			pgf.Type = PartTypeList
			pgf.HasParts = p.GetListItems(scn)
		}

		if len(pgf.Value) != 0 || len(pgf.Links) != 0 || len(pgf.HasParts) != 0 {
			pgs = append(pgs, pgf)
		}
	})

	return pgs
}

// GetListItems parses list of items based on provided element collection.
func (p *Parser) GetListItems(sel *goquery.Selection) []*Part {
	lis := []*Part{}
	bul := p.getBaseURL(sel)

	sel.Find("li").Each(func(_ int, scn *goquery.Selection) {
		prt := new(Part)
		prt.Type = PartTypeListItem
		prt.Value = p.removeInlineCitation(scn)
		prt.Links = p.GetLinks(scn, bul)

		if len(prt.Value) != 0 || len(prt.Links) != 0 {
			lis = append(lis, prt)
		}
	})

	return lis
}

// GetReferences parses list of references based on provided element collection.
func (p *Parser) GetReferences(sel *goquery.Selection) []*Reference {
	bul := p.getBaseURL(sel)
	rfs := []*Reference{}
	irf := func(i int, sel *goquery.Selection) bool {
		return isRef.MatchString(sel.Text())
	}

	sel.Find(".references a").
		FilterFunction(irf).
		ChildrenFiltered(".references li").Each(func(idx int, scn *goquery.Selection) {
		ref := new(Reference)
		ref.ID = scn.AttrOr("id", "")
		ref.Index = idx
		ref.Text = strings.TrimPrefix(p.GetText(scn), "↑ ")
		ref.Links = p.GetLinks(scn, bul)

		rfs = append(rfs, ref)
	})

	return rfs
}

// findFirstLevel finds the nodes matching the css, stops at the first level of the tree where the css is found
func (p *Parser) findFirstLevel(sel *goquery.Selection, css string) *goquery.Selection {
	// Find the nodes matching the css
	s := sel.Find(css)

	// Remove any nested items so we only get the first level of matching nodes
	cls := strings.Split(strings.ReplaceAll(css, " ", ""), ",")
	for _, a := range cls {
		for _, b := range cls {
			avoid := fmt.Sprintf("%s %s", a, b)
			s = s.Not(avoid)
		}
	}

	return s
}

// getList parses list of lists based on provided element collection.
func (p *Parser) getList(b *bytes.Buffer, n *html.Node, depth int) {
	if n.Type == html.ElementNode && (n.Data == "li" || n.Data == "dt" || n.Data == "dd") {
		b.WriteString(strings.Repeat("  ", depth) + "- ")
	}

	if n.Type == html.TextNode {
		b.WriteString(n.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol" || c.Data == "dl") {
			b.WriteString("\n")
			p.getList(b, c, depth+1)
		} else {
			p.getList(b, c, depth)
		}
	}

	if n.Type == html.ElementNode && (n.Data == "li" || n.Data == "dt" || n.Data == "dd") {
		b.WriteString("\n")
	}
}

// ReplaceLists replaces the HTML lists with plain text lists.
func (p *Parser) ReplaceLists(sel *goquery.Selection) {
	var b bytes.Buffer

	p.findFirstLevel(sel, "ul, ol, dl").Each(func(_ int, sli *goquery.Selection) {
		// avoid empty lists and protect against panic for Nodes[0]
		if len(sli.Nodes) == 0 {
			return
		}

		b.Reset()
		//start with a blank line
		b.WriteString("\n")

		//Get the list as Markdown, and replace the HTML with the Markdown
		p.getList(&b, sli.Nodes[0], 1)
		md := b.String()
		sli.ReplaceWithHtml(md)
	})
}

// RemoveWordBreakOpportunity remove WBR tags (optional line break) and concatenate the the adjacent text nodes into one text node
func (p *Parser) RemoveWordBreakOpportunity(sel *goquery.Selection) {
	sel.Find("wbr").Each(func(_ int, wbr *goquery.Selection) {
		if len(wbr.Nodes) == 0 {
			wbr.Remove()
			return
		}

		prv := wbr.Nodes[0].PrevSibling
		nxt := wbr.Nodes[0].NextSibling
		if prv == nil || nxt == nil {
			wbr.Remove()
			return
		}

		if prv.Type == html.TextNode && nxt.Type == html.TextNode {
			prv.Data += nxt.Data
			nxt.Parent.RemoveChild(nxt)
		}

		wbr.Remove()
	})
}
