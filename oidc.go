package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/oauth2"
)

var (
	authenticationRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scm_plugin_center_api_authentication_requests",
		Help: "Total number of authentication requests",
	}, []string{})

	authenticationsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scm_plugin_center_api_authentications",
		Help: "Total number of authentications",
	}, []string{})

	refreshCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scm_plugin_center_api_token_refresh",
		Help: "Total number of token refresh",
	}, []string{})
)

func NewOIDCHandler(configuration OidcConfiguration, templateFs fs.FS) (*OidcHandler, error) {
	provider, err := oidc.NewProvider(context.Background(), configuration.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create oidc provider: %w", err)
	}

	errorTemplate, err := template.ParseFS(templateFs, "layout.gohtml", "error.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to load error template: %w", err)
	}

	callbackTemplate, err := template.ParseFS(templateFs, "layout.gohtml", "callback.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to load callback template: %w", err)
	}

	endpoint := provider.Endpoint()
	if configuration.development {
		// mockoidc, fails without manually set AuthStyleInParams
		endpoint.AuthStyle = oauth2.AuthStyleInParams
	}

	config := oauth2.Config{
		ClientID:     configuration.ClientID,
		ClientSecret: configuration.ClientSecret,
		RedirectURL:  configuration.RedirectURL,
		Endpoint:     endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile", "offline_access"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: configuration.ClientID})
	return &OidcHandler{
		provider,
		config,
		errorTemplate,
		callbackTemplate,
		verifier,
	}, nil
}

type OidcHandler struct {
	provider         *oidc.Provider
	config           oauth2.Config
	errorTemplate    *template.Template
	callbackTemplate *template.Template
	verifier         *oidc.IDTokenVerifier
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (o *OidcHandler) validateInstance(instance string) (*url.URL, error) {
	if instance == "" {
		return nil, fmt.Errorf("is required")
	}

	u, err := url.ParseRequestURI(instance)
	if err != nil {
		return nil, fmt.Errorf("is not a valid url")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("uses a unsupported scheme %s, only http and https is supported", u.Scheme)
	}

	return u, nil
}

func (o *OidcHandler) bearerToken(authorizationHeader string) (string, error) {
	prefix := "Bearer "
	if !strings.HasPrefix(authorizationHeader, prefix) {
		return "", fmt.Errorf("malformed authorization header")
	}

	return authorizationHeader[len(prefix):], nil
}

func (o *OidcHandler) htmlError(w http.ResponseWriter, errorMessage string, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/html")
	err := o.errorTemplate.Execute(w, errorMessage)
	if err != nil {
		http.Error(w, "failed to execute template callback.gohtml", http.StatusInternalServerError)
		return
	}
}

func (o *OidcHandler) jsonError(w http.ResponseWriter, errorMessage string, code int) {
	data, err := json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: errorMessage,
	})
	if err != nil {
		http.Error(w, "failed to marshal json error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(data)
	if err != nil {
		log.Println("failed to write json error to client", err)
	}
}

func (o *OidcHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	_, err := o.validateInstance(instance)
	if err != nil {
		o.htmlError(w, fmt.Sprintf("Query parameter instance %v", err), 400)
		return
	}

	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		authenticationRequestCounter.WithLabelValues().Inc()
		http.Redirect(w, r, o.config.AuthCodeURL(instance), http.StatusFound)
		return
	}

	bearer, err := o.bearerToken(authorizationHeader)
	if err != nil {
		o.htmlError(w, err.Error(), 400)
		return
	}

	_, err = o.verifier.Verify(context.Background(), bearer)
	if err != nil {
		authenticationRequestCounter.WithLabelValues().Inc()
		http.Redirect(w, r, o.config.AuthCodeURL(instance), http.StatusFound)
		return
	}
}

func (o *OidcHandler) Callback(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("state")
	instanceUrl, err := o.validateInstance(instance)
	if err != nil {
		o.htmlError(w, fmt.Sprintf("State instance parameter %v", err), 400)
		return
	}

	oauth2Token, err := o.config.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		o.htmlError(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		o.htmlError(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	idToken, err := o.verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		o.htmlError(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	claim := OidcClaim{}
	err = idToken.Claims(&claim)
	if err != nil {
		o.htmlError(w, "Failed to extract claim from ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	authenticationsCounter.WithLabelValues().Inc()

	subject := claim.Email
	if subject == "" {
		subject = claim.Username
		if subject == "" {
			subject = idToken.Subject
		}
	}

	model := CallbackModel{
		Instance:     instanceUrl.Host,
		Subject:      subject,
		RefreshToken: oauth2Token.RefreshToken,
		Endpoint:     instance,
	}

	w.Header().Set("Content-Type", "text/html")
	err = o.callbackTemplate.Execute(w, model)
	if err != nil {
		http.Error(w, "failed to execute template callback.gohtml", http.StatusInternalServerError)
		return
	}
}

type OidcClaim struct {
	Name     string `json:"name"`
	Username string `json:"preferred_username"`
	Email    string `json:"email"`
}

type CallbackModel struct {
	Instance     string
	Subject      string
	RefreshToken string
	Endpoint     string
}

func (o *OidcHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		o.jsonError(w, "Request does not contain refresh token", http.StatusBadRequest)
		return
	}

	request := RefreshRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		o.jsonError(w, "Failed to unmarshal refresh token", http.StatusBadRequest)
		return
	}

	if request.RefreshToken == "" {
		o.jsonError(w, "Request contains empty refresh token", http.StatusBadRequest)
		return
	}

	source := o.config.TokenSource(context.Background(), &oauth2.Token{RefreshToken: request.RefreshToken})
	token, err := source.Token()
	if err != nil {
		o.jsonError(w, "Failed to refresh token", http.StatusUnauthorized)
		return
	}

	response, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "Failed to marshal token response", http.StatusInternalServerError)
		return
	}

	refreshCounter.WithLabelValues().Inc()

	_, err = w.Write(response)
	if err != nil {
		log.Println("failed to write response to client", err)
	}
}

func (o *OidcHandler) WithIdToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		bearer, err := o.bearerToken(authorizationHeader)
		if err != nil {
			o.jsonError(w, err.Error(), 400)
			return
		}

		idToken, err := o.verifier.Verify(context.Background(), bearer)
		if err != nil {
			o.jsonError(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		subject := Subject{
			Id: idToken.Subject,
		}
		ctx := context.WithValue(r.Context(), "subject", &subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type Subject struct {
	Id string
}
