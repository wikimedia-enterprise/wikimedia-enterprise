// Package filter contains filters to be used by handlers.
// Such as namespace and project filters.
package filter

// New returns new Filter instance.
func New() (*Filter, error) {
	projects, err := NewProjects()

	if err != nil {
		return nil, err
	}

	return &Filter{
		Projects:   projects,
		Namespaces: NewNamespaces(),
	}, nil
}

// Filter list of filters used across handlers.
type Filter struct {
	Projects   *Projects
	Namespaces *Namespaces
}
