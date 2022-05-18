package plugin_sets

import (
	"github.com/scm-manager/plugin-center-api/pkg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanDirectory_shouldFailIfDirectoryDoesNotExist(t *testing.T) {
	configuration := pkg.Configuration{DescriptorDirectory: "no/such/folder"}

	_, err := ScanDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfPluginsYmlDoesNotExist(t *testing.T) {
	configuration := pkg.Configuration{DescriptorDirectory: "testdata/plugin-sets-no-plugins-yml"}

	_, err := ScanDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldFailIfNoDescriptionYmlExist(t *testing.T) {
	configuration := pkg.Configuration{DescriptorDirectory: "testdata/plugin-sets-no-description-yml"}

	_, err := ScanDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func TestScanDirectory_shouldReturnPluginSets(t *testing.T) {
	configuration := pkg.Configuration{DescriptorDirectory: "testdata/plugin-sets"}

	pluginSets, err := ScanDirectory(configuration.DescriptorDirectory)

	assert.NoError(t, err)
	assert.Len(t, pluginSets, 2)
}

func TestScanDirectory_shouldReadAllDescriptionFiles(t *testing.T) {
	configuration := pkg.Configuration{DescriptorDirectory: "testdata/plugin-sets"}

	pluginSets, _ := ScanDirectory(configuration.DescriptorDirectory)
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
