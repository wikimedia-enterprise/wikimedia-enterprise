package httputil

import (
	"sort"

	"github.com/gin-gonic/gin"
)

// NewUser returns a pointer to a User struct containing user information extracted from the Gin context.
// It first attempts to retrieve the User struct from the context using the "user" key.
// If the User struct is found, it is returned.
// Otherwise, nil is returned.
func NewUser(gcx *gin.Context) *User {
	if umd, ok := gcx.Get("user"); ok {
		switch usr := umd.(type) {
		case *User:
			return usr
		}
	}

	return nil
}

// User represents a user and their associated groups.
type User struct {
	Username string   `json:"username,omitempty"`
	Groups   []string `json:"groups,omitempty"`
}

// GetUsername returns the user's username.
// If the User struct is nil, an empty string is returned.
func (u *User) GetUsername() string {
	if u != nil {
		return u.Username
	}

	return ""
}

// SetUsername sets the user's username.
func (u *User) SetUsername(unm string) {
	u.Username = unm
}

// GetGroups returns a list of groups the user belongs to.
// If the User struct is nil, an empty list is returned.
func (u *User) GetGroups() []string {
	if u != nil {
		return u.Groups
	}

	return []string{}
}

// GetGroup returns a highest group of the user.
// This is being used for segmentation purposes only.
func (u *User) GetGroup() string {
	if u == nil {
		return ""
	}

	if len(u.Groups) <= 0 {
		return ""
	}

	sort.Strings(u.Groups)

	return u.Groups[len(u.Groups)-1]
}

// SetGroups sets the list of groups the user belongs to.
func (u *User) SetGroups(gps []string) {
	u.Groups = gps
}

// IsInGroup checks whether the user belongs to any of the specified groups.
// If the User struct is nil, false is returned.
// If the user belongs to any of the specified groups, true is returned.
func (u *User) IsInGroup(gps ...string) bool {
	if u != nil {
		for _, lkp := range gps {
			for _, grp := range u.Groups {
				if lkp == grp {
					return true
				}
			}
		}
	}

	return false
}
