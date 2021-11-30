package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"net/http"
	"strconv"
)

type Links map[string]Link

type ConditionMap map[string]interface{}

type Link struct {
	Href string `json:"href"`
}

type PluginResult struct {
	Name                 string       `json:"name"`
	DisplayName          string       `json:"displayName"`
	Description          string       `json:"description"`
	Category             string       `json:"category"`
	Version              string       `json:"version"`
	Author               string       `json:"author"`
	Checksum             string       `json:"sha256sum"`
	Type                 string       `json:"type"`
	AvatarUrl            string       `json:"avatarUrl"`
	Conditions           ConditionMap `json:"conditions"`
	Dependencies         []string     `json:"dependencies"`
	OptionalDependencies []string     `json:"optionalDependencies"`
	Links                Links        `json:"_links"`
}

type PluginResults []PluginResult

type Embedded map[string]PluginResults

type Response struct {
	EmbeddedPlugins Embedded `json:"_embedded"`
}

type RequestConditions struct {
	Os      string
	Arch    string
	Jre     string
	Version version.Version
}

var (
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scm_plugin_center_api_requests",
		Help: "Total number of requests",
	}, []string{
		"version", "os", "arch", "jre", "authenticated",
	})
)

func NewPluginHandler(plugins []Plugin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pluginResults []PluginResult

		requestConditions, err := extractRequestConditions(r)
		if err != nil {
			log.Println("could not parse form data for request", err)
			http.Error(w, "could not parse form data for request", http.StatusBadRequest)
			return
		}

		log.Println("reading plugins for version", requestConditions.Version.Original())

		authenticated := r.Context().Value("idToken") != nil

		requestCounter.WithLabelValues(
			requestConditions.Version.String(),
			requestConditions.Os,
			requestConditions.Arch,
			requestConditions.Jre,
			strconv.FormatBool(authenticated),
		).Inc()

		urlGenerator := NewUrlGenerator(*r)

		for _, plugin := range plugins {
			pluginResults = appendIfOk(pluginResults, plugin, requestConditions, urlGenerator, authenticated)
		}

		embedded := make(map[string]PluginResults)
		embedded["plugins"] = pluginResults
		response := Response{EmbeddedPlugins: embedded}

		w.Header().Add("Content-Type", "application/json")

		data, err := json.Marshal(response)
		if err != nil {
			log.Println("could not marshal result for plugin call", err)
			http.Error(w, "failed to marshal response", http.StatusInternalServerError)
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
	vars := mux.Vars(r)
	requestVersion, err := version.NewVersion(vars["version"])
	if err != nil {
		return RequestConditions{}, err
	}
	requestConditions := RequestConditions{
		Os:      queryParameters.Get("os"),
		Arch:    queryParameters.Get("arch"),
		Jre:     queryParameters.Get("jre"),
		Version: *requestVersion,
	}
	return requestConditions, nil
}

func appendIfOk(results []PluginResult, plugin Plugin, conditions RequestConditions, generator UrlGenerator, authenticated bool) []PluginResult {
	for _, release := range plugin.Releases {
		if conditionsMatch(conditions, release.Conditions) {
			links := Links{}

			pluginType := plugin.Type
			if pluginType == "" {
				pluginType = "SCM"
			}
			if pluginType == "SCM" || authenticated {
				url := generator.DownloadUrl(plugin, release.Version)
				links["download"] = Link{Href: url}
			}
			if release.InstallLink != "" {
				links["install"] = Link{Href: release.InstallLink}
			}
			avatarUrl := plugin.AvatarUrl
			if avatarUrl != "" {
				avatarUrl = "https://scm-manager.org/img/" + avatarUrl
			}
			result := PluginResult{
				Name:                 plugin.Name,
				DisplayName:          plugin.DisplayName,
				Description:          plugin.Description,
				Category:             plugin.Category,
				Version:              release.Version,
				Author:               plugin.Author,
				Checksum:             release.Checksum,
				Type:                 pluginType,
				AvatarUrl:            avatarUrl,
				Conditions:           extractConditions(release.Conditions),
				Dependencies:         nullToEmpty(release.Dependencies),
				OptionalDependencies: nullToEmpty(release.OptionalDependencies),
				Links:                links,
			}
			return append(results, result)
		}
	}

	return results
}

func conditionsMatch(requestConditions RequestConditions, releaseConditions Conditions) bool {
	if len(releaseConditions.Os) > 0 && requestConditions.Os != "" {
		var contains = false
		for _, os := range releaseConditions.Os {
			if requestConditions.Os == os {
				contains = true
			}
		}
		if !contains {
			return false
		}
	}
	if releaseConditions.Arch != "" && requestConditions.Arch != "" && releaseConditions.Arch != requestConditions.Arch {
		return false
	}
	if releaseConditions.MinVersion == "" {
		return true
	}
	minVersion, err := version.NewVersion(releaseConditions.MinVersion)
	if err != nil {
		log.Println("could not parse version string", releaseConditions.MinVersion, "- ignoring release")
		return false
	}
	return requestConditions.Version.GreaterThanOrEqual(minVersion)
}

func extractConditions(conditions Conditions) ConditionMap {
	conditionMap := make(map[string]interface{})
	if len(conditions.Os) > 0 {
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

func nullToEmpty(strings []string) []string {
	if strings == nil {
		return []string{}
	} else {
		return strings
	}
}
