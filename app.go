package main

import (
	"embed"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
)

//go:embed html
var assets embed.FS

func main() {
	configuration := readConfiguration()
	r := configureRouter(configuration)

	log.Println("start plugin center api on port", configuration.Port)
	err := http.ListenAndServe(getListenerAddress(configuration.Port), r)
	if err != nil {
		log.Fatal("http server returned err: ", err)
	}
}

func getListenerAddress(port int) string {
	if os.Getenv("STAGE") == "development" {
		return "127.0.0.1:" + strconv.Itoa(port)
	}
	return ":" + strconv.Itoa(port)
}

func configureRouter(configuration Configuration) *mux.Router {
	plugins, err := scanDirectory(configuration.DescriptorDirectory)
	if err != nil {
		log.Fatalln("could not parse plugins", err)
	}

	static, err := fs.Sub(assets, "html")
	if err != nil {
		log.Fatal("failed to load static files", err)
	}

	r := mux.NewRouter()

	authentication := func(handler http.Handler) http.Handler {
		return handler
	}

	// oidc
	if configuration.Oidc.IsEnabled() {
		oidc, err := NewOIDCHandler(configuration.Oidc, static)
		if err != nil {
			log.Fatal(err)
		}

		authentication = oidc.WithIdToken

		r.HandleFunc("/api/v1/auth/oidc", oidc.Authenticate)
		r.HandleFunc("/api/v1/auth/oidc/callback", oidc.Callback)
		r.HandleFunc("/api/v1/auth/oidc/refresh", oidc.Refresh)
	} else {
		log.Println("plugin center api starts without authentication support")
	}

	// api
	r.Handle("/api/v1/plugins/{version}", authentication(NewPluginHandler(plugins)))
	r.Handle("/api/v1/download/{plugin}/{version}", authentication(NewDownloadHandler(plugins)))

	// static assets
	r.PathPrefix("/static").Handler(http.FileServer(http.FS(static)))

	// metrics
	r.Handle("/metrics", promhttp.Handler())

	// probes
	r.HandleFunc("/live", NewOkHandler())
	r.HandleFunc("/ready", NewOkHandler())

	return r
}

func NewOkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}
