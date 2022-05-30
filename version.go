package main

import (
	"github.com/blang/semver/v4"
)

func MustParseVersion(value string) Version {
	return Version{Version: semver.MustParse(value)}
}

type Version struct {
	semver.Version
}

func (v *Version) IsDefault() bool {
	return v.Version.Equals(semver.Version{})
}
