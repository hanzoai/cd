package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanzoai/deploy/util/localconfig"
)

const testConfig = `contexts:
- name: cd1.example.com:443
  server: cd1.example.com:443
  user: cd1.example.com:443
- name: cd2.example.com:443
  server: cd2.example.com:443
  user: cd2.example.com:443
- name: localhost:8080
  server: localhost:8080
  user: localhost:8080
current-context: localhost:8080
servers:
- server: cd1.example.com:443
- server: cd2.example.com:443
- plain-text: true
  server: localhost:8080
users:
- auth-token: vErrYS3c3tReFRe$hToken
  name: cd1.example.com:443
  refresh-token: vErrYS3c3tReFRe$hToken
- auth-token: vErrYS3c3tReFRe$hToken
  name: cd2.example.com:443
  refresh-token: vErrYS3c3tReFRe$hToken
- auth-token: vErrYS3c3tReFRe$hToken
  name: localhost:8080`

const testConfigFilePath = "./testdata/local.config"

func TestContextDelete(t *testing.T) {
	// Write the test config file
	err := os.WriteFile(testConfigFilePath, []byte(testConfig), os.ModePerm)
	require.NoError(t, err)
	defer os.Remove(testConfigFilePath)

	err = os.Chmod(testConfigFilePath, 0o600)
	require.NoError(t, err, "Could not change the file permission to 0600 %v", err)
	localConfig, err := localconfig.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", localConfig.CurrentContext)
	assert.Contains(t, localConfig.Contexts, localconfig.ContextRef{Name: "localhost:8080", Server: "localhost:8080", User: "localhost:8080"})

	// Delete a non-current context
	err = deleteContext("cd1.example.com:443", testConfigFilePath)
	require.NoError(t, err)

	localConfig, err = localconfig.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", localConfig.CurrentContext)
	assert.NotContains(t, localConfig.Contexts, localconfig.ContextRef{Name: "cd1.example.com:443", Server: "cd1.example.com:443", User: "cd1.example.com:443"})
	assert.NotContains(t, localConfig.Servers, localconfig.Server{Server: "cd1.example.com:443"})
	assert.NotContains(t, localConfig.Users, localconfig.User{AuthToken: "vErrYS3c3tReFRe$hToken", Name: "cd1.example.com:443"})
	assert.Contains(t, localConfig.Contexts, localconfig.ContextRef{Name: "cd2.example.com:443", Server: "cd2.example.com:443", User: "cd2.example.com:443"})
	assert.Contains(t, localConfig.Contexts, localconfig.ContextRef{Name: "localhost:8080", Server: "localhost:8080", User: "localhost:8080"})

	// Delete the current context
	err = deleteContext("localhost:8080", testConfigFilePath)
	require.NoError(t, err)

	localConfig, err = localconfig.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
	assert.Empty(t, localConfig.CurrentContext)
	assert.NotContains(t, localConfig.Contexts, localconfig.ContextRef{Name: "localhost:8080", Server: "localhost:8080", User: "localhost:8080"})
	assert.NotContains(t, localConfig.Servers, localconfig.Server{PlainText: true, Server: "localhost:8080"})
	assert.NotContains(t, localConfig.Users, localconfig.User{AuthToken: "vErrYS3c3tReFRe$hToken", Name: "localhost:8080"})
	assert.Contains(t, localConfig.Contexts, localconfig.ContextRef{Name: "cd2.example.com:443", Server: "cd2.example.com:443", User: "cd2.example.com:443"})
}
