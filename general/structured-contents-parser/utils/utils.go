package utils

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type filteredMatcher struct {
	goquery.Matcher
	predicate func(*html.Node) bool
}

// FilterMatcher returns a matcher equivalent to the provided predicate.
func FilterMatcher(predicate func(*html.Node) bool) goquery.Matcher {
	return &filteredMatcher{predicate: predicate}
}

func (m *filteredMatcher) Match(n *html.Node) bool {
	return m.predicate(n)
}

var nodeNames = []string{
	html.ErrorNode:    "#error",
	html.TextNode:     "#text",
	html.DocumentNode: "#document",
	html.CommentNode:  "#comment",
}

// NodeName returns the node name of the given HTML node.
func NodeName(node *html.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case html.ElementNode, html.DoctypeNode:
		return node.Data
	default:
		if int(node.Type) < len(nodeNames) {
			return nodeNames[node.Type]
		}
		return ""
	}
}

// AttrOr is equivalent to goquery.Selection.AttrOr, except for html.Node.
func AttrOr(n *html.Node, key, def string) string {
	if n == nil {
		return def
	}
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return def
}

// TopOnlyMatcher is equivalent to the embedded matcher, except that it doesn't match nodes
// with parents that also match the embedded matcher. That means it can stop recursing as soon
// as it finds a match.
// Useful specially for removals.
type topOnlyMatcher struct {
	goquery.Matcher
}

func (m topOnlyMatcher) MatchAll(n *html.Node) []*html.Node {
	return m.matchAllInto(n, nil)
}

func (m topOnlyMatcher) matchAllInto(n *html.Node, storage []*html.Node) []*html.Node {
	if m.Match(n) {
		storage = append(storage, n)
		return storage
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		storage = m.matchAllInto(child, storage)
	}

	return storage
}

// TopOnly returns a matcher equivalent to the provided selector, except that it doesn't match nodes
// with parents that also match the provided selector. That means it can stop recursing as soon
// as it finds a match.
// Useful specially for removals.
func TopOnly(sel string) goquery.Matcher {
	return TopOnlyMatcher(cascadia.MustCompile(sel))
}

// TopOnlyMatcher returns a matcher equivalent to the provided matcher, except that it doesn't match nodes
// with parents that also match the provided matcher. That means it can stop recursing as soon
// as it finds a match.
// Useful specially for removals.
func TopOnlyMatcher(m goquery.Matcher) goquery.Matcher {
	return topOnlyMatcher{m}
}

func NewSelection(base *goquery.Selection, nodes ...*html.Node) *goquery.Selection {
	newSel := base.Slice(0, 0)
	// Important: don't call .AddNodes(), the returned selection from Slice shares the underlying array with base.
	newSel.Nodes = nodes
	return newSel
}

// FindFirst returns the first element matched by the provided selector.
// It stops searching as soon as one is found.
func FindFirst(s *goquery.Selection, str string) *goquery.Selection {
	return FindFirstMatcher(s, cascadia.MustCompile(str))
}

// FindFirstMatcher returns the first element matched by the provided matcher.
// It stops searching as soon as one is found.
func FindFirstMatcher(s *goquery.Selection, m goquery.Matcher) *goquery.Selection {
	single := goquery.SingleMatcher(m)
	for _, node := range s.Nodes {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			match := single.MatchAll(c)
			if len(match) > 0 {
				return NewSelection(s, match[0])
			}
		}
	}

	return NewSelection(s)
}

// Remove removes elements matching the selector using TopOnly to avoid affecting parent contexts
func Remove(sel *goquery.Selection, selector string) {
	if selector == "" {
		return
	}
	matcher := cascadia.MustCompile(selector)
	sel.FindMatcher(TopOnlyMatcher(matcher)).Remove()
}

type findIgnoringMatcher struct {
	target goquery.Matcher
	ignore goquery.Matcher
}

func (m findIgnoringMatcher) Match(n *html.Node) bool {
	return m.target.Match(n)
}

func (m findIgnoringMatcher) MatchAll(n *html.Node) []*html.Node {
	return m.matchAllInto(n, nil)
}

func (m findIgnoringMatcher) matchAllInto(n *html.Node, storage []*html.Node) []*html.Node {
	if m.ignore.Match(n) {
		return storage
	}

	if m.target.Match(n) {
		storage = append(storage, n)
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		storage = m.matchAllInto(child, storage)
	}

	return storage
}

func (m findIgnoringMatcher) Filter(ns []*html.Node) []*html.Node {
	result := make([]*html.Node, 0, len(ns))
	for _, n := range ns {
		if m.Match(n) {
			result = append(result, n)
		}
	}
	return result
}

// FindIgnoring returns a matcher that matches the target selector but
// ignores any nodes that are descendants of nodes matching the ignore selector.
func FindIgnoring(target, ignore string) goquery.Matcher {
	return findIgnoringMatcher{
		target: cascadia.MustCompile(target),
		ignore: cascadia.MustCompile(ignore),
	}
}
