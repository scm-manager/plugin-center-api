package main

import (
	"io/ioutil"
	"os"
)

func scanDirectory(directory string) ([]Plugin, error) {

	var plugins []Plugin

	pluginDirectories, err := ioutil.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	for _, pluginDirectory := range pluginDirectories {
		println("checking", pluginDirectory)
		plugins = append(plugins, readPluginDirectory(pluginDirectory))
	}

	return plugins, nil
}

func readPluginDirectory(pluginDirectory os.FileInfo) Plugin {
	return Plugin{name: pluginDirectory.Name()}
}
