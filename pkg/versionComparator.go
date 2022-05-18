package pkg

import (
	"github.com/hashicorp/go-version"
	"log"
)

func less(releases []Release) func(int, int) bool {
	return func(i1 int, i2 int) bool {
		versionString1 := releases[i1].Version
		versionString2 := releases[i2].Version
		return isLess(versionString1, versionString2)
	}
}

func isLess(versionString1 string, versionString2 string) bool {
	v1, err1 := version.NewVersion(versionString1)
	v2, err2 := version.NewVersion(versionString2)
	if err1 != nil || err2 != nil {
		log.Println("cannot compare versions by semantic versioning: ", versionString1, ",", versionString2, "- Falling back to string compare")
		return versionString1 < versionString2
	}
	return v1.LessThan(v2)
}
