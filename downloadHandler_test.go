package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
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
		"https://froward.for/api/v1/download/scm-download-plguin/1.2.3",
		generator.DownloadUrl(Plugin{Name: "scm-download-plguin"}, "1.2.3"))
}

func TestUrlGeneratorWithoutForwardedHeader(t *testing.T) {
	generator := NewUrlGenerator(http.Request{
		URL:        nil,
		Proto:      "https",
		ProtoMajor: 0,
		ProtoMinor: 0,
		Header:     http.Header{},
		Host:       "scm.org",
		RequestURI: "https://scm.org/api/v1/plugins/2.0.0",
	})

	assert.Equal(t,
		"https://scm.org/api/v1/download/scm-download-plguin/1.2.3",
		generator.DownloadUrl(Plugin{Name: "scm-download-plguin"}, "1.2.3"))
}

func TestDownloadHandler(t *testing.T) {
	rr := initRouter("/api/v1/download/ssh-plugin/2.0", t)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
}
