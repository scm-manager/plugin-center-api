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
	data, err := unmarshalFrontMatter(indexContent)
	if err == nil {
		return &Plugin{
			name:        data.Name,
			displayName: data.DisplayName,
			description: data.Description,
			category:    data.Category,
			author:      data.Author,
		}
	} else {
		return nil
	}
}

type pluginFrontMatter struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"displayName"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
	Author      string `yaml:"author"`
}

func unmarshalFrontMatter(b []byte) (pluginFrontMatter, error) {
	var frontMatterDelimiter = []byte("---")
	var data pluginFrontMatter

	if !bytes.HasPrefix(b, frontMatterDelimiter) {
		return data, errors.New("index.md file has no front matter part")
	}

	parts := bytes.SplitN(b, frontMatterDelimiter, 3)
	err := yaml.Unmarshal(parts[1], &data)
	return data, err
}
