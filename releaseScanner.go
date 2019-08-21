package main

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
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
		releases := readReleases(filepath.Join(pluginDirectory, "releases"))
		plugin.Releases = releases
		return &plugin
	} else {
		return nil
	}
}

func readReleases(releaseDirectory string) []Release {
	var releases []Release
	releaseFiles, err := ioutil.ReadDir(releaseDirectory)
	if err != nil {
		return releases
	}

	for _, releaseFile := range releaseFiles {
		if strings.HasSuffix(releaseFile.Name(), ".yaml") || strings.HasSuffix(releaseFile.Name(), ".yml") {
			release, err := readRelease(filepath.Join(releaseDirectory, releaseFile.Name()))
			if err != nil {
				panic("could not read release file")
			}
			releases = append(releases, release)
		}
	}
	return releases
}

func readRelease(releaseFileName string) (Release, error) {
	releaseYaml, err := ioutil.ReadFile(releaseFileName)
	if err != nil {
		return Release{}, err
	}
	var release Release
	err = yaml.Unmarshal(releaseYaml, &release)
	return release, nil
}

func unmarshalFrontMatter(b []byte, plugin *Plugin) error {
	var frontMatterDelimiter = []byte("---")

	if !bytes.HasPrefix(b, frontMatterDelimiter) {
		return errors.New("index.md file has no front matter part")
	}

	parts := bytes.SplitN(b, frontMatterDelimiter, 3)

	err := yaml.Unmarshal(parts[1], plugin)
	if err != nil {
		return err
	}

	return nil
}
