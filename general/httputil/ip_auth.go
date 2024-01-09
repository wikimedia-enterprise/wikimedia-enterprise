package httputil

import (
	"bytes"
	"encoding/json"
	"net"

	"github.com/gin-gonic/gin"
)

// IPRange structure represents an IP range with a start and an end IP address.
type IPRange struct {
	Start net.IP `json:"start"`
	End   net.IP `json:"end"`
}

// IPUser contains an IP range and the user that corresponds to it.
type IPUser struct {
	IPRange *IPRange `json:"ip_range"`
	User    *User    `json:"user"`
}

// IPAllowList is a slice of IPUser pointers and represents a whitelist of IP ranges and corresponding users.
type IPAllowList []*IPUser

// UnmarshalEnvironmentValue unmarshals the whitelisted IP ranges and users which correspond to them from an environment variable.
// If the IPAllowList pointer is nil, it initializes it to an empty slice of IPUser pointers.
func (i *IPAllowList) UnmarshalEnvironmentValue(dta string) error {
	if i == nil {
		*i = []*IPUser{}
	}

	return json.Unmarshal([]byte(dta), i)
}

// GetUser returns the user corresponding to the given IP address if it falls within any of the whitelisted IP ranges, or nil if it doesn't.
func (i *IPAllowList) GetUser(ip string) *User {
	nip := net.ParseIP(ip)

	for _, ipu := range *i {
		if bytes.Compare(nip, ipu.IPRange.Start) >= 0 && bytes.Compare(nip, ipu.IPRange.End) <= 0 {
			return ipu.User
		}
	}

	return nil
}

// IPAuth is a middleware function that puts a user into the context if the client IP is in the whitelist.
func IPAuth(iwl IPAllowList) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		if iwl == nil {
			gcx.Next()
			return
		}

		if usr := iwl.GetUser(gcx.ClientIP()); usr != nil {
			gcx.Set("user", usr)
		}

		gcx.Next()
	}
}
