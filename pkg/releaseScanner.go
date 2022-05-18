package pkg

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ScanDirectory(directory string) ([]Plugin, error) {

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
	pluginYml := filepath.Join(pluginDirectory, "plugin.yml")
	if _, err := os.Stat(pluginYml); os.IsNotExist(err) {
		log.Printf("directory %s does not contain a plugin.yml", pluginDirectory)
		return nil
	}
	plugin, err := readPluginYml(pluginYml)
	if err == nil {
		releases := readReleases(filepath.Join(pluginDirectory, "releases"))
		plugin.Releases = releases
		return &plugin
	} else {
		log.Fatalln("could not read plugin directory", pluginDirectory)
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

	sort.SliceStable(releases, func(i1 int, i2 int) bool { return less(releases)(i2, i1) })

	return releases
}

func readPluginYml(pluginYmlFileName string) (Plugin, error) {
	log.Println("reading plugin file", pluginYmlFileName)

	pluginYml, err := ioutil.ReadFile(pluginYmlFileName)
	if err != nil {
		log.Println("failed to read plugin.yml at", pluginYmlFileName)
		return Plugin{}, nil
	}

	var plugin Plugin
	err = yaml.Unmarshal(pluginYml, &plugin)
	if err != nil {
		return plugin, errors.Wrapf(err, "failed to unmarshal plugin.yml at %s", pluginYmlFileName)
	}
	return plugin, nil
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
