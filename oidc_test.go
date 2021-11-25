package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func createTestOidcHandler(issuer string) *OidcHandler {
	return NewOIDCHandler(OidcConfiguration{
		issuer,
		"pc-unit-test",
		"unit-test-secret",
		"http://localhost:8080/api/v1/auth/oidc/callback",
	})
}

func createOidcTestServer() *OidcTestServer {
	oidcServer := &OidcTestServer{}
	oidcServer.server = httptest.NewServer(oidcServer)
	oidcServer.URL = oidcServer.server.URL
	return oidcServer
}

type IdTokenHandler struct {
}

func (oe *IdTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idToken := r.Context().Value("idToken")
	if idToken != nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

type OidcTestServer struct {
	server *httptest.Server
	URL    string
}

func (oe *OidcTestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("resources/test/oidc/openid-configuration.json")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read openid configuration: %v", err), 500)
		return
	}

	conf := strings.ReplaceAll(string(data), "{issuer}", oe.server.URL)

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(conf))
}

func (oe *OidcTestServer) Close() {
	oe.server.Close()
}

func authenticate(server *OidcTestServer, requestUrl string, authorization string) *httptest.ResponseRecorder {
	handler := createTestOidcHandler(server.URL)
	r := httptest.NewRequest(http.MethodGet, requestUrl, nil)

	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}

	w := httptest.NewRecorder()
	handler.Authenticate(w, r)
	return w
}

func TestOidcHandler_Authenticate_withoutInstance(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc", "")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withNonUrlInstance(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc?instance=xyz", "")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withNonHttpUrl(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc?instance=file:///etc/passwd", "")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc?instance=https://scm-manager.org", "")

	assert.Equal(t, 302, response.Code)
	u, err := url.ParseRequestURI(response.Header().Get("Location"))
	assert.NoError(t, err, "Failed to parse location header")

	assert.Equal(t, "/openid-connect/auth", u.Path)
	assert.Equal(t, "https://scm-manager.org", u.Query().Get("state"))
	assert.Equal(t, "pc-unit-test", u.Query().Get("client_id"))

	assert.True(t,
		strings.HasPrefix(u.Query().Get("redirect_uri"), "http://localhost:8080/api/v1/auth/oidc/callback"),
		"redirect uri does not match our configuration",
	)
}

func TestOidcHandler_Authenticate_withMalformedAuthorizationHeader(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc?instance=https://scm-manager.org", "Bearer")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withWrongAuthorizationScheme(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc?instance=https://scm-manager.org", "Basic abc")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withNonValidAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(server, "/oidc?instance=https://scm-manager.org", "Bearer abc")
	assert.Equal(t, 302, response.Code)
}

func TestOidcHandler_Authenticate_withValidAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(server.URL)
	handler.verify = func(ctx context.Context, token string) (*oidc.IDToken, error) {
		return &oidc.IDToken{}, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/oidc?instance=https://scm-manager.org", nil)
	r.Header.Set("Authorization", "Bearer xyz")

	w := httptest.NewRecorder()
	handler.Authenticate(w, r)

	response := authenticate(server, "/oidc?instance=https://scm-manager.org", "Bearer abc")
	assert.Equal(t, 302, response.Code)
}

func TestOidcHandler_WithIdToken_withoutAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(server.URL)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	w := httptest.NewRecorder()
	o.WithIdToken(&IdTokenHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestOidcHandler_WithIdToken_withMalformedAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(server.URL)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "xyz")

	w := httptest.NewRecorder()
	o.WithIdToken(&IdTokenHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_WithIdToken_withInvalidScheme(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(server.URL)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "Basic xyz")

	w := httptest.NewRecorder()
	o.WithIdToken(&IdTokenHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_WithIdToken_withFailedAuthentication(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(server.URL)
	o.verify = func(ctx context.Context, token string) (*oidc.IDToken, error) {
		return nil, fmt.Errorf("nope")
	}

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "Bearer abc")

	w := httptest.NewRecorder()
	o.WithIdToken(&IdTokenHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOidcHandler_WithIdToken(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(server.URL)
	o.verify = func(ctx context.Context, token string) (*oidc.IDToken, error) {
		return &oidc.IDToken{}, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "Bearer awesome")

	w := httptest.NewRecorder()
	o.WithIdToken(&IdTokenHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
