package utils_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"wikimedia-enterprise/general/parser/utils"
)

func mustParse(src string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(src))
	if err != nil {
		panic(err)
	}
	return doc
}

func TestFilterMatcher(t *testing.T) {
	doc := mustParse(`<div><span id="a"></span><span id="b"></span></div>`)
	nodes := doc.Find("span").Nodes

	calls := 0
	m := utils.FilterMatcher(func(n *html.Node) bool {
		calls++
		return utils.AttrOr(n, "id", "") == "b"
	})

	assert.False(t, m.Match(nodes[0]))
	assert.True(t, m.Match(nodes[1]))
	assert.Equal(t, 2, calls)
}

func TestAttrOr(t *testing.T) {
	doc := mustParse(`<div><span id="a"></span><span style="b"></span></div>`)
	def := "DEFAULT"

	assert.Equal(t, "DEFAULT", utils.AttrOr(nil, "id", def))

	n := doc.Find("span").Nodes

	assert.Equal(t, "a", utils.AttrOr(n[0], "id", def))
	assert.Equal(t, "b", utils.AttrOr(n[1], "style", def))
}

func TestTopOnlyMatcher(t *testing.T) {
	// Outer .rm contains inner .rm; TopOnly should only return the outer.
	doc := mustParse(`
		<div id="root">
			<div class="rm" id="outer">
				<div class="rm" id="inner"></div>
			</div>
			<div class="rm" id="sibling"></div>
		</div>
	`)

	root := doc.Find("#root").Nodes[0]
	m := utils.TopOnly(".rm")

	got := m.MatchAll(root)
	assert.Equal(t, 2, len(got))

	assert.Equal(t, "outer", utils.AttrOr(got[0], "id", ""))
	assert.Equal(t, "sibling", utils.AttrOr(got[1], "id", ""))
}

func TestFindFirst(t *testing.T) {
	doc := mustParse(`
		<div id="root">
			<div class="x"><span class="t">first</span></div>
			<div class="x"><span class="t">second</span></div>
		</div>
	`)
	root := doc.Find("#root")

	got := utils.FindFirst(root, ".t")
	assert.Equal(t, 1, got.Length())
	assert.Equal(t, "first", strings.TrimSpace(got.Text()))
}

// Prevent DCE from tampering with benchmarks.
var sink *goquery.Selection

func BenchmarkFindFirst(b *testing.B) {
	doc := mustParse(`
		<div id="root">
			<div class="x"><span class="t">first</span></div>
			<div class="x"><span class="t">second</span></div>
` + strings.Repeat(`<div><span class="t"></span></div>`, 500) +
		`
		</div>
	`)
	root := doc.Find("#root")
	m, _ := cascadia.Compile(".t")

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sink = utils.FindFirstMatcher(root, m)
	}
}

// BenchmarkFindThenFirst shows how even goquery.Single does a full tree search in certain cases.
// Here, for example, the matcher will be called on every child of #root, saving no time at all.
func BenchmarkFindThenFirst(b *testing.B) {
	doc := mustParse(`
		<div id="root">
			<div class="x"><span class="t">first</span></div>
			<div class="x"><span class="t">second</span></div>
` + strings.Repeat(`<div><span class="t"></span></div>`, 500) +
		`
		</div>
	`)
	root := doc.Find("#root")
	m := goquery.Single(".t")

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sink = root.FindMatcher(m).First()
	}
}

// TestUtilsRemove tests the utils.Remove function
func TestUtilsRemove(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		check    func(*testing.T, *goquery.Selection)
	}{
		{
			name: "remove direct child ul",
			html: `<div>
				<p>Keep this</p>
				<ul><li>Remove this</li></ul>
				<p>Keep this too</p>
			</div>`,
			selector: "ul",
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 0, sel.Find("ul").Length(), "ul should be removed")
				assert.Equal(t, 2, sel.Find("p").Length(), "paragraphs should remain")
			},
		},
		{
			name: "remove only top-level lists, not nested",
			html: `<div>
				<ul class="outer">
					<li>Outer item
						<ul class="inner"><li>Inner item</li></ul>
					</li>
				</ul>
			</div>`,
			selector: "ul",
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 0, sel.Find("ul").Length(), "all ul should be removed")
			},
		},
		{
			name: "remove multiple selectors",
			html: `<div>
				<ul><li>List</li></ul>
				<ol><li>Ordered</li></ol>
				<dl><dt>Term</dt></dl>
			</div>`,
			selector: "ul, ol, dl",
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 0, sel.Find("ul").Length(), "ul should be removed")
				assert.Equal(t, 0, sel.Find("ol").Length(), "ol should be removed")
				assert.Equal(t, 0, sel.Find("dl").Length(), "dl should be removed")
			},
		},
		{
			name: "remove with class selector",
			html: `<div>
				<p class="remove-me">Remove</p>
				<p class="keep-me">Keep</p>
			</div>`,
			selector: ".remove-me",
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 0, sel.Find(".remove-me").Length(), "class should be removed")
				assert.Equal(t, 1, sel.Find(".keep-me").Length(), "other class should remain")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			utils.Remove(doc.Selection, tt.selector)

			tt.check(t, doc.Selection)
		})
	}
}

// TestUtilsRemoveMultiple tests the utils.RemoveMultiple function
func TestUtilsRemoveMultiple(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		selectors []string
		check     func(*testing.T, *goquery.Selection)
	}{
		{
			name: "remove multiple list types",
			html: `<div>
				<p>Paragraph</p>
				<ul><li>Unordered</li></ul>
				<ol><li>Ordered</li></ol>
				<dl><dt>Definition</dt></dl>
			</div>`,
			selectors: []string{"ul", "ol", "dl"},
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 0, sel.Find("ul").Length(), "ul should be removed")
				assert.Equal(t, 0, sel.Find("ol").Length(), "ol should be removed")
				assert.Equal(t, 0, sel.Find("dl").Length(), "dl should be removed")
				assert.Equal(t, 1, sel.Find("p").Length(), "p should remain")
			},
		},
		{
			name: "remove with class and id selectors",
			html: `<div>
				<p class="remove">Remove 1</p>
				<p id="remove">Remove 2</p>
				<p>Keep</p>
			</div>`,
			selectors: []string{".remove", "#remove"},
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 0, sel.Find(".remove").Length(), "class selector should work")
				assert.Equal(t, 0, sel.Find("#remove").Length(), "id selector should work")
				assert.Equal(t, 1, sel.Find("p").Length(), "one p should remain")
			},
		},
		{
			name: "empty selectors list",
			html: `<div>
				<p>Keep everything</p>
				<ul><li>Keep this too</li></ul>
			</div>`,
			selectors: []string{},
			check: func(t *testing.T, sel *goquery.Selection) {
				assert.Equal(t, 1, sel.Find("p").Length(), "p should remain")
				assert.Equal(t, 1, sel.Find("ul").Length(), "ul should remain")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			utils.Remove(doc.Selection, strings.Join(tt.selectors, ", "))

			tt.check(t, doc.Selection)
		})
	}
}
