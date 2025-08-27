// Package filter contains filters to be used by handlers.
// Such as namespace and project filters.
package filter

import (
	"fmt"
	"slices"
	"wikimedia-enterprise/services/eventstream-listener/submodules/config"
)

// New returns new Filter instance.
func New(cfg config.API) *Filter {
	return &Filter{
		Config: cfg,
	}
}

// Filters used across handlers.
type Filter struct {
	Config config.API
}

// IsSupported returns true if the provided combination of project and namespaces
// can be processed by the WME pipelines, as configured in general/config.
func (flt Filter) IsSupported(dbname string, namespaces ...int) bool {
	// Text-based projects (Language projects)
	if slices.Contains(flt.Config.GetProjects(), dbname) {
		for _, ns := range namespaces {
			if slices.Contains(flt.Config.GetNamespaces(), ns) {
				return true
			}
		}
	}

	// Special projects
	for _, special := range flt.Config.GetSpecials() {
		if dbname == special.Project {
			for _, ns := range namespaces {
				if slices.Contains(special.Namespaces, ns) {
					return true
				}
			}
		}
	}

	return false
}

// ResolveLang returns the language corresponding to the provided project.
// If the provided project is not supported in general/config, it logs
// and returns an error.
func (flt Filter) ResolveLang(dbname string) (string, error) {
	// Text-based projects (Language projects)
	lang := flt.Config.GetLanguage(dbname)

	if len(lang) > 0 {
		return lang, nil
	}

	// Special projects
	for _, special := range flt.Config.GetSpecials() {
		if dbname == special.Project {
			return special.Language, nil
		}
	}

	return "", fmt.Errorf("could not find language for project %s", dbname)
}
