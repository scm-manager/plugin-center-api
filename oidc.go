package main

import (
	"context"
	"encoding/json"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func NewOIDCHandler() *OidcHandler {
	provider, err := oidc.NewProvider(context.Background(), "http://localhost:8080/auth/realms/master")
	if err != nil {
		panic(err)
	}
	config := oauth2.Config{
		ClientID:     "plugin-center",
		ClientSecret: "ce40c2dd-67fa-4e03-9628-954f0f895e3c",
		RedirectURL:  "http://localhost:8000/oidc/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: "plugin-center"})

	return &OidcHandler{provider, config, verifier, "plugin-center-state"}
}

type OidcHandler struct {
	provider *oidc.Provider
	config   oauth2.Config
	verifier *oidc.IDTokenVerifier
	state    string
}

func (o *OidcHandler) Handler(w http.ResponseWriter, r *http.Request) {

	rawAccessToken := r.Header.Get("Authorization")
	if rawAccessToken == "" {
		http.Redirect(w, r, o.config.AuthCodeURL(o.state), http.StatusFound)
		return
	}

	parts := strings.Split(rawAccessToken, " ")
	if len(parts) != 2 {
		w.WriteHeader(400)
		return
	}
	_, err := o.verifier.Verify(context.Background(), parts[1])

	if err != nil {
		http.Redirect(w, r, o.config.AuthCodeURL(o.state), http.StatusFound)
		return
	}
}

func (o *OidcHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != o.state {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := o.config.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}
	idToken, err := o.verifier.Verify(context.Background(), rawIDToken)
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

func (o *OidcHandler) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawAccessToken := r.Header.Get("Authorization")
		if rawAccessToken == "" {
			//TODO metrics for anonymous access
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(rawAccessToken, " ")
		if len(parts) != 2 {
			w.WriteHeader(400)
			return
		}
		idToken, err := o.verifier.Verify(context.Background(), parts[1])

		if err != nil {
			log.Println("authentication failed", err)
			w.WriteHeader(401)
			return
		}

		ctx := context.WithValue(r.Context(), "idToken", idToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (o *OidcHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		//TODO Log error
		w.WriteHeader(400)
		return
	}

	request := RefreshRequest{}

	err = json.Unmarshal(data, &request)

	if err != nil {
		w.WriteHeader(400)
		return
	}

	ts := o.config.TokenSource(context.Background(), &oauth2.Token{RefreshToken: request.RefreshToken})

	token, err := ts.Token()

	if err != nil {
		w.WriteHeader(401)
		return
	}

	response, err := json.Marshal(token)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Write(response)
}
