package main

import (
	"encoding/json"
	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
)

func MustParseVersionRange(value string) VersionRange {
	r := semver.MustParseRange(value)
	return VersionRange{Value: value, Range: r}
}

type VersionRange struct {
	Value string
	Range semver.Range
}

func (r *VersionRange) Contains(version Version) bool {
	return r.Range(version.Version)
}

func (r *VersionRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Value)
}

func (r *VersionRange) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var buf string
	err := unmarshal(&buf)
	if err != nil {
		return nil
	}

	ra, err := semver.ParseRange(buf)
	if err != nil {
		return errors.Wrap(err, "Failed to parse range")
	}
	r.Value = buf
	r.Range = ra
	return nil
}

func (r *VersionRange) UnmarshalJSON(data []byte) error {
	return r.UnmarshalYAML(func(i interface{}) error {
		return json.Unmarshal(data, i)
	})
}
