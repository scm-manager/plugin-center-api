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

func TestGetListenerAddress(t *testing.T) {
	address := getListenerAddress(42)

	assert.Equal(t, ":42", address)
}

func TestGetListenerAddress_shouldSetLocalhostIfStageIsDevelopment(t *testing.T) {
	t.Setenv("STAGE", "development")

	address := getListenerAddress(42)

	assert.Equal(t, "127.0.0.1:42", address)
}
