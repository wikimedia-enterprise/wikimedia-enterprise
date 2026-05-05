// Package parser implements utility routines for traversing and retrieving elements from a page html.
//
// The parser packages adds a series of utilities and methods to retrieve elements from wikipedia html page api.
// ex: https://en.wikipedia.org/api/rest_v1/page/html/Josephine_Baker
package parser

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	stdhtml "html"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"wikimedia-enterprise/general/parser/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
	nethtml "golang.org/x/net/html"
)

// Parts type for the broken down page.
const (
	PartTypeInfoBox        = "infobox"
	PartTypeField          = "field"
	PartTypeImage          = "image"
	PartTypeGallery        = "gallery"
	PartTypeSection        = "section"
	PartTypeList           = "list"
	PartTypeListItem       = "list_item"
	PartTypeParagraph      = "paragraph"
	PartTypeTable          = "table"
	PartTypeDefinitionList = "definition_list"
	PartTypeDefinitionTerm = "definition_term"
	PartTypeDefinition     = "definition"
	PartTypeOrderedList    = "ordered_list"
)

type ValMode int

// ValMode represents how a cell value is interpreted.
// Used to classify non-informative values.
const (
	Duplicate   ValMode = iota
	Empty       ValMode = iota
	Placeholder ValMode = iota
	Zero        ValMode = iota
)

type OrderedTable struct {
	ID    string
	Table *TableWithHeaders
}

var (
	// Regular expressions list for the package.
	spaces = regexp.MustCompile(`\s\s+`)
	//formula                        = regexp.MustCompile(`(\d|[A-Z][a-z]?|\))<su[bp]>[0-9+-]{1,3}<\/su[bp]>`)
	//mathExpr                       = regexp.MustCompile(`\(( |&nbsp;)*<i>([A-z]|[0-9+-]){1}<\/i>`)
	//brace                          = regexp.MustCompile(`\(|\)`)
	//nestedBraces                   = regexp.MustCompile(`([(（][^()（）]*)([(（][^()（）]+[)）])`)
	//bracesWithSpace                = regexp.MustCompile(`( |&nbsp;)\([^)]+ [^)]*\)`)
	//bracesWithSpaceColonComma      = regexp.MustCompile(`（[^）]+[ ：，][^）]+）`)
	//emptyBraces                    = regexp.MustCompile(`\(\s*\)`)
	//emptyBracket                   = regexp.MustCompile(`\[\s*\]`)
	//spaceBeforePunctuationNonLatin = regexp.MustCompile(`(\s|&nbsp;)([，。])`)
	//spaceBeforePunctuationLatin    = regexp.MustCompile(`(\s|&nbsp;)([,.!?])(\s|&nbsp;|<\/)`)
	spacesAroundParenthesis = regexp.MustCompile(`(\s*\()\s+|\s+(\))`)
	allInOneEmpty           = regexp.MustCompile(`\p{C}|\[[^\]]+\]|•`)
	allInOneSpace           = regexp.MustCompile(`\s+|&#160;|&nbsp;|\p{Zs}| `)
	allInOneDelimiter       = regexp.MustCompile(`\s+([:;：；, ，、])`)
	delimiterAtEnd          = regexp.MustCompile(`(?:[:;；,，、]+)$`)
	isRef                   = regexp.MustCompile(`\[[0-9]+\]`)
	editText                = regexp.MustCompile(`\(\s*edit\s*\)`)
	numeric                 = regexp.MustCompile(`^\d+$`)
	percentPattern          = regexp.MustCompile(`^\s*(\d+(\.\d+)?)\s*%\s*$`)
	infoboxSelector         = cascadia.MustCompile(`.infobox, .infobox_v3, .infobox_v2, .sinottico`)
	spacesAndTabs           = regexp.MustCompile(`[ \t]+`)

	rmSelector = cascadia.MustCompile(`
	footer, input, link, nav, noscript, script, style, sub,
	.book, .catlinks, .hatnote, .listaref, .metadata,
	.mw-authority-control, .mw-editsection-bracket, .mw-editsection-divider,
	.mw-editsection-like, .mw-editsection-visualeditor, .mw-editsection,
	.mw-footer-container, .mw-magiclink-isbn, .mw-mf-linked-projects, .mw-redirectedfrom,
	.NavFrame, .navigation-not-searchable, .noprint,
	.normdaten, .portal-bar, .printfooter,
	.refbegin-columns, .refbegin, .references, .reflist,
	.side-box, .sister-box, .sistersitebox,
	.vector-body-before-content, .vector-dropdown, .vector-header-container, .vector-page-toolbar, .vector-settings,
	*[role="navigation"]
	`)

	rmHeaderParent = cascadia.MustCompile(`#References, #Explanatory_notes, #Further_reading, #See_also, #External_links, #Notes, #Notable_people,
	#Primary_sources, #Secondary_sources, #Tertiary_sources, #Citations, #General_and_cited_sources, #Bibliography,
	#Referencias, #Véase_también, #Bibliografía, #Enlaces_externos, #Ciudades_hermanas, #Referencias_y_notas,
	#Literatur, #Weblinks, #Einzelnachweise, #Références, #Voir_aussi, #Bibliographie, #Liens_externes, #Notes_et_références,
	#Galerie_photos, #Referências, #Ver_também, #Bibliografia, #Ligações_externas, #Notas, #Annexes, #Article_connexe,
	#Voci_correlate, #Altri_progetti, #Collegamenti_esterni, #Note_e_riferimenti, #Note, #Riferimenti`)

	// Remove nodes for text extraction
	rmNodesText = []string{"script", "style", "#comment"}

	// HTML cleaning lookup for text/links/infobox and images.
	cleanHTMlLookup = [][]string{
		{"class", "noprint"},
		{"class", "noviewer"},
		{"class", "navigation-only"},
		{"style", "display:none"},
	}
)

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

// ArticlePartsGetter defines an interface for extracting structured parts
// (e.g., sections, tables) from an article HTML selection.
// The returned images slice contains all images from the document (including inside tables).
type ArticlePartsGetter interface {
	GetArticleParts(sel *goquery.Selection) ([]*Part, []*OrderedTable, []*Image, error)
}

// ParagraphsGetter interface to expose method to get the paragraphs of a part of the page.
type ParagraphsGetter interface {
	GetParagraphs(sel *goquery.Selection, baseUrl string) []*Part
}

// ListItemsGetter interface to expose method to get the list items of a part of the page.
type ListItemsGetter interface {
	GetListItems(sel *goquery.Selection, baseUrl string) []*Part
}

// ListsReplacer interface to expose method to replaces the HTML lists of a page with plain text lists.
type ListsReplacer interface {
	ReplaceLists(s *goquery.Selection)
}

// SuperscriptNumberToUnicodeConverter interface to expose method to convert superscript numbers to unicode.
type SuperscriptNumberToUnicodeConverter interface {
	ConvertSuperscriptNumberToUnicode(sel *goquery.Selection)
}

// CitationsGetter interface to expose method to get the citations of a selection.
type CitationsGetter interface {
	GetCitations(d *goquery.Selection) []*Citation
}

// ReferencesGetter interface to expose method to get the references of a page.
type ReferencesGetter interface {
	GetReferences(d *goquery.Selection) []*Reference
}

// API interface for the whole parser.
type API interface {
	ImagesGetter
	TextGetter
	LinksGetter
	InfoBoxesGetter
	ArticlePartsGetter
	ParagraphsGetter
	ListItemsGetter
	ReferencesGetter
	ListsReplacer
	SuperscriptNumberToUnicodeConverter
	CitationsGetter
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

// Tables maps table IDs to their corresponding TableWithHeaders structures.
type Tables map[string]*TableWithHeaders

// TableReference lists the referenced tables,  identified by a unique identifier.
type TableReference struct {
	Identifier      string  `json:"identifier"`
	ConfidenceScore float64 `json:"confidence_score,omitempty"`
}

// Part represents a part of the parsed page.
type Part struct {
	Name            string           `json:"name,omitempty"`
	Type            string           `json:"type,omitempty"`
	Value           string           `json:"value,omitempty"`
	Values          []string         `json:"values,omitempty"`
	HasParts        []*Part          `json:"has_parts,omitempty"`
	Images          []*Image         `json:"images,omitempty"`
	Links           []*Link          `json:"links,omitempty"`
	Citations       []*Citation      `json:"citations,omitempty"`
	TableReferences []TableReference `json:"table_references,omitempty"`
}

// Image represents an image that can be found on a Wikipedia page.
type Image struct {
	ContentUrl      string `json:"content_url,omitempty"`
	Identifier      string `json:"identifier,omitempty"` // SHA256 hex hash of name + srcset
	Name            string `json:"name,omitempty"`       // Includes localized namespace prefix (e.g. "File", "Datei").
	AlternativeText string `json:"alternative_text,omitempty"`
	Caption         string `json:"caption,omitempty"`
	Height          int    `json:"height,omitempty"`
	Width           int    `json:"width,omitempty"`
	EncodingFormat  string `json:"encoding_format,omitempty"`
	MediaType       string `json:"media_type,omitempty"` // Derived from the file extension.
}

// absoluteImageURL returns an absolute URL for an img src.
// Protocol-relative URLs (//...) get "https:" prepended; http/https URLs are returned as-is.
func absoluteImageURL(src string) string {
	if src == "" {
		return ""
	}
	u, err := url.Parse(src)
	if err != nil {
		return src
	}
	if u.Scheme != "" {
		return src
	}
	if strings.HasPrefix(src, "//") {
		u.Scheme = "https"
		return u.String()
	}
	return src
}

// imageIdentifierHash returns a SHA256 hex hash of name + srcset for a unique image identifier.
func imageIdentifierHash(name, srcset string) string {
	h := sha256.Sum256([]byte(name + srcset))
	return hex.EncodeToString(h[:])
}

// mimeTypeFromURL infers MIME type from the image URL path (file extension).
// Returns empty string if unknown. Supports jpeg, jpg, png, gif, webp, svg, tiff, tif, etc.
func mimeTypeFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return ""
	}
	dot := strings.LastIndex(u.Path, ".")
	if dot < 0 || dot >= len(u.Path)-1 {
		return ""
	}
	ext := strings.ToLower(u.Path[dot+1:])
	switch ext {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	case "tiff", "tif":
		return "image/tiff"
	case "bmp":
		return "image/bmp"
	case "ico":
		return "image/x-icon"
	default:
		return ""
	}
}

// Link represents a link that can be found on a Wikipedia page.
type Link struct {
	URL    string   `json:"url,omitempty"`
	Text   string   `json:"text,omitempty"`
	Images []*Image `json:"images,omitempty"`
}

// New returns the default parser with default parameters.
func New() API {
	return NewWithFlag(0.5)
}

// NewWithFlag allows explicit control over confidence threshold
func NewWithFlag(confidenceThreshold float64) API {
	return &Parser{ConfidenceThreshold: confidenceThreshold}
}

// Parser structure for defining the type.
type Parser struct {
	ConfidenceThreshold float64
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

// captionFromDataMw parses the data-mw attribute on the selection as JSON and returns the caption field if present.
func captionFromDataMw(sel *goquery.Selection) string {
	dta := sel.AttrOr("data-mw", "")
	if dta == "" {
		return ""
	}
	unescaped := stdhtml.UnescapeString(dta)
	result := gjson.Get(unescaped, "caption")
	if !result.Exists() {
		return ""
	}
	return strings.TrimSpace(result.String())
}

// isFileSpan returns true if the selection is a span with typeof indicating a Parsoid file (mw:File or mw:File/Frameless),
// and not mw:File/Frame or mw:File/Thumb (see https://www.mediawiki.org/wiki/Help:Images).
func isFileSpan(sel *goquery.Selection) bool {
	if goquery.NodeName(sel) != "span" {
		return false
	}

	typeof := sel.AttrOr("typeof", "")
	if strings.Contains(typeof, "mw:File/Thumb") {
		return false
	}
	if strings.Contains(typeof, "mw:File/Frameless") {
		return true
	}
	if strings.Contains(typeof, "mw:File/Frame") {
		return false
	}
	for _, t := range strings.Fields(typeof) {
		if t == "mw:File" {
			return true
		}
	}
	return false
}

// GetImages method to get images from html element.
func (p *Parser) GetImages(sel *goquery.Selection) []*Image {
	ims := []*Image{}

	// Use sel for full ancestor chain.
	sel.Find("img").Each(func(_ int, ssl *goquery.Selection) {
		// Skip images in tables, but allow infobox images
		if ssl.Closest("table").Length() > 0 && ssl.Closest(".infobox, .infobox_v2, .infobox_v3, .sinottico").Length() == 0 {
			return
		}
		if ssl.Closest(".noviewer").Length() > 0 {
			return
		}

		// Skip these
		if ssl.Closest("[class*='noprint']").Length() > 0 ||
			ssl.Closest("[class*='navigation-only']").Length() > 0 ||
			ssl.Closest("[style*='display:none']").Length() > 0 {
			return
		}
		img := new(Image)
		src := ssl.AttrOr("src", "")
		if src == "" {
			return
		}
		img.ContentUrl = absoluteImageURL(src)
		img.Name = strings.TrimPrefix(ssl.AttrOr("resource", ""), "./")
		img.Identifier = imageIdentifierHash(img.Name, ssl.AttrOr("srcset", ""))
		img.EncodingFormat = mimeTypeFromURL(src)
		img.MediaType = ssl.AttrOr("data-file-type", "")
		img.AlternativeText = ssl.AttrOr("alt", "")
		img.Width, _ = strconv.Atoi(ssl.AttrOr("width", "0"))
		img.Height, _ = strconv.Atoi(ssl.AttrOr("height", "0"))

		// we consider images smaller than 16x16 to be an icon, which we want to exclude.
		if (img.Width > 0 && img.Width < 16) || (img.Height > 0 && img.Height < 16) {
			return
		}

		figureParent := ssl.ParentsFiltered("figure")

		img.Caption = ssl.ParentsFiltered(".infobox-image").Find(".infobox-caption").Text()

		if len(img.Caption) == 0 {
			img.Caption = p.GetText(figureParent.Find("figcaption"))
		}

		// When image is not in a figure, check span file (mw:File or mw:File/Frameless).
		// For more details, see https://www.mediawiki.org/wiki/Help:Images.
		if figureParent.Length() == 0 {
			ssl.ParentsFiltered("span").EachWithBreak(func(_ int, span *goquery.Selection) bool {
				if !isFileSpan(span) {
					return true
				}
				if len(img.Caption) == 0 {
					img.Caption = captionFromDataMw(span)
				}
				if len(img.Caption) == 0 {
					img.Caption = span.Find("a.mw-file-description").AttrOr("title", "")
				}
				return false
			})
		}

		ims = append(ims, img)
	})

	return ims
}

// isInvisibleNode reports whether a node should be skipped during text traversal
func isInvisibleNode(nde *nethtml.Node) bool {
	if nde.Type == nethtml.CommentNode {
		return true
	}
	if nde.Type == nethtml.ElementNode {
		if slices.Contains(rmNodesText, nde.Data) {
			return true
		}
		for _, attr := range nde.Attr {
			for _, rule := range cleanHTMlLookup {
				if attr.Key == rule[0] && strings.Contains(attr.Val, rule[1]) {
					return true
				}
			}
		}
	}
	return false
}

// prevVisibleSibling walks backwards from n skipping wbr nodes, returning the first
// non-wbr previous sibling
func prevVisibleSibling(n *nethtml.Node) *nethtml.Node {
	for s := n.PrevSibling; s != nil; s = s.PrevSibling {
		if s.Type == nethtml.ElementNode && s.Data == "wbr" {
			continue
		}
		return s
	}
	return nil
}

// precedingWbr reports whether the closest previous sibling of n (skipping nothing)
// is a wbr element
func precedingWbr(n *nethtml.Node) bool {
	for s := n.PrevSibling; s != nil; s = s.PrevSibling {
		if s.Type == nethtml.ElementNode && s.Data == "wbr" {
			return true
		}
		if s.Type != nethtml.TextNode || strings.TrimSpace(s.Data) != "" {
			break
		}
	}
	return false
}

// GetText method to get text from html element, ign (ignore) is a list of HTML tags to skip when building the text string.
func (p *Parser) GetText(sel *goquery.Selection, ign ...string) string {
	var fnc func(*nethtml.Node)
	buf := new(bytes.Buffer)

	fnc = func(nde *nethtml.Node) {
		// Skip nodes
		if isInvisibleNode(nde) {
			return
		}

		// Skip wbr inline
		if nde.Type == nethtml.ElementNode && nde.Data == "wbr" {
			// Still recurse into children
			for sbg := nde.FirstChild; sbg != nil; sbg = sbg.NextSibling {
				fnc(sbg)
			}
			return
		}

		if nde.Type == nethtml.TextNode {
			// If this text node is directly preceded by a wbr, it must be concatenated
			// to whatever was written before
			if precedingWbr(nde) {
				// Remove the trailing space that the previous text node wrote
				b := buf.Bytes()
				if len(b) > 0 && b[len(b)-1] == ' ' {
					buf.Truncate(buf.Len() - 1)
				}
				_, _ = buf.WriteString(nde.Data + " ")
				return
			}

			nds := []*nethtml.Node{}
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

			// if the previous sibling is an html element (excluding wbr), add a space
			for _, tnd := range nds {
				if tnd != nil {
					pvs := prevVisibleSibling(tnd)
					if pvs != nil && pvs.Type == nethtml.ElementNode {
						if !strings.HasPrefix(tnd.Data, " ") {
							_, _ = buf.WriteString(" ")
						}
					}
				}
			}

			_, _ = buf.WriteString(nde.Data + " ")
		}

		if nde.FirstChild != nil {
			for sbg := nde.FirstChild; sbg != nil; sbg = sbg.NextSibling {
				// skip inner table rows
				if sbg.Type == nethtml.ElementNode && slices.Contains(ign, sbg.Data) {
					continue
				}

				fnc(sbg)
			}
		}
	}

	for _, nde := range sel.Nodes {
		fnc(nde)
	}

	// clean up the text
	str := buf.String()
	str = allInOneEmpty.ReplaceAllString(str, "")
	str = allInOneDelimiter.ReplaceAllString(str, "$1")
	if str != "" {
		str = allInOneSpace.ReplaceAllString(str, " ")

		// Preserve newlines but collapse multiple spaces/tabs within each line
		lines := strings.Split(str, "\n")
		for i, line := range lines {
			line = spacesAndTabs.ReplaceAllString(line, " ")
			lines[i] = strings.TrimSpace(line)
		}
		nonEmptyLines := []string{}
		for _, line := range lines {
			if line != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		str = strings.Join(nonEmptyLines, "\n")
	}
	str = spacesAroundParenthesis.ReplaceAllString(str, "$1$2")

	return strings.TrimSpace(str)
}

// getLinksInner is the shared implementation for GetLinks and GetLinksIgnoringSublists.
func (p *Parser) getLinksInner(bul string, find func() *goquery.Selection) []*Link {
	lks := []*Link{}

	find().Each(func(_ int, ssl *goquery.Selection) {
		href := ssl.AttrOr("href", "")

		if strings.HasPrefix(href, "./File:") {
			return
		}

		lnk := new(Link)
		lnk.Text = p.GetText(ssl)

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

// GetLinks parses list of links based on provided element collection.
func (p *Parser) GetLinks(sel *goquery.Selection, bul string) []*Link {
	if len(bul) == 0 {
		bul = p.getBaseURL(sel)
	}
	return p.getLinksInner(bul, func() *goquery.Selection {
		return sel.Find("a")
	})
}

// GetLinksIgnoringSublists is like GetLinks but never recurses into ul, ol, or dl nodes.
func (p *Parser) GetLinksIgnoringSublists(sel *goquery.Selection, bul string) []*Link {
	if len(bul) == 0 {
		bul = p.getBaseURL(sel)
	}
	return p.getLinksInner(bul, func() *goquery.Selection {
		return sel.FindMatcher(utils.FindIgnoring("a", "ul, ol, dl"))
	})
}

func (p *Parser) getInfoBoxPart(sel *goquery.Selection, base string, pendingLabel string, pendingLinks []*Link) *Part {
	gtx := func(sel *goquery.Selection) string {
		str := p.GetText(sel, "tr", "ul", "ol", "dl")

		// remove any trailing delimiters from list item
		str = delimiterAtEnd.ReplaceAllString(str, "")
		return str
	}

	var nestedParts []*Part
	hasLists := false

	sel.FindMatcher(utils.TopOnly(".infobox-data ul, .infobox-data ol")).Each(func(_ int, list *goquery.Selection) {
		items := p.GetListItems(list, base)
		if len(items) > 0 {
			hasLists = true
			// Append list items directly without wrapping
			nestedParts = append(nestedParts, items...)
		}
	})

	sel.FindMatcher(utils.TopOnly(".infobox-data dl")).Each(func(_ int, dl *goquery.Selection) {
		items := p.GetListItems(dl, base)
		if len(items) > 0 {
			hasLists = true
			nestedParts = append(nestedParts, items...)
		}
	})

	p.findFirstLevel(sel.Find(".infobox-full-data"), "ul, ol, dl").Each(func(_ int, list *goquery.Selection) {
		items := p.GetListItems(list, base)
		if len(items) > 0 {
			hasLists = true
			nestedParts = append(nestedParts, items...)
		}
	})

	sel.ChildrenFiltered("td").Each(func(_ int, td *goquery.Selection) {
		if len(td.AttrOr("class", "")) > 0 {
			return // already handled above via infobox-data / infobox-full-data
		}
		_, hasColspan := td.Attr("colspan")
		if !hasColspan {
			return
		}
		p.findFirstLevel(td, "ul, ol, dl").Each(func(_ int, list *goquery.Selection) {
			items := p.GetListItems(list, base)
			if len(items) > 0 {
				hasLists = true
				nestedParts = append(nestedParts, items...)
			}
		})
	})

	lks := p.GetLinksIgnoringSublists(sel, base)
	if len(pendingLinks) > 0 {
		lks = append(pendingLinks, lks...)
	}

	lbl := gtx(utils.FindFirst(sel, ".infobox-label"))
	if len(lbl) == 0 {
		lbl = gtx(sel.Find(".infobox-label tr"))
	}
	lblFromPending := false
	if len(lbl) == 0 && len(pendingLabel) > 0 {
		lbl = pendingLabel
		lblFromPending = true
	}

	dta := ""
	if !hasLists {
		dta = gtx(utils.FindFirst(sel, ".infobox-data"))
		if len(dta) == 0 {
			dta = gtx(sel.Find(".infobox-data tr"))
		}
	}

	if len(dta) == 0 && !hasLists {
		dta = gtx(sel.Find(".infobox-full-data"))
		if len(dta) == 0 {
			dta = gtx(sel.Find(".infobox-full-data tr"))
		}
	}

	if len(dta) == 0 && !hasLists {
		ths := sel.ChildrenFiltered("th")
		tds := sel.ChildrenFiltered("td")

		if tds.Length() == 2 && ths.Length() == 0 {
			lbl = gtx(tds.First())
			dta = gtx(tds.Last())
		} else if tds.Length() == 1 && ths.Length() == 1 {
			lbl = gtx(ths.First())
			dta = gtx(tds.First())
		}

		if len(dta) == 0 && !hasLists {
			dta = gtx(sel)
		}
	}

	ims := p.GetImages(sel)

	if lblFromPending && len(dta) == 0 && !hasLists && len(ims) == 0 {
		lbl, dta = "", lbl
	}

	if len(lbl) == 0 && !hasLists && len(dta) == 0 && len(ims) == 0 {
		return nil
	}

	typ := PartTypeField

	if len(dta) == 0 && len(ims) > 0 || sel.Find(".infobox-image").Length() > 0 {
		typ = PartTypeImage
	}

	if hasLists {
		typ = PartTypeList
	}

	return &Part{
		Type:     typ,
		Name:     lbl,
		Value:    dta,
		Images:   ims,
		Links:    lks,
		HasParts: nestedParts,
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
	itl := p.GetText(sel.Find(".infobox-title"))

	// specific to French wikipedia and infobox_v3 template
	// we need to try to get the title from the entete div
	if sel.HasClass("infobox_v3") && len(itl) == 0 {
		itl = p.GetText(sel.Find("div.entete"))
	}

	var pendingLabel string
	var pendingLinks []*Link

	slc.Find("tr").Each(func(i int, s *goquery.Selection) {
		directCells := s.Children().Filter("th, td")
		scs := []*goquery.Selection{
			directCells.Filter(".infobox-above"),
			directCells.Filter(".infobox-below"),
			directCells.Filter(".infobox-header"),
			directCells.Filter(".infobox-subheader"),
			directCells.Filter(".sinottico_divisione"),
			directCells.Filter(".sinottico_testata"),
			directCells.Filter(".sinottico_piede"),
			directCells.Filter(".sinottico_piede2"),
		}

		for _, s := range scs {
			if nme := p.GetText(s); len(nme) > 0 {
				scn = &Part{
					Name: nme,
					Type: PartTypeSection,
				}
				ibx.HasParts = append(ibx.HasParts, scn)
				pendingLabel = ""
				pendingLinks = nil
				return
			}
		}

		ths := s.ChildrenFiltered("th")
		tds := s.ChildrenFiltered("td")
		if ths.Length() == 1 && tds.Length() == 0 {
			_, hasColspan := ths.First().Attr("colspan")
			hasBlockChild := ths.First().Find("div, br, img, table").Length() > 0
			if hasColspan && !hasBlockChild {
				pendingLabel = p.GetText(ths.First())
				pendingLinks = p.GetLinks(ths.First(), bse)
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
					scn.Images[0].Caption = p.GetText(utils.FindFirst(sel, ".legend"))
				}
			}

			ibx.HasParts = append(ibx.HasParts, scn)
		}

		rowHasOwnLabel := s.Find(".infobox-label").Length() > 0 ||
			(s.ChildrenFiltered("td").Length() == 2 && s.ChildrenFiltered("th").Length() == 0) ||
			(s.ChildrenFiltered("td").Length() == 1 && s.ChildrenFiltered("th").Length() == 1)
		if rowHasOwnLabel && len(pendingLabel) > 0 {
			headerPrt := &Part{
				Type:  PartTypeField,
				Value: pendingLabel,
				Links: pendingLinks,
			}
			scn.HasParts = append(scn.HasParts, headerPrt)
			pendingLabel = ""
			pendingLinks = nil
		}

		pendingLabelSnapshot := pendingLabel
		pendingLinksSnapshot := pendingLinks
		prt := p.getInfoBoxPart(s, bse, pendingLabel, pendingLinks)
		pendingLabel = ""
		pendingLinks = nil

		if prt == nil {
			if len(pendingLabelSnapshot) > 0 {
				scn.HasParts = append(scn.HasParts, &Part{
					Type:  PartTypeField,
					Value: pendingLabelSnapshot,
					Links: pendingLinksSnapshot,
				})
			}
			return
		}

		scn.HasParts = append(scn.HasParts, prt)
	})

	if len(pendingLabel) > 0 && scn != nil {
		scn.HasParts = append(scn.HasParts, &Part{
			Type:  PartTypeField,
			Value: pendingLabel,
			Links: pendingLinks,
		})
	}

	return ibx
}

// GetInfoBoxes function extracts all the infoboxes from the HTML.
func (p *Parser) GetInfoBoxes(sel *goquery.Selection) []*Part {
	ibs := make([]*Part, 0)

	sel.FindMatcher(infoboxSelector).Each(func(_ int, scn *goquery.Selection) {
		ibx := p.getInfoBox(scn)
		ibs = append(ibs, ibx)
	})

	return ibs
}

// getSections parses list of sections based provided element collection.
func (p *Parser) getSectionsRecursive(
	sel *goquery.Selection,
	parent *Part,
	sectionPath string,
	sectionTableCount map[string]int,
	baseUrl string,
) ([]*Part, []*OrderedTable, error) {
	parts := []*Part{}
	orderedTables := []*OrderedTable{}
	var errs []error

	sel.Children().Each(func(_ int, scn *goquery.Selection) {
		switch tag := goquery.NodeName(scn); tag {
		case "section":
			sectionName := p.getSectionName(scn)
			fullPath := sanitizeSectionPath(sectionName)
			if sectionPath != "" {
				fullPath = sectionPath + "." + fullPath
			}

			section := &Part{
				Type: PartTypeSection,
				Name: sectionName,
			}
			subParts, subTables, err := p.getSectionsRecursive(scn, section, fullPath, sectionTableCount, baseUrl)
			if err != nil {
				errs = append(errs, err)
			}
			section.HasParts = append(section.HasParts, subParts...)
			orderedTables = append(orderedTables, subTables...)
			parts = append(parts, section)

		case "table":
			table, err := p.GetTableRows(scn, Duplicate)
			if err != nil {
				errs = append(errs, err)
			}
			if len(table.Headers) == 0 && len(table.Rows) == 0 {
				return
			}
			tempParts := []*Part{}
			tempTables := []*OrderedTable{}
			promoteWrapperTables(&table, sectionPath, sectionTableCount, &tempTables, parent, &tempParts)
			parts = append(parts, tempParts...)
			orderedTables = append(orderedTables, tempTables...)

		case "p":
			if parent != nil {
				parent.HasParts = append(parent.HasParts, p.GetParagraphs(scn, baseUrl)...)
				subParts, subTables, err := p.getSectionsRecursive(scn, parent, sectionPath, sectionTableCount, baseUrl)
				if err != nil {
					errs = append(errs, err)
				}
				parts = append(parts, subParts...)
				orderedTables = append(orderedTables, subTables...)
			}

		case "span":
			if parent != nil && isFileSpan(scn) {
				images := p.GetImages(scn)
				if len(images) > 0 {
					parent.HasParts = append(parent.HasParts, &Part{
						Type:   PartTypeImage,
						Images: images,
					})
				}
			} else {
				subParts, subTables, err := p.getSectionsRecursive(scn, parent, sectionPath, sectionTableCount, baseUrl)
				if err != nil {
					errs = append(errs, err)
				}
				parts = append(parts, subParts...)
				orderedTables = append(orderedTables, subTables...)
			}

		case "ul", "ol":
			if parent == nil {
				break
			}
			if scn.HasClass("gallery") || scn.HasClass("mw-gallery-packed") {
				var images []*Image
				scn.Find("li.gallerybox").Each(func(_ int, li *goquery.Selection) {
					thumb := utils.FindFirst(li, "div.thumb")
					if thumb.Length() == 0 {
						return
					}
					liImages := p.GetImages(thumb)
					if len(liImages) == 0 {
						return
					}
					caption := p.GetText(utils.FindFirst(li, "div.gallerytext"))
					for _, img := range liImages {
						if len(img.Caption) == 0 && len(caption) > 0 {
							img.Caption = caption
						}
					}
					images = append(images, liImages...)
				})
				if len(images) > 0 {
					parent.HasParts = append(parent.HasParts, &Part{
						Type:   PartTypeGallery,
						Images: images,
					})
				}
				break
			}
			parent.HasParts = append(parent.HasParts, p.GetParagraphs(scn, baseUrl)...)

		case "dl":
			if parent != nil {
				parent.HasParts = append(parent.HasParts, p.GetParagraphs(scn, baseUrl)...)
			}

		case "figure":
			if parent == nil {
				return
			}
			images := p.GetImages(scn)
			if len(images) > 0 {
				parent.HasParts = append(parent.HasParts, &Part{
					Type:   PartTypeImage,
					Images: images,
				})
			}

		default:
			subParts, subTables, err := p.getSectionsRecursive(scn, parent, sectionPath, sectionTableCount, baseUrl)
			if err != nil {
				errs = append(errs, err)
			}
			parts = append(parts, subParts...)
			orderedTables = append(orderedTables, subTables...)
		}
	})

	// Compose error
	var err error
	if len(errs) == 1 {
		err = errs[0]
	} else if len(errs) > 1 {
		err = fmt.Errorf("multiple errors in getSectionsRecursive: %v", errs)
	}

	return parts, orderedTables, err
}

// GetArticleParts parses list of sections, tables and images in document order.
func (p *Parser) GetArticleParts(sel *goquery.Selection) ([]*Part, []*OrderedTable, []*Image, error) {
	allImages := p.GetImages(sel)

	slc := sel.Clone()
	baseUrl := p.getBaseURL(slc)

	// Clean up the document
	slc.FindMatcher(utils.TopOnlyMatcher(infoboxSelector)).Remove()
	slc.FindMatcher(utils.TopOnly(`section > div.reflist, section > *[id="External_links"]`)).Parent().Remove()
	slc.FindMatcher(utils.TopOnlyMatcher(rmSelector)).Remove()
	slc.FindMatcher(utils.TopOnlyMatcher(rmHeaderParent)).Parent().Remove()

	slc.Find("section").Each(func(i int, s *goquery.Selection) {
		if s.Children().Length() == 1 {
			firstChild := s.Children().First()
			if firstChild.Is("h1,h2,h3,h4,h5,h6") {
				s.Remove()
			}
		}
	})

	// Get initial parts and tables, now with error propagation
	parts, tables, err := p.getSectionsRecursive(
		slc.FindMatcher(utils.TopOnly("body")),
		nil,
		"",
		make(map[string]int),
		baseUrl,
	)
	// Filter tables by confidence
	tables = p.filterOrderedTablesByConfidence(tables)
	return parts, tables, allImages, err
}

// removeInlineCitation removes citation numbers (example `[12]`) from the text while preserving the appropriate spaces, and removes extea space before any fullstop
func (p *Parser) removeInlineCitation(sel *goquery.Selection, ign ...string) string {
	txt := editText.ReplaceAllString(isRef.ReplaceAllString(p.GetText(sel, ign...), " "), " ")
	return strings.TrimSpace(strings.ReplaceAll(txt, " .", "."))
}

// isEmptyMarkdownList checks if a markdown list contains only empty items
// Returns true for strings like "  - \n  - \n  - " (only whitespace and hyphens)
func isEmptyMarkdownList(markdown string) bool {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines
		if trimmed == "" {
			continue
		}
		// If any line has content beyond just a hyphen, it's not empty
		if trimmed != "-" && !strings.HasPrefix(trimmed, "- ") {
			return false
		}
		// Check if line is just hyphen with optional whitespace
		withoutPrefix := strings.TrimPrefix(strings.TrimSpace(line), "-")
		if strings.TrimSpace(withoutPrefix) != "" {
			return false
		}
	}
	return true
}

// GetParagraphs parses list of paragraphs based on provided element collection.
func (p *Parser) GetParagraphs(sel *goquery.Selection, baseUrl string) []*Part {
	pgs := []*Part{}

	bul := baseUrl

	if len(baseUrl) == 0 {
		bul = p.getBaseURL(sel)
	}

	sel.Each(func(_ int, scn *goquery.Selection) {
		pgf := new(Part)

		switch tag := goquery.NodeName(scn); tag {
		case "p":
			pgf.Type = PartTypeParagraph
			pgf.Citations = p.GetCitations(scn)
			pgf.Value = p.removeInlineCitation(scn)
			pgf.Links = p.GetLinks(scn, bul)

		case "ol", "ul":
			if tag == "ol" {
				pgf.Type = PartTypeOrderedList
			} else {
				pgf.Type = PartTypeList
			}
			pgf.HasParts = p.GetListItems(scn, bul)

		case "dl":
			pgf.Type = PartTypeDefinitionList
			pgf.HasParts = p.GetListItems(scn, bul)

		}

		if len(pgf.Value) != 0 || len(pgf.Links) != 0 || len(pgf.HasParts) != 0 {
			pgs = append(pgs, pgf)
		}
	})

	return pgs
}

// GetListItems parses list of items based on provided element collection. (handles nested lists and definition lists)
func (p *Parser) GetListItems(sel *goquery.Selection, baseUrl string) []*Part {
	lis := []*Part{}
	bul := baseUrl
	if len(bul) == 0 {
		bul = p.getBaseURL(sel)
	}

	listType := goquery.NodeName(sel)

	// Handle definition lists specially
	if listType == "dl" {
		// Track which dd nodes have been claimed by a dt
		claimedDds := map[*nethtml.Node]bool{}

		sel.ChildrenFiltered("dt").Each(func(_ int, child *goquery.Selection) {
			term := new(Part)
			term.Type = PartTypeDefinitionTerm
			term.Value = p.removeInlineCitation(child, "ul", "ol", "dl")
			term.Links = p.GetLinksIgnoringSublists(child, bul)

			// Collect associated definitions
			definitions := []*Part{}
			for next := child.Nodes[0].NextSibling; next != nil; next = next.NextSibling {
				if next.Type != nethtml.ElementNode {
					continue
				}

				if next.Data == "dt" {
					break
				}

				if next.Data != "dd" {
					continue
				}

				// Process dd
				claimedDds[next] = true
				ddSel := goquery.NewDocumentFromNode(next).Selection

				def := &Part{
					Type:  PartTypeDefinition,
					Value: p.removeInlineCitation(ddSel, "ul", "ol", "dl"),
					Links: p.GetLinksIgnoringSublists(ddSel, bul),
				}

				// Handle nested lists in dd
				nestedLists := ddSel.ChildrenFiltered("ul, ol, dl")
				if nestedLists.Length() > 0 {
					nestedLists.Each(func(_ int, nested *goquery.Selection) {
						nodeName := goquery.NodeName(nested)
						nestedType := PartTypeList
						switch nodeName {
						case "dl":
							nestedType = PartTypeDefinitionList
						case "ol":
							nestedType = PartTypeOrderedList
						}
						def.HasParts = append(def.HasParts, &Part{
							Type:     nestedType,
							HasParts: p.GetListItems(nested, bul),
						})
					})
				}

				definitions = append(definitions, def)
			}

			term.HasParts = definitions

			// Handle nested lists in dt
			dtNestedLists := child.ChildrenFiltered("ul, ol, dl")
			if dtNestedLists.Length() > 0 {
				dtNestedLists.Each(func(_ int, nested *goquery.Selection) {
					nodeName := goquery.NodeName(nested)
					nestedType := PartTypeList
					switch nodeName {
					case "dl":
						nestedType = PartTypeDefinitionList
					case "ol":
						nestedType = PartTypeOrderedList
					}
					term.HasParts = append(term.HasParts, &Part{
						Type:     nestedType,
						HasParts: p.GetListItems(nested, bul),
					})
				})
			}

			if len(term.Value) != 0 || len(term.Links) != 0 || len(term.HasParts) != 0 {
				lis = append(lis, term)
			}
		})

		// Handle orphaned dd elements (no associated dt)
		sel.ChildrenFiltered("dd").Each(func(_ int, child *goquery.Selection) {
			if claimedDds[child.Nodes[0]] {
				return
			}

			def := &Part{
				Type:  PartTypeDefinition,
				Value: p.removeInlineCitation(child, "ul", "ol", "dl"),
				Links: p.GetLinksIgnoringSublists(child, bul),
			}

			nestedLists := child.ChildrenFiltered("ul, ol, dl")
			if nestedLists.Length() > 0 {
				nestedLists.Each(func(_ int, nested *goquery.Selection) {
					nodeName := goquery.NodeName(nested)
					nestedType := PartTypeList
					switch nodeName {
					case "dl":
						nestedType = PartTypeDefinitionList
					case "ol":
						nestedType = PartTypeOrderedList
					}
					def.HasParts = append(def.HasParts, &Part{
						Type:     nestedType,
						HasParts: p.GetListItems(nested, bul),
					})
				})
			}

			if len(def.Value) != 0 || len(def.Links) != 0 || len(def.HasParts) != 0 {
				lis = append(lis, def)
			}
		})

		return lis
	}

	// Handle regular lists (ul, ol) with nesting support
	sel.ChildrenFiltered("li").Each(func(_ int, scn *goquery.Selection) {
		prt := new(Part)
		prt.Type = PartTypeListItem
		prt.Value = p.removeInlineCitation(scn, "ul", "ol", "dl")
		prt.Links = p.GetLinksIgnoringSublists(scn, bul)

		// Recursively process nested lists
		nestedLists := scn.ChildrenFiltered("ul, ol, dl")
		if nestedLists.Length() > 0 {
			nestedLists.Each(func(_ int, nested *goquery.Selection) {
				nodeName := goquery.NodeName(nested)
				nestedType := PartTypeList
				switch nodeName {
				case "dl":
					nestedType = PartTypeDefinitionList
				case "ol":
					nestedType = PartTypeOrderedList
				}
				items := p.GetListItems(nested, bul)
				if len(items) == 0 {
					return
				}

				prt.HasParts = append(prt.HasParts, &Part{
					Type:     nestedType,
					HasParts: items,
				})
			})
		}

		if len(prt.Value) != 0 || len(prt.Links) != 0 || len(prt.HasParts) != 0 {
			lis = append(lis, prt)
		}
	})

	return lis
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
func (p *Parser) getList(b *bytes.Buffer, n *nethtml.Node, depth int) {
	if n.Type == nethtml.ElementNode && (n.Data == "li" || n.Data == "dt" || n.Data == "dd") {
		hasText := p.hasTextContent(n)

		if hasText {
			b.WriteString(strings.Repeat("  ", depth) + "- ")
		}

		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == nethtml.ElementNode && (c.Data == "ul" || c.Data == "ol" || c.Data == "dl") {
				if hasText {
					b.WriteString("\n")
				}
				p.getList(b, c, depth+1)
			} else {
				p.getList(b, c, depth)
			}
		}

		if hasText {
			b.WriteString("\n")
		}
		return
	}

	if n.Type == nethtml.TextNode {
		b.WriteString(n.Data)
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == nethtml.ElementNode && (c.Data == "ul" || c.Data == "ol" || c.Data == "dl") {
			b.WriteString("\n")
			p.getList(b, c, depth+1)
		} else {
			p.getList(b, c, depth)
		}
	}
}

// hasTextContent checks if a node has any meaningful text content (not just whitespace, styles, scripts, etc.)
func (p *Parser) hasTextContent(n *nethtml.Node) bool {
	if n.Type == nethtml.TextNode {
		text := strings.TrimSpace(n.Data)
		return len(text) > 0
	}

	// Skip non-content elements
	if n.Type == nethtml.ElementNode {
		switch n.Data {
		case "script", "style", "link", "meta", "noscript":
			return false
		}
	}

	// Recursively check children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if p.hasTextContent(c) {
			return true
		}
	}

	return false
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

		if prv.Type == nethtml.TextNode && nxt.Type == nethtml.TextNode {
			prv.Data += nxt.Data
			nxt.Parent.RemoveChild(nxt)
		}

		wbr.Remove()
	})
}

// ConvertSuperscriptNumberToUnicode converts superscript numbers to unicode.
func (p *Parser) ConvertSuperscriptNumberToUnicode(sel *goquery.Selection) {
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

func (p *Parser) getSectionName(sel *goquery.Selection) string {
	name := isRef.ReplaceAllString(p.GetText(utils.FindFirst(sel, "h2,h3,h4,h5,h6,h7")), " ")
	if name == "" && sel.AttrOr("data-mw-section-id", "") == "0" {
		name = "Abstract"
	}

	return name
}

func sanitizeSectionPath(name string) string {
	// Sanitize for use in IDs
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = regexp.MustCompile(`[^\w-]`).ReplaceAllString(name, "")
	return name
}

// isWrapperTable checks if a table is just a wrapper
func isWrapperTable(table *TableWithHeaders) bool {
	if len(table.Rows) == 0 {
		return false
	}

	hasInner := false

	for _, row := range table.Rows {
		for _, cell := range row {
			val := strings.TrimSpace(cell.Value)
			if val != "" {
				// If any cell has a real value, not a wrapper
				return false
			}
			if len(cell.Inner) > 0 {
				hasInner = true
			}
			// If cell has no value and no inner, that's ok for a wrapper row
			// (do nothing, just keep looping)
		}
	}
	return hasInner
}

// promoteWrapperTables promotes inner tables from wrapper tables to top-level tables with new IDs.
func promoteWrapperTables(table *TableWithHeaders,
	sectionKey string,
	sectionTableCount map[string]int,
	orderedTables *[]*OrderedTable,
	parent *Part,
	parts *[]*Part,
) {
	if sectionKey == "" {
		sectionKey = "top"
	}
	if isWrapperTable(table) {
		// For each inner table, promote and add to orderedTables
		for _, row := range table.Rows {
			for _, cell := range row {
				for _, inner := range cell.Inner {
					sectionTableCount[sectionKey]++
					newID := fmt.Sprintf("%s_table%d", sectionKey, sectionTableCount[sectionKey])

					// Add to orderedTables
					*orderedTables = append(*orderedTables, &OrderedTable{
						ID:    newID,
						Table: &inner, // #nosec G601
					})

					// Create a Part referencing this table
					tablePart := &Part{
						Type: PartTypeTable,
						TableReferences: []TableReference{{
							Identifier:      newID,
							ConfidenceScore: inner.ConfidenceScore,
						}},
					}

					if parent != nil {
						parent.HasParts = append(parent.HasParts, tablePart)
					} else if parts != nil {
						*parts = append(*parts, tablePart)
					}
				}
			}
		}
	} else {
		// Normal table: add to orderedTables
		sectionTableCount[sectionKey]++
		tableID := fmt.Sprintf("%s_table%d", sectionKey, sectionTableCount[sectionKey])

		*orderedTables = append(*orderedTables, &OrderedTable{
			ID:    tableID,
			Table: table,
		})

		tablePart := &Part{
			Type: PartTypeTable,
			TableReferences: []TableReference{{
				Identifier:      tableID,
				ConfidenceScore: table.ConfidenceScore,
			}},
		}

		if parent != nil {
			parent.HasParts = append(parent.HasParts, tablePart)
		} else if parts != nil {
			*parts = append(*parts, tablePart)
		}
	}
}

func (p *Parser) filterOrderedTablesByConfidence(tables []*OrderedTable) []*OrderedTable {
	threshold := p.ConfidenceThreshold
	if threshold == 0 {
		threshold = 0.5 // default threshold if not set
	}
	filtered := make([]*OrderedTable, 0, len(tables))
	for _, ot := range tables {
		if ot.Table != nil && ot.Table.ConfidenceScore >= threshold {
			filtered = append(filtered, ot)
		}
	}
	return filtered
}
