package genericclioptions

import (
	"fmt"
	"os/user"
	"runtime"
	"strings"
	"time"

	"code.cestus.io/tools/fabricator"
	"github.com/Scardiecat/svermaker"
	"github.com/Scardiecat/svermaker/semver"
)

// BuildInfo contains information about the build
type BuildInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Platform  string `json:"platform"`
}

// set with ldFlags
var (
	name      string
	version   string
	buildDate string
)

var buildInfo BuildInfo

func init() {
	if len(version) == 0 {
		version, name, buildDate = Generate()
	}
	if len(version) == 0 {
		version = localBuild()
	}
	buildInfo = BuildInfo{
		Name:      name,
		Version:   version,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Platform:  runtime.GOARCH,
	}
}

func localBuild() string {
	return fmt.Sprintf("%s-localbuild", getDeveloperName())
}
func getDeveloperName() string {
	usr, err := user.Current()

	if err != nil {
		return "unknown"
	}
	// On Windows the username may contain the domain, which we won't want.
	if strings.Contains(usr.Username, `\`) {
		parts := strings.Split(usr.Username, `\`)
		usr.Username = parts[len(parts)-1]
	}

	return usr.Username
}

func GetVersion() BuildInfo {
	return buildInfo
}

func Generate() (version string, name string, date string) {
	var serializer = fabricator.NewSerializer()
	var pvs = semver.ProjectVersionService{Serializer: serializer}
	var meta []string
	v, err := pvs.Get()
	if err != nil {
		return
	}
	version, err = MakeTags(*v, meta)
	if err != nil {
		return
	}
	name = "fabricator"
	date = time.Now().UTC().Format(time.RFC3339)
	return
}
func MakeTags(p svermaker.ProjectVersion, buildMetadata []string) (string, error) {
	m := semver.Manipulator{}

	isRelease := m.Compare(p.Current, p.Next) == 0
	c := p.Current
	if !isRelease {
		md, err := m.SetMetadata(c, buildMetadata)
		if err != nil {
			return "", err
		}
		c = md
		pre := c.Pre
		pre = append(pre, svermaker.PRVersion{VersionStr: localBuild(), VersionNum: 0, IsNum: false})
		c.Pre = pre

	}
	return c.String(), nil
}
