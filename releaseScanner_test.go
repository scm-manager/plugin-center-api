package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIfReleaseFilesAreRead(t *testing.T) {
	configuration := Configuration{DescriptorDirectory: "resources/test/plugins"}

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	assert.Nil(t, err, "unexpected error reading directory", err)
	assert.Equal(t, 3, len(plugins), "wrong number of plugins")
}
