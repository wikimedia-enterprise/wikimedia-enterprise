// Package enforcer is responsible for dependency injection of Casbin enforcer.
package enforcer

import (
	"wikimedia-enterprise/api/main/config/env"

	"github.com/casbin/casbin/v2"
)

// New creates new instance of Casbin Enforcer.
func New(env *env.Environment) (*casbin.Enforcer, error) {
	return casbin.NewEnforcer(env.AccessModel.Path, env.AccessPolicy.Path)
}
