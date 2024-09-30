package common

type PageChangeIdentifier string

const (
	EventIdentifierContextKey PageChangeIdentifier = "event.identifier"
)

var (
	SupportedProjects   = []string{"commonswiki"}
	SupportedNamespaces = []int{6}
)
