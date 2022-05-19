package main

import (
	"encoding/json"
	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
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

func (v *Version) UnmarshalJSON(data []byte) error {
	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return nil
	}

	version, err := semver.Parse(value)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse version %s", value)
	}
	v.Version = version
	return nil
}
