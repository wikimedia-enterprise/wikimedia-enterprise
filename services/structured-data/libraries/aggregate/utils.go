package aggregate

import (
	"strings"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"
)

// IsNotFatalError checks whether an error is a known non-fatal error and returns true if it is.
func IsNonFatalErr(err error) bool {
	if err == nil {
		return false
	}

	// This error indicates that a requested page was not found in a WMF (Wikimedia Foundation) API response.
	if strings.Contains(err.Error(), wmf.ErrPageNotFound.Error()) {
		return true
	}

	// This error message indicates that a requested resource was not found on the server.
	if strings.Contains(err.Error(), "404 Not Found") {
		return true
	}

	// This error message indicates that the we do not have permission to access the requested resource.
	if strings.Contains(err.Error(), "403 Forbidden") {
		return true
	}

	// If the error is not one of the known non-fatal errors, return false.
	// This indicates that the error is not a known non-fatal error and may need to be handled differently.
	return false
}
