package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Links map[string]Link

type ConditionMap map[string]string

type Link struct {
	Href string `json:"href"`
}

type PluginResult struct {
	Name        string       `json:"name"`
	DisplayName string       `json:"displayName"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	Conditions  ConditionMap `json:"conditions"`
	Links       Links        `json:"_links"`
}

type PluginResults []PluginResult

type Embedded map[string]PluginResults

type Response struct {
	EmbeddedPlugins Embedded `json:"_embedded"`
}

func NewPluginHandler(plugins []Plugin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var plugins []PluginResult
		embedded := make(map[string]PluginResults)
		embedded["plugins"] = plugins
		response := Response{EmbeddedPlugins: embedded}

		w.Header().Add("Content-Type", "application/json")

		data, err := json.Marshal(response)
		if err != nil {
			log.Println("could not marshal result for plugin call", err)
			http.Error(w, "failed to marshal response", 500)
		}
		_, err = w.Write(data)
		if err != nil {
			log.Println("failed to write response", err)
		}
	}
}
