package parser

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"wikimedia-enterprise/general/parser/utils"
)

const defaultTextType = "text"

const referenceSpanTextCleanTags = "style"

var citationTags = []string{"cite", "cita", "lien"}

type DataMwRef struct {
	Name                 string            `json:"name"`
	Attrs                map[string]string `json:"attrs"`
	MainRef              string            `json:"mainRef"`
	MainBody             string            `json:"mainBody"`
	IsSubRefWithMainBody int               `json:"isSubRefWithMainBody"`
}

type DataMw struct {
	Parts []struct {
		Template struct {
			Target struct {
				Wt string `json:"wt"`
			} `json:"target"`
			Params map[string]struct {
				Wt string `json:"wt"`
			} `json:"params"`
		} `json:"template"`
	} `json:"parts"`
}

type Citation struct {
	Identifier string `json:"identifier,omitempty"`
	Group      string `json:"group,omitempty"`
	Text       string `json:"text,omitempty"`
}

type RawData struct {
	Value string  `json:"value,omitempty"`
	Links []*Link `json:"links,omitempty"`
}

type Reference struct {
	Identifier      string            `json:"identifier,omitempty"`
	Group           string            `json:"group,omitempty"`
	Type            string            `json:"type,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Text            *RawData          `json:"text,omitempty"`
	Source          *RawData          `json:"source,omitempty"`
	ParentReference string            `json:"parent_reference,omitempty"` // backlink to parent
}

type CiterefCache struct {
	cache       map[string]*goquery.Selection
	initialized bool
}

func cleanType(refType string) string {
	cleaned := strings.TrimSpace(refType)
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.ToLower(cleaned)
	return cleaned
}

// ParseCOinSMetadata parses out metadata found next to <cite> tags in references, following COinS format, e.g.:
//
//	<cite class="citation cs1"></cite><span title="key=value&..."></span>
func ParseCOinSMetadata(data string) map[string]string {
	// Reference: https://web.archive.org/web/20190501061437/http://epub.mimas.ac.uk/openurl/KEV_Guidelines_Ap-200706.html#sectB_5
	// Keys may not contain & or =, and values will have them encoded as %26 and %3D.
	items := strings.Split(data, "&")
	result := map[string]string{}
	for _, item := range items {
		if !strings.Contains(item, "=") {
			continue
		}
		kv := strings.SplitN(item, "=", 2)
		if kv[0] == "" {
			continue
		}
		val, err := url.QueryUnescape(kv[1])
		if err != nil {
			// Fallback to raw value if decoding fails
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = val
		}
	}
	if len(result) > 0 {
		result["is_coins"] = "true"
	}
	return result
}

// ExtractCOinSTypeAndMetadata attempts to parse <cite> tags in references and extract COinS metadata
func ExtractCOinSTypeAndMetadata(s *goquery.Selection) (string, map[string]string, bool) {
	citeTag := utils.FindFirst(s, "cite")
	if class, exists := citeTag.Attr("class"); exists && class != "" {
		classes := strings.Fields(class)
		// Look for the type after "citation"
		for i, cls := range classes {
			if cls == "citation" && i+1 < len(classes) {
				var metadata map[string]string
				span := citeTag.Next()
				if title, exists := span.Attr("title"); span.Is("span") && exists {
					metadata = ParseCOinSMetadata(title)
				}
				return cleanType(classes[i+1]), metadata, true
			}
		}
	}
	return "", map[string]string{}, false
}

// extractCiteTypeFromDataMw extracts the citation type from data-mw attribute
func extractCiteTypeFromDataMw(dataMwAttr string) (string, bool) {
	if dataMwAttr == "" {
		return "", false
	}
	unescaped := html.UnescapeString(dataMwAttr)
	var dataMw DataMw
	if err := json.Unmarshal([]byte(unescaped), &dataMw); err != nil {
		return "", false
	}
	if dataMw.Parts == nil {
		return "", false
	}
	for _, part := range dataMw.Parts {
		//TODO: Refactor if statement to add nil checks

		if target := part.Template.Target.Wt; target != "" {

			wt := target

			words := strings.Fields(wt)
			//TODO: We need to have cite cita and lien as part of parameterization of parser from language
			//This logic needs to be looked at for example article: Calvi for french wikipedia
			if len(words) > 1 && slices.Contains(citationTags, cleanType(words[0])) {
				return cleanType(words[1]), true
			} else if len(words) > 0 {
				return cleanType(words[0]), true
			}
		}
	}
	return "", false
}

// GetCitations gets the citations of a selection.
func (p *Parser) GetCitations(doc *goquery.Selection) []*Citation {
	citations := make([]*Citation, 0)
	doc.Find("sup.reference").Each(func(i int, s *goquery.Selection) {
		refText := s.Text()
		val, _ := s.Attr("class")
		if strings.Contains(val, "cite_virgule") {
			return
		}
		link := utils.FindFirst(s, "a")
		group := ""
		if groupAttr, exists := link.Attr("data-mw-group"); exists && groupAttr != "" {
			return
		}
		var id string
		if href, ok := link.Attr("href"); ok {
			index := strings.LastIndex(href, "#")
			id = href[index+1:]
		} else {
			return
		}
		_, identifier := p.cleanIdentifier(id)
		citation := &Citation{
			Identifier: identifier,
			Group:      group,
			Text:       refText,
		}
		citations = append(citations, citation)
	})
	return citations
}

func (p *Parser) getSpanText(t *goquery.Selection) string {
	if utils.FindFirst(t, referenceSpanTextCleanTags).Length() > 0 {
		t = t.Clone()
		t.FindMatcher(utils.TopOnly(referenceSpanTextCleanTags)).Remove()
	}
	return strings.TrimSpace(t.Text())
}

func (p *Parser) getIdentifierNumber(identifier string) string {
	idx := strings.LastIndex(identifier, "-")
	if idx >= 0 {
		return identifier[idx+1:]
	}
	return identifier
}

func (p *Parser) cleanIdentifier(val string) (bool, string) {
	processed := strings.Contains(val, "FOOTNOTE")
	cleanedId := p.getIdentifierNumber(val)
	return processed, fmt.Sprintf("cite_note-%s", cleanedId)
}

func (p *Parser) extractFootnote(d *goquery.Selection, links []*Link, baseUrl string, citerefCache *CiterefCache) *RawData {
	var citeRef string
	for _, link := range links {
		if strings.Contains(link.URL, "CITEREF") {
			citeRef = link.URL[strings.LastIndex(link.URL, "#")+1:]
		}
	}
	if citeRef == "" {
		return nil
	}

	if !citerefCache.initialized {
		citerefCache.cache = map[string]*goquery.Selection{}
		d.Find("[id*=CITEREF]").Each(func(i int, s *goquery.Selection) {
			citerefCache.cache[s.AttrOr("id", "")] = s
		})
		citerefCache.initialized = true
	}

	sel, ok := citerefCache.cache[citeRef]
	if !ok || sel.Length() == 0 {
		return nil
	}

	t := sel.Text()
	if len(t) > 0 {
		return &RawData{
			Value: t,
			Links: p.GetLinks(sel, baseUrl),
		}
	}
	return nil
}

// GetReferences extracts all references from a Wikipedia article
func (p *Parser) GetReferences(d *goquery.Selection) []*Reference {
	references := make(map[string][]*Reference)
	baseUrl := p.getBaseURL(d)

	var dataMwCache map[string]*goquery.Selection
	citerefCache := CiterefCache{}

	d.Find("ol.references").Each(func(i int, s *goquery.Selection) {
		if groupAttr, exists := s.Attr("data-mw-group"); exists && groupAttr != "" {
			return
		}
		if s.ParentsFiltered(".navbox, .infobox").Length() > 0 {
			// Ignore navbox and infobox references, they are very often footnotes.
			return
		}
		s.Find("li").Each(func(i int, s *goquery.Selection) {
			refSpan := s.Find("span.reference-text, span.mw-reference-text")
			if refSpan.Length() == 0 {
				return
			}
			if id, exists := s.Attr("id"); exists {
				group := ""
				hasFootnote, identifier := p.cleanIdentifier(id)

				var source *RawData
				links := p.GetLinks(refSpan, baseUrl)
				if hasFootnote {
					source = p.extractFootnote(d, links, baseUrl, &citerefCache)
				}

				rawData := &RawData{
					Value: p.getSpanText(refSpan),
					Links: links,
				}
				reference := &Reference{
					Identifier: identifier,
					Group:      group,
					Type:       defaultTextType, // default type
					Metadata:   make(map[string]string),
					Text:       rawData,
					Source:     source,
				}

				datamwSelection := utils.FindFirst(refSpan, "[data-mw]:not(abbr)")

				// First try to get type from HTML class
				if citeType, metadata, found := ExtractCOinSTypeAndMetadata(refSpan); found {
					reference.Type = citeType
					reference.Metadata = metadata
				} else {
					// If not found in class, try to get from data-mw
					if dataMwAttr, exists := datamwSelection.Attr("data-mw"); exists {
						if citeType, found := extractCiteTypeFromDataMw(dataMwAttr); found {
							reference.Type = citeType
						}
					}
				}

				// Extract metadata from data-mw attribute, unless already extracted in a previous case.
				if dataMwAttr, exists := datamwSelection.Attr("data-mw"); exists {
					unescaped := html.UnescapeString(dataMwAttr)
					var dataMw DataMw
					if err := json.Unmarshal([]byte(unescaped), &dataMw); err == nil {
						if dataMw.Parts != nil {
							for _, part := range dataMw.Parts {
								// Extract and flatten the parameters
								metadata := make(map[string]string)
								//TODO: add nil checks to avoid runtime errors
								for key, param := range part.Template.Params {
									metadata[cleanType(key)] = param.Wt
								}

								// NOTE: we override COinS metadata on purpose, hoping the other metadata will be
								//  more relevant.
								reference.Metadata = metadata
							}
						}
					}
				}

				// --- NEW: detect parent ref via Parsoid data-mw ---
				if dataMwCache == nil {
					dataMwCache = map[string]*goquery.Selection{}
					d.Find("sup.mw-ref a").Each(func(i int, s *goquery.Selection) {
						href := s.AttrOr("href", "")
						if href == "" {
							return
						}
						dataMwCache[href] = s
					})
				}
				if sup, ok := dataMwCache["#"+id]; ok {
					sup := sup.Closest("sup")
					if sup.Length() > 0 {
						if dataMwAttr := sup.AttrOr("data-mw", ""); dataMwAttr != "" {
							var refData DataMwRef
							if err := json.Unmarshal([]byte(html.UnescapeString(dataMwAttr)), &refData); err == nil {
								if refData.IsSubRefWithMainBody == 1 && refData.MainBody != "" {
									_, parentIdentifier := p.cleanIdentifier(refData.MainBody)
									reference.ParentReference = parentIdentifier
								}
							}
						}
					}
				}

				if strings.EqualFold(reference.Type, "harvsp") || strings.EqualFold(reference.Type, "harvsb") {
					for _, lnk := range links {
						if strings.Contains(lnk.URL, "#") {
							id := lnk.URL[strings.LastIndex(lnk.URL, "#")+1:]
							sel := d.Find(fmt.Sprintf(`[id="%s"]`, id)).First()
							if len(sel.Text()) > 0 {
								reference.Source = &RawData{
									Value: sel.Text(),
									Links: p.GetLinks(sel, baseUrl),
								}
							}
						}
					}
				}

				references[group] = append(references[group], reference)
			}
		})
	})

	return references[""]
}
