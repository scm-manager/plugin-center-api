package main

type Plugins struct {
	Id       string       `yaml:"id"`
	Versions VersionRange `yaml:"versions"`
	Sequence int          `yaml:"sequence"`
	Plugins  []string     `yaml:"plugins"`
}

type Description struct {
	Name     string   `yaml:"name" json:"name"`
	Features []string `yaml:"features" json:"features"`
}

type Descriptions map[string]Description

type PluginSet struct {
	Id           string       `json:"id"`
	Versions     VersionRange `json:"versions"`
	Sequence     int          `json:"sequence"`
	Plugins      []string     `json:"plugins"`
	Descriptions Descriptions `json:"descriptions"`
}
