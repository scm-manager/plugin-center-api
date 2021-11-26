package main

import (
	"embed"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/fs"
	"log"
	"net/http"
	"strconv"
)

//go:embed html
var assets embed.FS

func main() {
	configuration := readConfiguration()

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	if err != nil {
		log.Fatalln("could not parse plugins", err)
	}

	static, err := fs.Sub(assets, "html")
	if err != nil {
		log.Fatal("failed to load static files", err)
	}

	oidc := NewOIDCHandler(configuration.Oidc, static)

	r := mux.NewRouter()

	// api
	r.Handle("/api/v1/plugins/{version}", oidc.WithIdToken(NewPluginHandler(plugins)))
	r.Handle("/api/v1/download/{plugin}/{version}", oidc.WithIdToken(NewDownloadHandler(plugins)))

	// oidc
	r.HandleFunc("/api/v1/auth/oidc", oidc.Authenticate)
	r.HandleFunc("/api/v1/auth/oidc/callback", oidc.Callback)
	r.HandleFunc("/api/v1/auth/oidc/refresh", oidc.Refresh)

	// static assets
	r.PathPrefix("/static").Handler(http.FileServer(http.FS(static)))

	// metrics
	r.Handle("/metrics", promhttp.Handler())

	// probes
	r.HandleFunc("/live", NewOkHandler())
	r.HandleFunc("/ready", NewOkHandler())

	log.Println("start plugin center api on port", configuration.Port)

	err = http.ListenAndServe(":"+strconv.Itoa(configuration.Port), r)
	if err != nil {
		log.Fatal("http server returned err: ", err)
	}
}

func NewOkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}
