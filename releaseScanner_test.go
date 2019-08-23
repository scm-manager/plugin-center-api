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
	assert.Len(t, plugins, 3, "wrong number of plugins")
}

func TestIfPluginMetadataIsRead(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugins"}

	plugins, _ := scanDirectory(configuration.DescriptorDirectory)
	plugin := findPluginByName(plugins, "scm-auth-ldap-plugin")

	assert.Equal(t, "scm-auth-ldap-plugin", plugin.Name)
	assert.Equal(t, "LDAP", plugin.DisplayName)
	assert.Equal(t, "LDAP Authentication", plugin.Description)
	assert.Equal(t, "authentication", plugin.Category)
	assert.Equal(t, "Cloudogu GmbH", plugin.Author)
}

func TestIfReleasesAreRead(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugins"}

	plugins, _ := scanDirectory(configuration.DescriptorDirectory)
	plugin := findPluginByName(plugins, "scm-cas-plugin")

	assert.Equal(t, 2, len(plugin.Releases), "wrong number of releases")
	release := plugin.Releases[1]
	assert.Equal(t, "1.0.0", release.Version, "wrong version for release")
	assert.Equal(t, "2009-01-01T12:00:00+01:00", release.Date, "wrong date for release")
	assert.Equal(t, "https://download.scm-manager.org/plugins/1.0.0/scm-cas-plugin.smp", release.Url, "wrong url for release")
	assert.Equal(t, "f464372baf1ce0d7d0f67e5283f7c4210e24dcf330f955a3261317a77330c19f", release.Checksum, "wrong checksum for release")
	assert.Equal(t, "linux", release.Conditions.Os, "wrong os for release conditions")
	assert.Equal(t, "amd64", release.Conditions.Arch, "wrong arch for release conditions")
}

func TestReleasesAreSorted(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugins"}

	plugins, _ := scanDirectory(configuration.DescriptorDirectory)
	plugin := findPluginByName(plugins, "scm-script-plugin")

	assert.Equal(t, 3, len(plugin.Releases), "wrong number of releases")
	assert.Equal(t, "1.0.10", plugin.Releases[0].Version, "wrong version for release")
	assert.Equal(t, "1.0.1", plugin.Releases[1].Version, "wrong version for release")
	assert.Equal(t, "1.0.0", plugin.Releases[2].Version, "wrong version for release")
}

func findPluginByName(plugins []Plugin, name string) *Plugin {
	for _, plugin := range plugins {
		if name == plugin.Name {
			return &plugin
		}
	}
	return nil
}
