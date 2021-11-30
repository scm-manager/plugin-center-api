package main

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Configuration struct {
	DescriptorDirectory string `yaml:"descriptor-directory" envconfig:"CONFIG_DESCRIPTOR_DIRECTORY"`
	Port                int    `yaml:"port" envconfig:"CONFIG_PORT" default:"8000"`
	Oidc                OidcConfiguration
}

type OidcConfiguration struct {
	Issuer       string `yaml:"issuer" envconfig:"CONFIG_OIDC_ISSUER"`
	ClientID     string `yaml:"client-id" envconfig:"CONFIG_OIDC_CLIENT_ID"`
	ClientSecret string `yaml:"client-secret" envconfig:"CONFIG_OIDC_CLIENT_SECRET"`
	RedirectURL  string `yaml:"redirect-url" envconfig:"CONFIG_OIDC_REDIRECT_URL"`
}

func (oc OidcConfiguration) IsEnabled() bool {
	return oc.Issuer != ""
}

func readConfiguration() Configuration {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config := Configuration{}
	config.Oidc = OidcConfiguration{}

	exists, err := exists(configPath)
	if err != nil {
		log.Fatalf("failed to check stat of %s: %v", configPath, err)
	}

	if exists {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Fatalf("failed to read configuration %s: %v", configPath, err)
		}

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			log.Fatalf("failed to unmarshal configuration %s: %v", configPath, err)
		}
	}

	err = envconfig.Process("CONFIG", &config)
	if err != nil {
		log.Fatalf("failed to read configuration from environment: %v", err)
	}

	return config
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
