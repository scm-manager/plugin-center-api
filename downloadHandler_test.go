package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUrlGeneratorWithForwardedHeader(t *testing.T) {
	generator := NewUrlGenerator(http.Request{
		Header: http.Header{
			"X-Forwarded-Host":  {"froward.for"},
			"X-Forwarded-Proto": {"https"},
		},
	})

	assert.Equal(t,
		"https://froward.for/api/v1/download/scm-download-plugin/1.2.3",
		generator.DownloadUrl(Plugin{Name: "scm-download-plugin"}, "1.2.3"))
}

func TestUrlGeneratorWithoutForwardedHeader(t *testing.T) {
	generator := NewUrlGenerator(http.Request{
		URL:        nil,
		Proto:      "https",
		ProtoMajor: 0,
		ProtoMinor: 0,
		Header:     http.Header{},
		Host:       "scm.org",
		RequestURI: "http://scm.org/api/v1/plugins/2.0.0",
	})

	assert.Equal(t,
		"http://scm.org/api/v1/download/scm-download-plugin/1.2.3",
		generator.DownloadUrl(Plugin{Name: "scm-download-plugin"}, "1.2.3"))
}

func TestDownloadHandler(t *testing.T) {

	getMock := func(url string) (resp *http.Response, err error) {
		assert.Equal(t, "http://example.com", url)
		return &http.Response{Body: ioutil.NopCloser(strings.NewReader("content"))}, nil
	}

	downloadHandler := DownloadHandler{plugins: createMap(testData), downloadPlugin: getMock}

	rr := initRouter("/api/v1/download/ssh-plugin/2.0", t, downloadHandler.handle)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "content", rr.Body.String())
}
