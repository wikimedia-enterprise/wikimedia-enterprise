package filter

import "wikimedia-enterprise/general/schema"

// NewNamespaces returns supported namespaces filter.
func NewNamespaces() *Namespaces {
	return &Namespaces{
		map[int]struct{}{
			schema.NamespaceArticle:  {},
			schema.NamespaceFile:     {},
			schema.NamespaceCategory: {},
			schema.NamespaceTemplate: {},
		},
	}
}

// Namespaces supported namespaces filter.
type Namespaces struct {
	supported map[int]struct{}
}

// IsSupported check if the namespace is supported by our system.
func (n *Namespaces) IsSupported(ns int) bool {
	_, ok := n.supported[ns]
	return ok
}
