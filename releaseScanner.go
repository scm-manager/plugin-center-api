package main

import (
	"bytes"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

func scanDirectory(directory string) ([]Plugin, error) {

	var plugins []Plugin

	pluginDirectories, err := ioutil.ReadDir(directory)

	if err != nil {
		return nil, errors.Wrap(err, "could not open plugin directory "+directory)
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
	log.Println("reading plugin file", indexFileName)
	indexContent, err := ioutil.ReadFile(indexFileName)
	if err != nil {
		log.Println("no index.md file found in directory", pluginDirectory)
		return nil
	}
	var plugin Plugin
	err = unmarshalFrontMatter(indexContent, &plugin)
	if err == nil {
		releases := readReleases(filepath.Join(pluginDirectory, "releases"))
		plugin.Releases = releases
		return &plugin
	} else {
		log.Fatalln("could not read  plugin directory", pluginDirectory)
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
			releaseFilePath := filepath.Join(releaseDirectory, releaseFile.Name())
			log.Println("reading release file", releaseFilePath)
			release, err := readRelease(releaseFilePath)
			if err != nil {
				log.Fatalln("could not read release file", releaseFilePath)
			}
			releases = append(releases, release)
		}
	}
	sort.Slice(releases, less(releases))
	return releases
}

func readRelease(releaseFileName string) (Release, error) {
	releaseYaml, err := ioutil.ReadFile(releaseFileName)
	if err != nil {
		return Release{}, errors.Wrap(err, "could not read release file "+releaseFileName)
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
		return errors.Wrap(err, "could not parse front matter content")
	}

	return nil
}
