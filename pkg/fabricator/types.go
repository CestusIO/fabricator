package fabricator

import (
	"bytes"
	"io"
	"os"
	"syscall"

	"github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// TerminationSignals are signals that cause the program to exit in the
// supported platforms (linux, darwin, windows).
var TerminationSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}

// Environment is a key value map for environment variables
type Environment map[string]string

// IOStreams provides the standard names for iostreams.  This is useful for embedding and for unit testing.
// Inconsistent and different names make it hard to read and review code
type IOStreams struct {
	// In think, os.Stdin
	In io.Reader
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

// NewTestIOStreams returns a valid IOStreams and in, out, errout buffers for unit tests
func NewTestIOStreams() (IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	return IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

// NewTestIOStreamsDiscard returns a valid IOStreams that just discards
func NewTestIOStreamsDiscard() IOStreams {
	in := &bytes.Buffer{}
	return IOStreams{
		In:     in,
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
}

// NewGinkoTestIOStreams returns a valid IOStreams for use with ginkgotests
func NewGinkoTestIOStreams() IOStreams {
	in := &bytes.Buffer{}
	return IOStreams{
		In:     in,
		Out:    ginkgo.GinkgoWriter,
		ErrOut: ginkgo.GinkgoWriter,
	}
}

// NewStdIOStreams returns an IOStreams instance using os std streams
func NewStdIOStreams() IOStreams {
	return IOStreams{
		In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr,
	}
}

// OptionProvider is an interface for command options
type OptionProvider interface {
	RegisterOptions(flagset *pflag.FlagSet)
}

// RootOptions defines a common set of options for all plugins
type RootOptions struct {
	FabricatorFile string
	RootDirectory  string
	PluginPath     string
	Help           bool
	FlagParser     FlagParser
}

// RegisterOptions implements the OptionsProvider interface
func (o *RootOptions) RegisterOptions(flagset *pflag.FlagSet) {
	flagset.StringVar(&o.FabricatorFile, "fabfile", "./.fabricator.yml", "fab-file to load")
	flagset.StringVar(&o.RootDirectory, "rootdir", "./", "root directory for all file operations")
	flagset.StringVarP(&o.PluginPath, "plugin-path", "p", "./", "path extension where plugins will be loaded from")
	flagset.BoolP("help", "h", false, "Help for")
}

// FlagParser defines the signature for a function to parse commandline flags. It exists so that cobra's flag parsing can be less magical and give control over when and what is actually parsed
type FlagParser func(cmd *cobra.Command) error

// Typed for the fabricator config file

type FabricatorConfig struct {
	ApiVersion string               `yaml:"apiVersion" json:"apiVersion"`
	Kind       string               `yaml:"kind" json:"kind"`
	Components FabricatorComponents `yaml:"components" json:"components"`
}

type FabricatorComponent struct {
	Name      string    `yaml:"name" json:"name"`
	Generator string    `yaml:"generator" json:"generator"`
	Spec      yaml.Node `yaml:"spec" json:"spec"`
}

type FabricatorComponents []FabricatorComponent
