package pkg

import (
	"fmt"
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

func createMock(t *testing.T) func(url string) (resp *http.Response, err error) {
	return func(url string) (resp *http.Response, err error) {
		assert.Equal(t, "http://example.com", url)
		return &http.Response{Body: ioutil.NopCloser(strings.NewReader("content"))}, nil
	}
}

func TestDownloadHandler(t *testing.T) {
	downloadHandler := DownloadHandler{plugins: createMap(testData), downloadPlugin: createMock(t)}

	rr := initRouter(t, "/api/v1/download/ssh-plugin/2.0", "trillian", downloadHandler.handle)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "content", rr.Body.String())
}

func TestDownloadHandlerPluginWithoutAuthentication(t *testing.T) {
	downloadHandler := DownloadHandler{plugins: createMap(testData), downloadPlugin: createMock(t)}

	rr := initRouter(t, "/api/v1/download/ad-plugin/1.0", "", downloadHandler.handle)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "content", rr.Body.String())
}

func TestDownloadHandlerCloudoguPluginWithoutSubject(t *testing.T) {
	downloadHandler := DownloadHandler{plugins: createMap(testData), downloadPlugin: createMock(t)}

	rr := initRouter(t, "/api/v1/download/ssh-plugin/2.0", "", downloadHandler.handle)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestDownloadHandlerReleaseNotFound(t *testing.T) {
	var plugins []Plugin
	downloadHandler := DownloadHandler{plugins: createMap(plugins), downloadPlugin: nil}

	rr := initRouter(t, "/api/v1/download/ssh-plugin/2.0", "trillian", downloadHandler.handle)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDownloadHandlerRemoteFailes(t *testing.T) {
	getMock := func(url string) (resp *http.Response, err error) {
		return nil, fmt.Errorf("failed to handle request: %s", url)
	}

	downloadHandler := DownloadHandler{plugins: createMap(testData), downloadPlugin: getMock}
	rr := initRouter(t, "/api/v1/download/ssh-plugin/2.0", "dent", downloadHandler.handle)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
