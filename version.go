package fabricator

import (
	"errors"

	_ "embed"

	"github.com/Scardiecat/svermaker"
	"github.com/Scardiecat/svermaker/semver"
	"gopkg.in/yaml.v3"
)

//go:embed version.yml
var version string

// Serializer implements the Serializer interface
type Serializer struct {
	// Services
	projectVersionService semver.ProjectVersionService
}

// NewSerializer creates a Serializer
func NewSerializer() *Serializer {
	s := &Serializer{}
	s.projectVersionService.Serializer = s
	return s
}

// Exists checks if the file exists
func (s *Serializer) Exists() bool {
	return len(version) > 0
}

type projectVersion struct {
	Current string
	Next    string
}

// Serialize writes the ProjectVersion to a yml
func (s *Serializer) Serialize(p svermaker.ProjectVersion) error {
	return errors.New("not implemented")
}

// Deserialize reads a projectcersion from a yml
func (s *Serializer) Deserialize() (*svermaker.ProjectVersion, error) {
	v := projectVersion{}
	m := semver.Manipulator{}
	projectVersion := svermaker.ProjectVersion{}
	if s.Exists() {
		if err := yaml.Unmarshal([]byte(version), &v); err == nil {

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
