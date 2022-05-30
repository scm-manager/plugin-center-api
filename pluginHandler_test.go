package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPluginHandlerHasEmbeddedCollections(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1?os=linux&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	var response Response
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Embedded["plugins"])
	assert.Len(t, response.Embedded["plugin-sets"], 2)
}

func TestPluginHandlerReturnsLatestPluginRelease(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1?os=linux&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

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
	rr := initRouter(t, "/api/v1/plugins/2.0.1?os=linux&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"conditions":{`)
	assert.Contains(t, rr.Body.String(), `"os":["linux"]`)
	assert.Contains(t, rr.Body.String(), `"arch":"64"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.1"`)
}

func TestPluginHandlerReturnsDependenciesFromRelease(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1?os=linux&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"dependencies":["scm-mail-plugin"]`)
	assert.Contains(t, rr.Body.String(), `"optionalDependencies":["scm-review-plugin"]`)
}

func TestPluginHandlerReturnsEmptyDependenciesWhenNotSetInRelease(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/1.0.0?os=windows", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"dependencies":[]`)
}

func TestPluginHandlerFiltersForScmVersion(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.0?os=linux&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"version":"1.1"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.0"`)
	var response Response
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Embedded["plugin-sets"], 1)
}

func TestPluginHandlerFiltersForOs(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1?os=windows&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.NotContains(t, rr.Body.String(), `"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"ad-plugin"`)
}

func TestPluginHandlerFiltersForArch(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1?os=linux&arch=32", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"version":"0.1"`)
}

func TestPluginHandlerTreatsOsAndArchAsOptional(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
}

func TestPluginHandlerRewritesDownloadUrl(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"type":"SCM"`)
}

func TestPluginHandlerGetsRightDataForCloudoguPlugin(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.1", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"type":"CLOUDOGU"`)
	assert.Contains(t, rr.Body.String(), `"install":{"href":"myCloudogu.com/install/my_plugin"}`)
}

func TestPluginHandlerReturnsPluginsSets(t *testing.T) {
	rr := initRouter(t, "/api/v1/plugins/2.0.0?os=linux&arch=64", "", NewPluginHandler(testData, testDataPluginSets))

	assert.Equal(t, http.StatusOK, rr.Code)

	var response Response
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Embedded["plugin-sets"], 1)

	println(rr.Body.String())
	assert.Contains(t, rr.Body.String(), `"id":"plug-and-play"`)
	assert.Contains(t, rr.Body.String(), `"sequence":1`)
	assert.Contains(t, rr.Body.String(), `"plugins":["scm-editor-plugin","scm-readme-plugin"]`)
	assert.Contains(t, rr.Body.String(), `"de":{"name":"Anklicken und loslegen","features":["Merkmal 1","Merkmal 2","Merkmal 3"]`)
	assert.Contains(t, rr.Body.String(), `"en":{"name":"Plug'n Play","features":["Feature 1","Feature 2","Feature 3"]`)
}
