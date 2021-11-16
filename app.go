package main

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
)

func main() {
	configuration := readConfiguration()

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	if err != nil {
		log.Fatalln("could not parse plugins", err)
	}
	oidc := NewOIDCHandler()

	r := mux.NewRouter()
	r.Handle("/api/v1/plugins/{version}", oidc.Authenticate(NewPluginHandler(plugins)))
	r.Handle("/api/v1/download/{plugin}/{version}", oidc.Authenticate(NewDownloadHandler(plugins)))
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/live", NewOkHandler())
	r.HandleFunc("/ready", NewOkHandler())
	r.HandleFunc("/oidc", oidc.Handler)
	r.HandleFunc("/oidc/callback", oidc.Callback)
	r.HandleFunc("/oidc/refresh", oidc.Refresh)

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
