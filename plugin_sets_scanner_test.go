package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanDirectory_shouldFailIfDirectoryDoesNotExist(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "no/such/folder"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfPluginsYmlDoesNotExist(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-plugins-yml"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfIdIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-id"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfVersionsIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-versions"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfPluginsAreMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-plugins"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfNoDescriptionYmlExist(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-description-yml"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfNameIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-name"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfFeaturesAreMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-features"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldReturnPluginSets(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/proper-plugin-sets"}

	pluginSets, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.NoError(t, err)
	assert.Len(t, pluginSets, 2)

	pluginSet := findPluginSetById(pluginSets, "plug-and-play")
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.0.0")))
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.29.9")))
	assert.False(t, pluginSet.Versions.Contains(MustParseVersion("1.0.0")))
	assert.False(t, pluginSet.Versions.Contains(MustParseVersion("2.30.0")))

	pluginSet = findPluginSetById(pluginSets, "administration-and-management")
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.0.0")))
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.32.0")))
	assert.False(t, pluginSet.Versions.Contains(MustParseVersion("1.0.0")))
}

func TestScanDirectory_shouldReadAllDescriptionFiles(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/proper-plugin-sets"}

	pluginSets, _ := scanPluginSetsDirectory(configuration.DescriptorDirectory)
	pluginSet := findPluginSetById(pluginSets, "plug-and-play")

	assert.Len(t, pluginSet.Plugins, 3)
	assert.Equal(t, "scm-auth-ldap-plugin", pluginSet.Plugins[0])
	assert.Equal(t, "scm-script-plugin", pluginSet.Plugins[1])
	assert.Equal(t, "scm-editor-plugin", pluginSet.Plugins[2])

	assert.Len(t, pluginSet.Descriptions, 2)

	german := pluginSet.Descriptions["de"]
	assert.Equal(t, "Anklicken und loslegen", german.Name)
	assert.Len(t, german.Features, 3)
	assert.Equal(t, "Empfehlungen der SCM-Manager-Entwickler", german.Features[0])
	assert.Equal(t, "Grundlegende Plugins für kleine und mittlere Teams", german.Features[1])
	assert.Equal(t, "Ermöglichen eine fortschrittliche Suche und weitere kleine Hilfsmittel", german.Features[2])

	english := pluginSet.Descriptions["en"]
	assert.Equal(t, "Plug'n Play", english.Name)
	assert.Len(t, english.Features, 3)
	assert.Equal(t, "recommendations of SCM-Manager-developers", english.Features[0])
	assert.Equal(t, "basic plugins for small to medium sized teams", english.Features[1])
	assert.Equal(t, "enable an advanced search and more neat tricks", english.Features[2])
}

func findPluginSetById(pluginSets []PluginSet, id string) *PluginSet {
	for _, pluginSet := range pluginSets {
		if id == pluginSet.Id {
			return &pluginSet
		}
	}
	return nil
}
