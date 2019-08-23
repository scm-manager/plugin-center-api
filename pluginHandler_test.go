package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPluginHandlerHasEmbeddedCollection(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := NewPluginHandler(testData)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `{"_embedded":{"plugins":`)
}

func TestPluginHandlerReturnsLatestPluginRelease(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := NewPluginHandler(testData)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"name":"ssh-plugin"`)
	assert.Contains(t, rr.Body.String(), `"displayName":"ssh plugin"`)
	assert.Contains(t, rr.Body.String(), `"description":"description for ssh plugin"`)
	assert.Contains(t, rr.Body.String(), `"category":"test"`)
	assert.Contains(t, rr.Body.String(), `"version":"2.0"`)
	assert.Contains(t, rr.Body.String(), `"author":"Cloudogu"`)
}

func TestPluginHandlerReturnsConditionsFromRelease(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := NewPluginHandler(testData)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), `"conditions":{`)
	assert.Contains(t, rr.Body.String(), `"os":"Linux"`)
	assert.Contains(t, rr.Body.String(), `"arch":"64"`)
	assert.Contains(t, rr.Body.String(), `"minVersion":"2.0.1"`)
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
					Os:         "Linux",
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
					Os:         "Linux",
					Arch:       "64",
					MinVersion: "2.0.0",
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
					Os: "Windows",
				},
				Url:      "http://example.com",
				Date:     "1.01.2019",
				Checksum: "abc",
			},
		},
		Author: "Microsoft",
	},
}
