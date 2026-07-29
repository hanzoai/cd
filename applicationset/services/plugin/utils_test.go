package plugin

import (
	"github.com/hanzoai/deploy/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSecretKey(t *testing.T) {
	t.Parallel()
	secretName, tokenKey := ParseSecretKey("#my-secret:my-token")
	assert.Equal(t, "my-secret", secretName)
	assert.Equal(t, "$my-token", tokenKey)

	secretName, tokenKey = ParseSecretKey("#my-secret")
	assert.Equal(t, common.ArgoCDSecretName, secretName)
	assert.Equal(t, "#my-secret", tokenKey)
}
