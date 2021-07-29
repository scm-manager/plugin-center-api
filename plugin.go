package main

type Conditions struct {
	Os         []string `yaml:"os"`
	Arch       string   `yaml:"arch"`
	MinVersion string   `yaml:"minVersion"`
}

type Release struct {
	Version              string     `yaml:"tag"`
	Conditions           Conditions `yaml:"conditions"`
	Dependencies         []string   `yaml:"dependencies"`
	OptionalDependencies []string   `yaml:"optionalDependencies"`
	Url                  string     `yaml:"url"`
	Date                 string     `yaml:"date"`
	Checksum             string     `yaml:"checksum"`
	InstallLink          string     `yaml:"installLink"`
}

type Plugin struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"displayName"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
	Releases    []Release
	Author      string `yaml:"author"`
	Type        string `yaml:"type"`
	AvatarUrl   string `yaml:"avatarUrl"`
}
