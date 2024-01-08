// Package abstract is being used to contain helper functions for article
// abstract processing.
package abstract

import "wikimedia-enterprise/general/schema"

// IsValid function that checks if abstract can be extracted from
// current article. By checking namespace and content model.
func IsValid(nsp int, cml string) bool {
	return nsp == schema.NamespaceArticle && cml == "wikitext"
}
