package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestReadConfigurationFromConfigYaml(t *testing.T) {
	config := readConfiguration()
	assert.Equal(t, "resources/test/plugins", config.DescriptorDirectory)
	assert.False(t, config.Oidc.IsEnabled())
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
	assert.True(t, config.Oidc.IsEnabled())
}

func TestReadConfigurationFromNonDefaultPath(t *testing.T) {
	t.Setenv("CONFIG", "resources/test/oidc/config.yaml")

	config := readConfiguration()
	assert.Equal(t, "/plugins", config.DescriptorDirectory)
	assert.Equal(t, "http://localhost:8080/auth/realms/master", config.Oidc.Issuer)
	assert.Equal(t, "plugin-center", config.Oidc.ClientID)
	assert.Equal(t, "secret", config.Oidc.ClientSecret)
	assert.Equal(t, "http://localhost:8080/api/v1/auth/oidc/callback", config.Oidc.RedirectURL)
}

func TestReadConfigurationWithoutConfigYaml(t *testing.T) {
	workDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(workDir)

	err = os.Chdir(os.TempDir())
	assert.NoError(t, err)

	t.Setenv("CONFIG_DESCRIPTOR_DIRECTORY", "/plugins")

	config := readConfiguration()
	assert.Equal(t, "/plugins", config.DescriptorDirectory)
}

func TestReadConfigurationFromConfigYamlAndEnvironment(t *testing.T) {
	t.Setenv("CONFIG_PORT", "8082")

	config := readConfiguration()
	assert.Equal(t, "resources/test/plugins", config.DescriptorDirectory)
	assert.Equal(t, 8082, config.Port)
}

func TestReadConfigurationAndUseDefaults(t *testing.T) {
	config := readConfiguration()
	assert.Equal(t, 8000, config.Port)
}
