package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func initRouter(url string, t *testing.T) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/plugins/{version}", NewPluginHandler(testData))
	router.HandleFunc("/api/v1/download/{plugin}/{version}", NewDownloadHandler(testData))
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
