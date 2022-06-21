package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_scanPluginSetsDirectory_shouldFailIfDirectoryDoesNotExist(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "no/such/folder"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not open plugin sets directory")
}

func Test_scanPluginSetsDirectory_shouldFailIfPluginsYmlDoesNotExist(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-plugins-yml"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain plugins.yml")
}

func Test_scanPluginSetsDirectory_shouldFailIfIdIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-id"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is missing at")
}

func Test_scanPluginSetsDirectory_shouldFailIfVersionsIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-versions"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "versions is missing at")
}

func Test_scanPluginSetsDirectory_shouldFailIfSequenceIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-sequence"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sequence is missing or less than one at")
}

func Test_scanPluginSetsDirectory_shouldFailIfSequenceIsLessThanOne(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-sequence-lt-one"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func Test_scanPluginSetsDirectory_shouldFailIfPluginsAreMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-plugins"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func Test_scanPluginSetsDirectory_shouldFailIfNoDescriptionYmlExist(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-description-yml"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain any description_*.yml")
}

func Test_scanPluginSetsDirectory_shouldFailIfNameIsMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-name"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is missing at")
}

func Test_scanPluginSetsDirectory_shouldFailIfFeaturesAreMissing(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/plugin-sets-no-features"}

	_, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.Error(t, err)
}

func Test_scanPluginSetsDirectory_shouldReturnPluginSets(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/proper-plugin-sets"}

	pluginSets, err := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	assert.NoError(t, err)
	assert.Len(t, pluginSets, 2)

	pluginSet := findPluginSetById(pluginSets, "plug-and-play")
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.0.0")))
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.29.9")))
	assert.False(t, pluginSet.Versions.Contains(MustParseVersion("1.0.0")))

	pluginSet = findPluginSetById(pluginSets, "administration")
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.0.0")))
	assert.True(t, pluginSet.Versions.Contains(MustParseVersion("2.32.0")))
	assert.False(t, pluginSet.Versions.Contains(MustParseVersion("1.0.0")))
}

func Test_scanPluginSetsDirectory_shouldReadAllDescriptionFiles(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/proper-plugin-sets"}

	pluginSets, _ := scanPluginSetsDirectory(configuration.DescriptorDirectory)
	pluginSet := findPluginSetById(pluginSets, "plug-and-play")

	assert.Len(t, pluginSet.Plugins, 9)
	assert.Equal(t, "scm-landingpage-plugin", pluginSet.Plugins[0])
	assert.Equal(t, "scm-editor-plugin", pluginSet.Plugins[1])
	assert.Equal(t, "scm-content-search-plugin", pluginSet.Plugins[2])

	assert.Len(t, pluginSet.Descriptions, 2)

	german := pluginSet.Descriptions["de"]
	assert.Equal(t, "Anklicken und loslegen", german.Name)
	assert.Len(t, german.Features, 3)
	assert.Equal(t, "Empfehlungen der SCM-Manager-Entwickler", german.Features[0])
	assert.Equal(t, "Grundlegende Plugins für kleine und mittlere Teams", german.Features[1])
	assert.Equal(t, "Ermöglichen eine fortschrittliche Suche und weitere kleine Hilfsmittel", german.Features[2])

	english := pluginSet.Descriptions["en"]
	assert.Equal(t, "Plug and Play", english.Name)
	assert.Len(t, english.Features, 3)
	assert.Equal(t, "recommendations of SCM-Manager-developers", english.Features[0])
	assert.Equal(t, "basic plugins for small to medium sized teams", english.Features[1])
	assert.Equal(t, "enable an advanced search and more neat tricks", english.Features[2])
}

func Test_scanPluginSetsDirectory_shouldReadAllImageFiles(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugin-sets/proper-plugin-sets"}

	pluginSets, _ := scanPluginSetsDirectory(configuration.DescriptorDirectory)

	pluginSet := findPluginSetById(pluginSets, "administration")
	assert.Len(t, pluginSet.Images, 2)
	assert.NotEmpty(t, pluginSet.Images["check"])
	assert.NotEmpty(t, pluginSet.Images["standard"])
}

func Test_readPluginsYml_shouldFailIfFileDoesNotExist(t *testing.T) {
	yml, err := readPluginsYml("missing.yml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read plugins.yml at")
	assert.Empty(t, yml)
}

func Test_readPluginsYml_shouldFailIfYmlIsBroken(t *testing.T) {
	yml, err := readPluginsYml("resources/test/plugin-sets/broken.yml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
	assert.Empty(t, yml)
}

func Test_readDescriptionsYml_shouldFailIfFileDoesNotExist(t *testing.T) {
	yml, err := readDescriptionsYml("missing.yml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read description_*.yml at")
	assert.Empty(t, yml)
}

func Test_readDescriptionsYml_shouldFailIfYmlIsBroken(t *testing.T) {
	yml, err := readDescriptionsYml("resources/test/plugin-sets/broken.yml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal description_*.yml at")
	assert.Empty(t, yml)
}

func Test_appendImages_shouldReturnNilIfNoImageExists(t *testing.T) {
	var images map[string]string

	err := appendImages(images, "resources/test/plugin-sets/plugin-set-no-images/plugin-set")

	assert.NoError(t, err)
	assert.Empty(t, images)
}

func findPluginSetById(pluginSets []PluginSet, id string) *PluginSet {
	for _, pluginSet := range pluginSets {
		if id == pluginSet.Id {
			return &pluginSet
		}
	}
	return nil
}
