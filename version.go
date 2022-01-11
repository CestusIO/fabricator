package fabricator

import (
	_ "embed"

	"code.cestus.io/libs/buildinfo"
)

//go:embed version.yml
var version string

func init() {
	buildinfo.GenerateVersionFromVersionYaml(GetVersionYaml(), "fabricator")
}

func GetVersionYaml() []byte {
	return []byte(version)
}
