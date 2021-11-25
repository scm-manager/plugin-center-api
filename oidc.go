package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func NewOIDCHandler(configuration OidcConfiguration) *OidcHandler {
	provider, err := oidc.NewProvider(context.Background(), configuration.Issuer)
	if err != nil {
		log.Fatal("failed to create oidc provider", err)
	}

	config := oauth2.Config{
		ClientID:     configuration.ClientID,
		ClientSecret: configuration.ClientSecret,
		RedirectURL:  configuration.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}

	verify := provider.Verifier(&oidc.Config{ClientID: configuration.ClientID}).Verify
	return &OidcHandler{provider, config, verify}
}

type OidcHandler struct {
	provider *oidc.Provider
	config   oauth2.Config
	verify   func(ctx context.Context, token string) (*oidc.IDToken, error)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (o *OidcHandler) validateInstance(instance string) error {
	if instance == "" {
		return fmt.Errorf("is required")
	}

	u, err := url.ParseRequestURI(instance)
	if err != nil {
		return fmt.Errorf("is not a valid url")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("uses a unsupported scheme %s, only http and https is supported", u.Scheme)
	}

	return nil
}

func (o *OidcHandler) bearerToken(authorizationHeader string) (string, error) {
	prefix := "Bearer "
	if !strings.HasPrefix(authorizationHeader, prefix) {
		return "", fmt.Errorf("malformed authorization header")
	}

	return authorizationHeader[len(prefix):], nil
}

func (o *OidcHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	err := o.validateInstance(instance)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query parameter instance %v", err), 400)
		return
	}

	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		http.Redirect(w, r, o.config.AuthCodeURL(instance), http.StatusFound)
		return
	}

	bearer, err := o.bearerToken(authorizationHeader)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err = o.verify(context.Background(), bearer)
	if err != nil {
		http.Redirect(w, r, o.config.AuthCodeURL(instance), http.StatusFound)
		return
	}
}

func (o *OidcHandler) Callback(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("state")
	err := o.validateInstance(instance)
	if err != nil {
		http.Error(w, fmt.Sprintf("State instance parameter %v", err), 400)
		return
	}

	oauth2Token, err := o.config.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	idToken, err := o.verify(context.Background(), rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (o *OidcHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Request does not contain refresh token", http.StatusBadRequest)
		return
	}

	request := RefreshRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		http.Error(w, "Failed to unmarshal refresh token", http.StatusBadRequest)
		return
	}

	if request.RefreshToken == "" {
		http.Error(w, "Request contains empty refresh token", http.StatusBadRequest)
		return
	}

	source := o.config.TokenSource(context.Background(), &oauth2.Token{RefreshToken: request.RefreshToken})
	token, err := source.Token()
	if err != nil {
		http.Error(w, "Failed to refresh token", http.StatusUnauthorized)
		return
	}

	response, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "Failed to marshal token response", http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func (o *OidcHandler) WithIdToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			//TODO metrics for anonymous access
			next.ServeHTTP(w, r)
			return
		}

		bearer, err := o.bearerToken(authorizationHeader)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		idToken, err := o.verify(context.Background(), bearer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "idToken", idToken)
		// TODO metrics for authenticated access
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
