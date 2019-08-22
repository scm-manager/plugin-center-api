package main

import (
	"strconv"
	"strings"
)

type version []string

func less(releases []Release) func(int, int) bool {
	return func(i1 int, i2 int) bool {
		return isLess(releases[i1].Version, releases[i2].Version)
	}
}

func isLess(version1 string, version2 string) bool {
	parsedVersion1 := parse(version1)
	parsedVersion2 := parse(version2)

	var comparisonLength int
	if len(parsedVersion1) < len(parsedVersion2) {
		comparisonLength = len(parsedVersion1)
	} else {
		comparisonLength = len(parsedVersion2)
	}

	for i := 0; i < comparisonLength; i++ {
		if areNumbers(parsedVersion1[i], parsedVersion2[i]) {
			n1, _ := strconv.Atoi(parsedVersion1[i])
			n2, _ := strconv.Atoi(parsedVersion2[i])
			if n1 < n2 {
				return true
			} else if n1 > n2 {
				return false
			}
		} else if parsedVersion1[i] < parsedVersion2[i] {
			return true
		} else if parsedVersion1[i] > parsedVersion2[i] {
			return false
		}
	}

	return len(parsedVersion1) < len(parsedVersion2)
}

func areNumbers(strings ...string) bool {
	for _, s := range strings {
		_, err := strconv.Atoi(s)
		if err != nil {
			return false
		}
	}
	return true
}

func parse(versionString string) version {
	return strings.Split(versionString, ".")
}
