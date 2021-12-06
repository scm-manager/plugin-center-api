package main

import (
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func createTestOidcHandler(t *testing.T, server *OidcTestServer) *OidcHandler {
	static, err := fs.Sub(assets, "html")
	assert.NoError(t, err)

	mockOIDC := server.server
	handler, err := NewOIDCHandler(OidcConfiguration{
		mockOIDC.Issuer(),
		mockOIDC.ClientID,
		mockOIDC.ClientSecret,
		"http://localhost:8080/api/v1/auth/oidc/callback",
		true,
	}, static)
	assert.NoError(t, err)
	return handler
}

func createOidcTestServer() *OidcTestServer {
	m, err := mockoidc.Run()
	if err != nil {
		panic(err)
	}

	return &OidcTestServer{
		server: m,
		URL:    m.Issuer(),
	}
}

type SubjectHandler struct {
}

func (oe *SubjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subject := r.Context().Value("subject")
	if subject != nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

type OidcTestServer struct {
	server *mockoidc.MockOIDC
	URL    string
}

func (oe *OidcTestServer) Close() {
	oe.server.Shutdown()
}

func authenticate(t *testing.T, server *OidcTestServer, requestUrl string, authorization string) *httptest.ResponseRecorder {
	handler := createTestOidcHandler(t, server)
	r := httptest.NewRequest(http.MethodGet, requestUrl, nil)

	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}

	w := httptest.NewRecorder()
	handler.Authenticate(w, r)
	return w
}

func TestNewOIDCHandlerWithoutIssuer(t *testing.T) {
	_, err := NewOIDCHandler(OidcConfiguration{}, assets)
	assert.Contains(t, err.Error(), "provider")
}

func TestNewOIDCHandlerWithoutTemplates(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	_, err := NewOIDCHandler(OidcConfiguration{
		server.server.Issuer(),
		"client",
		"secret",
		"http://localhost:8080/api/v1/auth/oidc/callback",
		true,
	}, assets)
	assert.Contains(t, err.Error(), "error template")
}

func TestNewOIDCHandlerWithoutCallbackTemplate(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	templates := os.DirFS("./resources/test/oidc/error-template")
	_, err := NewOIDCHandler(OidcConfiguration{
		server.server.Issuer(),
		"client",
		"secret",
		"http://localhost:8080/api/v1/auth/oidc/callback",
		true,
	}, templates)

	assert.Contains(t, err.Error(), "callback template")
}

func TestOidcHandler_Authenticate_withoutInstance(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc", "")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withNonUrlInstance(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc?instance=xyz", "")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withNonHttpUrl(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc?instance=file:///etc/passwd", "")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc?instance=https://scm-manager.org", "")

	assert.Equal(t, 302, response.Code)
	u, err := url.ParseRequestURI(response.Header().Get("Location"))
	assert.NoError(t, err, "Failed to parse location header")

	assert.Equal(t, "/oidc/authorize", u.Path)
	assert.Equal(t, "https://scm-manager.org", u.Query().Get("state"))
	assert.Equal(t, server.server.ClientID, u.Query().Get("client_id"))

	assert.True(t,
		strings.HasPrefix(u.Query().Get("redirect_uri"), "http://localhost:8080/api/v1/auth/oidc/callback"),
		"redirect uri does not match our configuration",
	)
}

func TestOidcHandler_Authenticate_withMalformedAuthorizationHeader(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc?instance=https://scm-manager.org", "Bearer")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withWrongAuthorizationScheme(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc?instance=https://scm-manager.org", "Basic abc")
	assert.Equal(t, 400, response.Code)
}

func TestOidcHandler_Authenticate_withNonValidAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	response := authenticate(t, server, "/oidc?instance=https://scm-manager.org", "Bearer abc")
	assert.Equal(t, 302, response.Code)
}

func TestOidcHandler_Authenticate_withValidAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/oidc?instance=https://scm-manager.org", nil)
	r.Header.Set("Authorization", "Bearer xyz")

	w := httptest.NewRecorder()
	handler.Authenticate(w, r)

	response := authenticate(t, server, "/oidc?instance=https://scm-manager.org", "Bearer abc")
	assert.Equal(t, 302, response.Code)
}

func TestOidcHandler_WithIdToken_withoutAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	w := httptest.NewRecorder()
	o.WithIdToken(&SubjectHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestOidcHandler_WithIdToken_withMalformedAuthorization(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "xyz")

	w := httptest.NewRecorder()
	o.WithIdToken(&SubjectHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_WithIdToken_withInvalidScheme(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "Basic xyz")

	w := httptest.NewRecorder()
	o.WithIdToken(&SubjectHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_WithIdToken_withFailedAuthentication(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	o := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	r.Header.Set("Authorization", "Bearer abc")

	w := httptest.NewRecorder()
	o.WithIdToken(&SubjectHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOidcHandler_WithIdToken(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	s, err := server.server.SessionStore.NewSession(oidc.ScopeOpenID+" profile email", "12345", mockoidc.DefaultUser())
	assert.NoError(t, err)
	s.Granted = true

	o := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	at, err := s.AccessToken(server.server.Config(), server.server.Keypair, time.Now())
	assert.NoError(t, err)
	r.Header.Set("Authorization", "Bearer "+at)

	w := httptest.NewRecorder()
	o.WithIdToken(&SubjectHandler{}).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOidcHandler_RefreshFailWithoutBody(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodPost, "/id-token", nil)
	w := httptest.NewRecorder()

	handler.Refresh(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_RefreshWithInvalidJson(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodPost, "/id-token", bytes.NewBuffer([]byte("abc")))
	w := httptest.NewRecorder()

	handler.Refresh(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_RefreshWithoutRefreshToken(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodPost, "/id-token", bytes.NewBuffer([]byte("{}")))
	w := httptest.NewRecorder()

	handler.Refresh(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_Refresh(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	s, err := server.server.SessionStore.NewSession("openid email profile", "12345", mockoidc.DefaultUser())
	assert.NoError(t, err)
	rt, err := s.RefreshToken(server.server.Config(), server.server.Keypair, time.Now())
	assert.NoError(t, err)

	handler := createTestOidcHandler(t, server)

	data, err := json.Marshal(&RefreshRequest{
		RefreshToken: rt,
	})
	assert.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(data))
	w := httptest.NewRecorder()

	handler.Refresh(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	data, err = ioutil.ReadAll(w.Body)
	assert.NoError(t, err)

	token := oauth2.Token{}
	err = json.Unmarshal(data, &token)
	assert.NoError(t, err)

	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.RefreshToken)
}

func TestOidcHandler_CallbackWithoutState(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token", nil)
	w := httptest.NewRecorder()

	handler.Callback(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_CallbackWithInvalidState(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token?state=nonUrl", nil)
	w := httptest.NewRecorder()

	handler.Callback(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOidcHandler_CallbackWithInvalidCode(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	handler := createTestOidcHandler(t, server)

	r := httptest.NewRequest(http.MethodGet, "/id-token?state=https://scm-manager.org&code=abc", nil)
	w := httptest.NewRecorder()

	handler.Callback(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func callback(t *testing.T, server *OidcTestServer) *httptest.ResponseRecorder {
	response := authenticate(t, server, "/oidc/?instance=https://scm-manager.org", "")

	r := httptest.NewRequest(http.MethodGet, response.Header().Get("Location"), nil)
	w := httptest.NewRecorder()

	server.server.Authorize(w, r)

	handler := createTestOidcHandler(t, server)

	r = httptest.NewRequest(http.MethodGet, w.Header().Get("Location"), nil)
	w = httptest.NewRecorder()

	handler.Callback(w, r)

	return w
}

func TestNewOIDCHandler_CallbackShouldRenderRefreshToken(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	server.server.QueueUser(mockoidc.DefaultUser())

	response := callback(t, server)
	assert.Equal(t, http.StatusOK, response.Code)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	assert.NoError(t, err)

	rt, _ := doc.Find("input[name=refresh_token]").Attr("value")
	assert.NotEmpty(t, rt)
}

func TestNewOIDCHandler_CallbackShouldRenderInstance(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	server.server.QueueUser(mockoidc.DefaultUser())

	response := callback(t, server)
	assert.Equal(t, http.StatusOK, response.Code)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	assert.NoError(t, err)

	assert.Equal(t, "scm-manager.org", doc.Find("#instance").Text())
}

func TestNewOIDCHandler_CallbackShouldUseEmailAsSubject(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	server.server.QueueUser(&mockoidc.MockUser{
		Subject:           "1234567890",
		Email:             "trillian@hitchhiker.com",
		PreferredUsername: "trillian",
	})

	response := callback(t, server)
	assert.Equal(t, http.StatusOK, response.Code)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	assert.NoError(t, err)

	assert.Equal(t, "trillian@hitchhiker.com", doc.Find("#subject").Text())
}

func TestNewOIDCHandler_CallbackShouldUsePreferredUsernameAsSubject(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	server.server.QueueUser(&mockoidc.MockUser{
		Subject:           "1234567890",
		PreferredUsername: "trillian",
	})

	response := callback(t, server)
	assert.Equal(t, http.StatusOK, response.Code)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	assert.NoError(t, err)

	assert.Equal(t, "trillian", doc.Find("#subject").Text())
}

func TestNewOIDCHandler_CallbackShouldUseSubject(t *testing.T) {
	server := createOidcTestServer()
	defer server.Close()

	server.server.QueueUser(&mockoidc.MockUser{
		Subject: "1234567890",
	})

	response := callback(t, server)
	assert.Equal(t, http.StatusOK, response.Code)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	assert.NoError(t, err)

	assert.Equal(t, "1234567890", doc.Find("#subject").Text())
}
