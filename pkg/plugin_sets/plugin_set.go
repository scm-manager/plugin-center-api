package plugin_sets

type Plugins struct {
	Id       string       `yaml:"id"`
	Versions VersionRange `yaml:"versions"`
	Plugins  []string     `yaml:"plugins"`
}

type Description struct {
	Name     string   `yaml:"name"`
	Features []string `yaml:"features"`
}

type Descriptions map[string]Description

type PluginSet struct {
	Id           string
	Versions     VersionRange
	Plugins      []string
	Descriptions Descriptions
}
