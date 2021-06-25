package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPluginHandlerHasEmbeddedCollection(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `{"_embedded":{"plugins":`)
}

func TestPluginHandlerReturnsLatestPluginRelease(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"displayName":"ssh plugin"`)
	assert.Contains(t, rr.Body.String(), `"description":"description for ssh plugin"`)
	assert.Contains(t, rr.Body.String(), `"category":"test"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"author":"Cloudogu"`)
	assert.Contains(t, rr.Body.String(), `"sha256sum":"abc"`)
}

func TestPluginHandlerReturnsConditionsFromRelease(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"conditions":{`)
	assert.Contains(t, rr.Body.String(), `"os":["linux"]`)
	assert.Contains(t, rr.Body.String(), `"arch":"64"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.1"`)
}

func TestPluginHandlerReturnsDependenciesFromRelease(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"dependencies":["scm-mail-plugin"]`)
	assert.Contains(t, rr.Body.String(), `"optionalDependencies":["scm-review-plugin"]`)
}

func TestPluginHandlerReturnsEmptyDependenciesWhenNotSetInRelease(t *testing.T) {
	rr := initRouter("/api/v1/plugins/1.0.0?os=windows", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"dependencies":[]`)
}

func TestPluginHandlerFiltersForScmVersion(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.0?os=linux&arch=64", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"version":"1.1"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.0"`)
}

func TestPluginHandlerFiltersForOs(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=windows&arch=64", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.NotContains(t, rr.Body.String(), `"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"ad-plugin"`)
}

func TestPluginHandlerFiltersForArch(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=32", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"version":"0.1"`)
}

func TestPluginHandlerTreatsOsAndArchAsOptional(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
}

func TestPluginHandlerRewritesDownloadUrl(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"type":"SCM"`)
}

func TestPluginHandlerGetsRightDataForCloudoguPlugin(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1", t, NewPluginHandler(testData))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"type":"CLOUDOGU"`)
	assert.Contains(t, rr.Body.String(), `"install":{"href":"myCloudogu.com/install/my_plugin"}`)
}
