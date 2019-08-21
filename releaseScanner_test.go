package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFailureForNotExistingPluginsFolder(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "no/such/folder"}

	_, err := scanDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err, "expected error missing", err)
}

func TestIfReleaseFilesAreRead(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugins"}

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	assert.Nil(t, err, "unexpected error reading directory", err)
	assert.Equal(t, 3, len(plugins), "wrong number of plugins")
}

func TestIfPluginMetadataIsRead(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugins"}

	plugins, _ := scanDirectory(configuration.DescriptorDirectory)
	plugin := findPluginByName(plugins, "scm-auth-ldap-plugin")

	assert.Equal(t, "scm-auth-ldap-plugin", plugin.name)
	assert.Equal(t, "LDAP", plugin.displayName)
	assert.Equal(t, "LDAP Authentication", plugin.description)
	assert.Equal(t, "authentication", plugin.category)
	assert.Equal(t, "Cloudogu GmbH", plugin.author)
}

func findPluginByName(plugins []Plugin, name string) *Plugin {
	for _, plugin := range plugins {
		if name == plugin.name {
			return &plugin
		}
	}
	return nil
}
