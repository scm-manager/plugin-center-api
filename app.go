package main

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func main() {
	configuration := readConfiguration()

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	if err != nil {
		log.Fatalln("could not parse plugins", err)
	}

	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/live", NewOkHandler())
	r.HandleFunc("/ready", NewOkHandler())
	r.Handle("/api/v1/plugins/{version}", NewPluginHandler(plugins))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Println("start plugin center api on port", port)

	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal("http server returned err: ", err)
	}
}

func NewOkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}
