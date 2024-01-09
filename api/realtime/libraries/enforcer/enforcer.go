package enforcer

import (
	"wikimedia-enterprise/api/realtime/config/env"

	"github.com/casbin/casbin/v2"
)

// New create new casbin enforcer instance.
func New(env *env.Environment) (*casbin.Enforcer, error) {
	return casbin.NewEnforcer(env.AccessModel.Path, env.AccessPolicy.Path)
}
