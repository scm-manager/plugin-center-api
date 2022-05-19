package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion_IsDefault(t *testing.T) {
	v := Version{}
	assert.True(t, v.IsDefault())

	v = MustParseVersion("1.0.0")
	assert.False(t, v.IsDefault())
}
