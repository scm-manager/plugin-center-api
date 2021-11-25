package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestReadConfigurationFromConfigYaml(t *testing.T) {
	config := readConfiguration()
	assert.Equal(t, "resources/test/plugins", config.DescriptorDirectory)
	assert.Equal(t, "http://localhost:8080/auth/realms/master", config.Oidc.Issuer)
	assert.Equal(t, "plugin-center", config.Oidc.ClientID)
	assert.Equal(t, "secret", config.Oidc.ClientSecret)
	assert.Equal(t, "http://localhost:8080/api/v1/auth/oidc/callback", config.Oidc.RedirectURL)
}

func TestReadConfigurationFromEnv(t *testing.T) {
	t.Setenv("CONFIG_DESCRIPTOR_DIRECTORY", "/plugins")
	t.Setenv("CONFIG_OIDC_ISSUER", "http://keycloak:8000")
	t.Setenv("CONFIG_OIDC_CLIENT_ID", "pc")
	t.Setenv("CONFIG_OIDC_CLIENT_SECRET", "secret123")
	t.Setenv("CONFIG_OIDC_REDIRECT_URL", "https://lo:3000/cb")

	config := readConfiguration()

	assert.Equal(t, "/plugins", config.DescriptorDirectory)
	assert.Equal(t, "http://keycloak:8000", config.Oidc.Issuer)
	assert.Equal(t, "pc", config.Oidc.ClientID)
	assert.Equal(t, "secret123", config.Oidc.ClientSecret)
	assert.Equal(t, "https://lo:3000/cb", config.Oidc.RedirectURL)
}

func TestReadConfigurationFromNonDefaultPath(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "test-")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	text := []byte("descriptor-directory: /plugins")
	_, err = tmpFile.Write(text)
	assert.NoError(t, err)

	err = os.Setenv("CONFIG", tmpFile.Name())
	assert.NoError(t, err)
	defer os.Unsetenv("CONFIG")

	config := readConfiguration()
	assert.Equal(t, "/plugins", config.DescriptorDirectory)
}

func TestReadConfigurationWithoutConfigYaml(t *testing.T) {
	workDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(workDir)

	err = os.Chdir(os.TempDir())
	assert.NoError(t, err)

	err = os.Setenv("CONFIG_DESCRIPTOR_DIRECTORY", "/plugins")
	assert.NoError(t, err)
	defer os.Unsetenv("CONFIG_DESCRIPTOR_DIRECTORY")

	config := readConfiguration()
	assert.Equal(t, "/plugins", config.DescriptorDirectory)
}

func TestReadConfigurationFromConfigYamlAndEnvironment(t *testing.T) {
	err := os.Setenv("CONFIG_PORT", "8082")
	assert.NoError(t, err)
	defer os.Unsetenv("CONFIG_PORT")

	config := readConfiguration()
	assert.Equal(t, "resources/test/plugins", config.DescriptorDirectory)
	assert.Equal(t, 8082, config.Port)
}

func TestReadConfigurationAndUseDefaults(t *testing.T) {
	config := readConfiguration()
	assert.Equal(t, 8000, config.Port)
}
