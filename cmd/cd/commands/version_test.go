package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cdclient "github.com/hanzoai/cd/pkg/apiclient"
	"github.com/hanzoai/cd/pkg/apiclient/version"
)

func TestShortVersionClient(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewVersionCmd(&cdclient.ClientOptions{}, nil)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version", "--short", "--client"})
	require.NoError(t, cmd.Execute(), "Failed to execute short version command")
	assert.Equal(t, "cd: v99.99.99+unknown\n", buf.String())
}

func TestShortVersion(t *testing.T) {
	serverVersion := &version.VersionMessage{Version: "v99.99.99+unknown"}
	buf := new(bytes.Buffer)
	cmd := NewVersionCmd(&cdclient.ClientOptions{}, serverVersion)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cd", "version", "--short"})
	require.NoError(t, cmd.Execute(), "Failed to execute short version command")
	assert.Equal(t, "cd: v99.99.99+unknown\ncd-server: v99.99.99+unknown\n", buf.String())
}
