package genericclioptions

import (
	"errors"
	"fmt"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/Scardiecat/svermaker"
	"github.com/Scardiecat/svermaker/semver"
	"gopkg.in/yaml.v3"
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

func SetupVersion(versionyaml []byte, appName string) {
	if len(version) == 0 {
		version, name, buildDate = generate(versionyaml, appName)
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

func generate(versionyaml []byte, appName string) (version string, name string, date string) {
	var serializer = newSerializer(versionyaml)
	var pvs = semver.ProjectVersionService{Serializer: serializer}
	var meta []string
	v, err := pvs.Get()
	if err != nil {
		return
	}
	version, err = buildVersionString(*v, meta)
	if err != nil {
		return
	}
	name = appName
	date = time.Now().UTC().Format(time.RFC3339)
	return
}
func buildVersionString(p svermaker.ProjectVersion, buildMetadata []string) (string, error) {
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

// serializer implements the Serializer interface
type serializer struct {
	versionyaml []byte
	// Services
	projectVersionService semver.ProjectVersionService
}

// newSerializer creates a Serializer
func newSerializer(versionyaml []byte) *serializer {
	s := &serializer{
		versionyaml: versionyaml,
	}
	s.projectVersionService.Serializer = s
	return s
}

// Exists checks if the file exists
func (s *serializer) Exists() bool {
	return len(s.versionyaml) > 0
}

type projectVersion struct {
	Current string
	Next    string
}

// Serialize writes the ProjectVersion to a yml
func (s *serializer) Serialize(p svermaker.ProjectVersion) error {
	return errors.New("not implemented")
}

// Deserialize reads a projectcersion from a yml
func (s *serializer) Deserialize() (*svermaker.ProjectVersion, error) {
	v := projectVersion{}
	m := semver.Manipulator{}
	projectVersion := svermaker.ProjectVersion{}
	if s.Exists() {
		if err := yaml.Unmarshal(s.versionyaml, &v); err == nil {

		} else {
			return nil, err
		}
	} else {
		return nil, errors.New("version.yaml does not exist")
	}

	if current, err := m.Create(v.Current); err == nil {
		projectVersion.Current = *current
	} else {
		return nil, err
	}

	if next, err := m.Create(v.Next); err == nil {
		projectVersion.Next = *next
	} else {
		return nil, err
	}

	return &projectVersion, nil
}
