package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigureRouter(t *testing.T) {
	configuration := readConfiguration()
	r := configureRouter(configuration)
	assert.NotNil(t, r)
}
