package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnforceToCurrentRoot(t *testing.T) {
	t.Parallel()
	cleanDir, err := EnforceToCurrentRoot("/home/cd/helmapp/", "/home/cd/helmapp/values.yaml")
	require.NoError(t, err)
	assert.Equal(t, "/home/cd/helmapp/values.yaml", cleanDir)

	// File is outside current working directory
	_, err = EnforceToCurrentRoot("/home/cd/helmapp/", "/home/values.yaml")
	require.Error(t, err)

	// File is outside current working directory
	_, err = EnforceToCurrentRoot("/home/cd/helmapp/", "/home/cd/helmapp/../differentapp/values.yaml")
	require.Error(t, err)

	// Goes back and forth, but still legal
	cleanDir, err = EnforceToCurrentRoot("/home/cd/helmapp/", "/home/cd/helmapp/../../cd/helmapp/values.yaml")
	require.NoError(t, err)
	assert.Equal(t, "/home/cd/helmapp/values.yaml", cleanDir)
}
