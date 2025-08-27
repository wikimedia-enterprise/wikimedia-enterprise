package builder

import (
	"time"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"
)

var (
	void = struct{}{} // zero memory declaration
	// PatrollerRoles behaves as a hash set that holds the roles with patroller privileges.
	PatrollerRoles = map[string]struct{}{
		"rollback":     void,
		"abusefilter":  void,
		"patroller":    void,
		"reviewer":     void,
		"autoreview":   void,
		"autoreviewer": void,
		"editor":       void,
		"autoeditor":   void,
		"eliminator":   void,
	}

	// AdvancedRoles behaves as a hash set that holds the roles with advanced privileges.
	AdvancedRoles = map[string]struct{}{
		"checkuser":  void,
		"bureaucrat": void,
		"oversight":  void,
		"steward":    void,
	}
)

// EditorBuilder follows the builder pattern for the Editor schema
type EditorBuilder struct {
	editor *schema.Editor
}

// NewEditorBuilder returns an editor builder with an empty editor.
func NewEditorBuilder() *EditorBuilder {
	return &EditorBuilder{new(schema.Editor)}
}

// Identifier sets an identifier to the editor schema.
func (eb *EditorBuilder) Identifier(identifier int) *EditorBuilder {
	eb.editor.Identifier = identifier
	return eb
}

// Name sets a name to the editor schema.
func (eb *EditorBuilder) Name(name string) *EditorBuilder {
	eb.editor.Name = name
	return eb
}

// IsAnonymous sets the anonymous flag to the editor schema.
func (eb *EditorBuilder) IsAnonymous(isAnonymous bool) *EditorBuilder {
	eb.editor.IsAnonymous = isAnonymous
	return eb
}

// DateStarted sets the date started to the editor schema.
func (eb *EditorBuilder) DateStarted(dateStarted *time.Time) *EditorBuilder {
	eb.editor.DateStarted = dateStarted
	return eb
}

// EditCount sets the edit count to the editor schema.
func (eb *EditorBuilder) EditCount(editCount int) *EditorBuilder {
	eb.editor.EditCount = editCount
	return eb
}

// Groups sets the groups to the editor schema.
func (eb *EditorBuilder) Groups(groups []string) *EditorBuilder {
	eb.editor.Groups = groups
	return eb
}

// IsBot determines if the editor is a bot based on the provided groups.
func (eb *EditorBuilder) IsBot(isBot bool) *EditorBuilder {
	eb.editor.IsBot = isBot
	return eb
}

// IsAdmin determines if the editor has admin privileges based on the provided groups.
func (eb *EditorBuilder) IsAdmin(groups []string) *EditorBuilder {
	for _, group := range groups {
		if group == "sysop" {
			eb.editor.IsAdmin = true
			break
		}
	}

	return eb
}

// IsPatroller determines if the editor has a patroller role based on the provided groups.
func (eb *EditorBuilder) IsPatroller(groups []string) *EditorBuilder {
	for _, group := range groups {
		if _, ok := PatrollerRoles[group]; ok {
			eb.editor.IsPatroller = true
			break
		}
	}

	return eb
}

// HasAdvancedRights determines if the editor has advanced rights based on the provided groups.
func (eb *EditorBuilder) HasAdvancedRights(groups []string) *EditorBuilder {
	for _, group := range groups {
		if _, ok := AdvancedRoles[group]; ok {
			eb.editor.HasAdvancedRights = true
			break
		}
	}

	return eb
}

// Build returns the builder's editor schema.
func (eb *EditorBuilder) Build() *schema.Editor {
	return eb.editor
}
