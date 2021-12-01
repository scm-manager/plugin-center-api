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

func TestConfigureRouterWithOidc(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()
	t.Setenv("CONFIG_OIDC_ISSUER", server.URL)

	configuration := readConfiguration()
	r := configureRouter(configuration)
	assert.NotNil(t, r)
}
