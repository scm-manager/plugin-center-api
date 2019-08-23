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

type RequestConditions struct {
	Os         string
	Arch       string
	MinVersion string
}

func NewPluginHandler(plugins []Plugin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pluginResults []PluginResult

		requestConditions, err := extractRequestConditions(r)
		if err != nil {
			log.Println("could not parse form data for request", err)
			w.WriteHeader(400)
			w.Write([]byte("could not parse form data for request"))
			return
		}

		for _, plugin := range plugins {
			pluginResults = appendIfOk(pluginResults, plugin, requestConditions)
		}

		embedded := make(map[string]PluginResults)
		embedded["plugins"] = pluginResults
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

func extractRequestConditions(r *http.Request) (RequestConditions, error) {
	err := r.ParseForm()
	if err != nil {
		return RequestConditions{}, err
	}
	queryParameters := r.Form
	requestConditions := RequestConditions{
		Os:         queryParameters.Get("os"),
		Arch:       "",
		MinVersion: "",
	}
	return requestConditions, nil
}

func appendIfOk(results []PluginResult, plugin Plugin, conditions RequestConditions) []PluginResult {
	if len(plugin.Releases) > 0 {
		result := PluginResult{
			Name:        plugin.Name,
			DisplayName: plugin.DisplayName,
			Description: plugin.Description,
			Category:    plugin.Category,
			Version:     plugin.Releases[0].Version,
			Author:      plugin.Author,
			Conditions:  extractConditions(plugin.Releases[0].Conditions),
			Links:       nil,
		}
		return append(results, result)
	} else {
		return results
	}
}

func extractConditions(conditions Conditions) ConditionMap {
	conditionMap := make(map[string]string)
	if conditions.Os != "" {
		conditionMap["os"] = conditions.Os
	}
	if conditions.Arch != "" {
		conditionMap["arch"] = conditions.Arch
	}
	if conditions.MinVersion != "" {
		conditionMap["minVersion"] = conditions.MinVersion
	}
	return conditionMap
}
