package schema

import (
	"time"
)

// Commons schema for commonswiki page.
// Tries to be compliant with articles schema.
type Commons struct {
	Name         string     `json:"name,omitempty"`
	Identifier   int        `json:"identifier,omitempty"`
	DateModified *time.Time `json:"date_modified,omitempty"`
	URL          string     `json:"url,omitempty"`
	Namespace    *Namespace `json:"namespace,omitempty"`
	InLanguage   *Language  `json:"in_language,omitempty"`
	IsPartOf     *Project   `json:"is_part_of,omitempty"`
	Version      *Version   `json:"version,omitempty"`
	Event        *Event     `json:"event,omitempty"`
	Files        []*File    `json:"files,omitempty"`
}
