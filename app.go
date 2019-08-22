package main

import (
	"github.com/gorilla/mux"
)

func main() {
	configuration := readConfiguration()

	plugins, err := scanDirectory(configuration.DescriptorDirectory)

	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.Handle("/api/v1/plugins", NewPluginHandler(plugins))
}
