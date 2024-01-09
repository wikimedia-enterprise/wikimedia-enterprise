package cognito_test

import (
	"testing"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/libraries/cognito"

	"github.com/stretchr/testify/assert"
)

func TestCognito(t *testing.T) {
	assert.NotNil(t, cognito.New(new(env.Environment)))
}
