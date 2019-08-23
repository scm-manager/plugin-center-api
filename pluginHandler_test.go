package main

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPluginHandlerHasEmbeddedCollection(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `{"_embedded":{"plugins":`)
}

func TestPluginHandlerReturnsLatestPluginRelease(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"displayName":"ssh plugin"`)
	assert.Contains(t, rr.Body.String(), `"description":"description for ssh plugin"`)
	assert.Contains(t, rr.Body.String(), `"category":"test"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"author":"Cloudogu"`)
}

func TestPluginHandlerReturnsConditionsFromRelease(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=64", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"conditions":{`)
	assert.Contains(t, rr.Body.String(), `"os":"linux"`)
	assert.Contains(t, rr.Body.String(), `"arch":"64"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.1"`)
}

func TestPluginHandlerFiltersForScmVersion(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.0?os=linux&arch=64", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"version":"1.1"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.0"`)
}

func TestPluginHandlerFiltersForOs(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=windows&arch=64", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.NotContains(t, rr.Body.String(), `"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"ad-plugin"`)
}

func TestPluginHandlerFiltersForArch(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1?os=linux&arch=32", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"version":"0.1"`)
}

func TestPluginHandlerTreatsOsAndArchAsOptional(t *testing.T) {
	rr := initRouter("/api/v1/plugins/2.0.1", t)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
}

func initRouter(url string, t *testing.T) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/plugins/{version}", NewPluginHandler(testData))
	router.ServeHTTP(rr, req)

	return rr
}

var testData = []Plugin{
	{
		Name:        "ssh-plugin",
		DisplayName: "ssh plugin",
		Description: "description for ssh plugin",
		Category:    "test",
		Releases: []Release{
			{
				Version: "2.0",
				Conditions: Conditions{
					Os:         "linux",
					Arch:       "64",
					MinVersion: "2.0.1",
				},
				Url:      "http://example.com",
				Date:     "1.01.2019",
				Checksum: "abc",
			},
			{
				Version: "1.1",
				Conditions: Conditions{
					Os:         "linux",
					Arch:       "64",
					MinVersion: "2.0.0",
				},
				Url:      "http://example.com",
				Date:     "1.01.2019",
				Checksum: "abc",
			},
			{
				Version: "0.1",
				Conditions: Conditions{
					Os: "linux",
				},
				Url:      "http://example.com",
				Date:     "1.01.2019",
				Checksum: "abc",
			},
		},
		Author: "Cloudogu",
	},
	{
		Name:        "ad-plugin",
		DisplayName: "active directory plugin",
		Description: "description for ad plugin",
		Category:    "test",
		Releases: []Release{
			{
				Version: "1.0",
				Conditions: Conditions{
					Os: "windows",
				},
				Url:      "http://example.com",
				Date:     "1.01.2019",
				Checksum: "abc",
			},
		},
		Author: "Microsoft",
	},
}
