package main

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func scanPluginSetsDirectory(directory string) ([]PluginSet, error) {
	var pluginSets []PluginSet

	pluginSetsDirectory, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open plugin sets directory %s", directory)
	}

	for _, pluginSetDirectory := range pluginSetsDirectory {
		pluginSet, err := readPluginSetDirectory(filepath.Join(directory, pluginSetDirectory.Name()))
		if err != nil || pluginSet == nil {
			return pluginSets, err
		}
		pluginSets = append(pluginSets, *pluginSet)
	}

	return pluginSets, nil
}

func readPluginSetDirectory(pluginSetDirectory string) (*PluginSet, error) {
	log.Println("reading plugin set from", pluginSetDirectory)

	pluginSetYml := filepath.Join(pluginSetDirectory, "plugins.yml")
	if _, err := os.Stat(pluginSetYml); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "directory %s does not contain plugins.yml", pluginSetDirectory)
	}

	descriptionYmls, err := filepath.Glob(filepath.Join(pluginSetDirectory, "description_*.yml"))
	if err != nil || descriptionYmls == nil || len(descriptionYmls) == 0 {
		return nil, errors.New(fmt.Sprintf("directory %s does not contain any description_*.yml", pluginSetDirectory))
	}

	plugins, err := readPluginsYml(pluginSetYml)
	if err != nil {
		return nil, err
	}

	pluginSet := PluginSet{
		Id:           plugins.Id,
		Versions:     plugins.Versions,
		Sequence:     plugins.Sequence,
		Plugins:      plugins.Plugins,
		Descriptions: make(map[string]Description),
		Images:       make(map[string]string),
	}

	for _, descriptionYml := range descriptionYmls {
		description, err := readDescriptionsYml(descriptionYml)
		if err != nil {
			return nil, err
		}
		lang := filepath.Base(descriptionYml)[12:14]
		pluginSet.Descriptions[lang] = description
	}

	if err = appendImages(pluginSet.Images, pluginSetDirectory); err != nil {
		return nil, err
	}

	return &pluginSet, nil
}

func readPluginsYml(pluginsYmlPath string) (Plugins, error) {
	pluginsYml, err := ioutil.ReadFile(pluginsYmlPath)
	if err != nil {
		return Plugins{}, errors.Wrapf(err, "failed to read pugins.yml at %s", pluginsYmlPath)
	}
	var plugins Plugins
	if err = yaml.Unmarshal(pluginsYml, &plugins); err != nil {
		return Plugins{}, errors.Wrapf(err, "failed to unmarshal plugins.yml at %s", pluginsYmlPath)
	}
	if plugins.Id == "" {
		return Plugins{}, errors.New(fmt.Sprintf("id is missing at %s", pluginsYmlPath))
	}
	if plugins.Versions.Value == "" {
		return Plugins{}, errors.New(fmt.Sprintf("versions is missing at %s", pluginsYmlPath))
	}
	if plugins.Sequence < 1 {
		return Plugins{}, errors.New(fmt.Sprintf("sequence is missing or less than one at %s", pluginsYmlPath))
	}
	if len(plugins.Plugins) == 0 {
		return Plugins{}, errors.New(fmt.Sprintf("plugins are missing at %s", pluginsYmlPath))
	}
	return plugins, nil
}

func readDescriptionsYml(descriptionsYmlPath string) (Description, error) {
	descriptionYml, err := ioutil.ReadFile(descriptionsYmlPath)
	if err != nil {
		return Description{}, errors.Wrapf(err, "failed to read description_*.yml at %s", descriptionsYmlPath)
	}
	var description Description
	if err = yaml.Unmarshal(descriptionYml, &description); err != nil {
		return Description{}, errors.Wrapf(err, "failed to unmarshal description_*.yml at %s", descriptionsYmlPath)
	}
	if description.Name == "" {
		return Description{}, errors.New(fmt.Sprintf("name is missing at %s", descriptionsYmlPath))
	}
	if len(description.Features) == 0 {
		return Description{}, errors.New(fmt.Sprintf("features are missing at %s", descriptionsYmlPath))
	}
	return description, nil
}

func appendImages(images map[string]string, pluginSetDirectory string) error {
	imagePaths, err := filepath.Glob(filepath.Join(pluginSetDirectory, "*.svg"))
	if err != nil || imagePaths == nil || len(imagePaths) == 0 {
		log.Println("no images found at", pluginSetDirectory)
		return nil
	}
	for _, imagePath := range imagePaths {
		image, err := ioutil.ReadFile(imagePath)
		if err != nil {
			return errors.Wrapf(err, "failed to read image at %s", imagePath)
		}
		key := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))
		images[key] = base64.StdEncoding.EncodeToString(image)
	}
	return nil
}
