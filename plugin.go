package main

type Condition struct {
	name  string
	value string
}

type Release struct {
	version    string
	conditions []Condition
	url        string
	date       string
	checksum   string
}

type Plugin struct {
	name        string
	displayName string
	description string
	category    string
	Releases    []Release
	author      string
}
