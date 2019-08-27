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
}

func TestReadConfigurationFromEnv(t *testing.T) {
	err := os.Setenv("CONFIG_DESCRIPTOR_DIRECTORY", "/plugins")
	assert.NoError(t, err)
	config := readConfiguration()
	assert.Equal(t, "/plugins", config.DescriptorDirectory)
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
	config := readConfiguration()
	assert.Equal(t, "/plugins", config.DescriptorDirectory)
}

func TestReadConfigurationFromConfigYamlAndEnvironment(t *testing.T) {
	err := os.Setenv("CONFIG_PORT", "8082")
	assert.NoError(t, err)
	config := readConfiguration()
	assert.Equal(t, "resources/test/plugins", config.DescriptorDirectory)
	assert.Equal(t, 8082, config.Port)
}

func TestReadConfigurationAndUseDefaults(t *testing.T) {
	config := readConfiguration()
	assert.Equal(t, 8000, config.Port)
}
