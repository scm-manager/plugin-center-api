package main

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

func scanDirectory(directory string) ([]Plugin, error) {

	var plugins []Plugin

	pluginDirectories, err := ioutil.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	for _, pluginDirectory := range pluginDirectories {
		plugin := readPluginDirectory(filepath.Join(directory, pluginDirectory.Name()))
		if plugin != nil {
			plugins = append(plugins, *plugin)
		}
	}

	return plugins, nil
}

func readPluginDirectory(pluginDirectory string) *Plugin {
	indexFileName := filepath.Join(pluginDirectory, "index.md")
	indexContent, err := ioutil.ReadFile(indexFileName)
	if err != nil {
		return nil
	}
	var plugin Plugin
	err = unmarshalFrontMatter(indexContent, &plugin)
	if err == nil {
		return &plugin
	} else {
		return nil
	}
}

func unmarshalFrontMatter(b []byte, plugin *Plugin) error {
	var frontMatterDelimiter = []byte("---")

	if !bytes.HasPrefix(b, frontMatterDelimiter) {
		return errors.New("index.md file has no front matter part")
	}

	parts := bytes.SplitN(b, frontMatterDelimiter, 3)

	dataMap := map[string]string{}
	err := yaml.Unmarshal(parts[1], &dataMap)
	if err != nil {
		return err
	}

	(*plugin).name = dataMap["name"]
	(*plugin).displayName = dataMap["displayName"]
	(*plugin).description = dataMap["description"]
	(*plugin).category = dataMap["category"]
	(*plugin).author = dataMap["author"]

	return nil
}
