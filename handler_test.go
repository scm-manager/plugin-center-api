package main

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func initRouter(t *testing.T, url string, subject string, handler func(http.ResponseWriter, *http.Request)) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	if subject != "" {
		ctx := context.WithValue(req.Context(), "subject", &Subject{Id: subject})
		req = req.WithContext(ctx)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/plugins/{version}", handler)
	router.HandleFunc("/api/v1/download/{plugin}/{version}", handler)
	router.ServeHTTP(rr, req)

	return rr
}

var testData = []Plugin{
	{
		Name:        "ssh-plugin",
		DisplayName: "ssh plugin",
		Description: "description for ssh plugin",
		Category:    "test",
		Type:        "CLOUDOGU",
		AvatarUrl:   "/images/ssh-logo.png",
		Releases: []Release{
			{
				Version: "2.0",
				Conditions: Conditions{
					Os:         []string{"linux"},
					Arch:       "64",
					MinVersion: "2.0.1",
				},
				Dependencies:         []string{"scm-mail-plugin"},
				OptionalDependencies: []string{"scm-review-plugin"},
				Url:                  "http://example.com",
				Date:                 "1.01.2019",
				Checksum:             "abc",
				InstallLink:          "myCloudogu.com/install/my_plugin",
			},
			{
				Version: "1.1",
				Conditions: Conditions{
					Os:         []string{"linux"},
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
					Os: []string{"linux"},
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
					Os: []string{"windows", "linux"},
				},
				Url:      "http://example.com",
				Date:     "1.01.2019",
				Checksum: "abc",
			},
		},
		Author: "Microsoft",
	},
}

var testDataPluginSets = []PluginSet{
	{
		Id:       "plug-and-play",
		Versions: MustParseVersionRange(">=2.0.0 <3.0.0"),
		Sequence: 1,
		Plugins:  []string{"scm-editor-plugin", "scm-readme-plugin"},
		Descriptions: map[string]Description{
			"en": {
				Name:     "Plug'n Play",
				Features: []string{"Feature 1", "Feature 2", "Feature 3"},
			},
			"de": {
				Name:     "Anklicken und loslegen",
				Features: []string{"Merkmal 1", "Merkmal 2", "Merkmal 3"},
			},
		},
	},
	{
		Id:       "administration-and-management",
		Versions: MustParseVersionRange(">=2.0.1 <3.0.0"),
		Sequence: 2,
		Plugins:  []string{"scm-cas-plugin"},
		Descriptions: map[string]Description{
			"en": {
				Name:     "Administration and Management",
				Features: []string{"Feature 1", "Feature 2", "Feature 3"},
			},
		},
	},
}
