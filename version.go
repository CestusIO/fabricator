package fabricator

import (
	_ "embed"

	"code.cestus.io/tools/fabricator/pkg/genericclioptions"
)

//go:embed version.yml
var version string

func init() {
	genericclioptions.SetupVersion(GetVersionYaml(), "fabricator")
}

func GetVersionYaml() []byte {
	return []byte(version)
}
