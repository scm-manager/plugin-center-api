package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	configuration := readConfiguration()

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
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
