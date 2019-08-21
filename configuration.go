package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Configuration struct {
	DescriptorDirectory string
}

func readConfiguration() Configuration {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("failed to read configuration %s: %v", configPath, err)
	}
	configMap := map[string]string{}
	err = yaml.Unmarshal(data, &configMap)
	if err != nil {
		log.Fatalf("failed to unmarshal configuration %s: %v", configPath, err)
	}

	return Configuration{DescriptorDirectory: configMap["descriptor-directory"]}
}
