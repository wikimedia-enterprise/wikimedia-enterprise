package schema

import "time"

// File schema for a mediawiki file.
type File struct {
	Sha1         string     `json:"sha1,omitempty"`
	Size         *Size      `json:"size,omitempty" `
	Width        int        `json:"width,omitempty"`
	Height       int        `json:"height,omitempty"`
	Mime         string     `json:"mime,omitempty"`
	DateModified *time.Time `json:"date_modified,omitempty"`
	Editor       *Editor    `json:"editor,omitempty"`
	License      []*License `json:"license,omitempty"`
}
