package plugin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code.cestus.io/libs/buildinfo"
	"code.cestus.io/tools/fabricator/internal/pkg/util"
	"code.cestus.io/tools/fabricator/pkg/fabricator"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	pluginLong = `
		Provides utilities for interacting with plugins.
		Plugins provide extended functionality that is not part of the fabricator base feature set`

	pluginListLong = `
		List all available plugin files on a user's PATH.

		Available plugin files are those that are:
		- executable
		- anywhere on the user's PATH (or in the )
		- begin with "fabricator-"`

	ValidPluginFilenamePrefixes = []string{"fabricator"}
)

// PluginHandler is capable of parsing command line arguments
// and performing executable filename lookups to search
// for valid plugin files, and execute found plugins.
type PluginHandler interface {
	// exists at the given filename, or a boolean false.
	// Lookup will iterate over a list of given prefixes
	// in order to recognize valid plugin filenames.
	// The first filepath to match a prefix is returned.
	Lookup(ctx context.Context, filename string, paths []string) (string, bool)
	// Execute receives an executable's filepath, a slice
	// of arguments, and a slice of environment variables
	// to relay to the executable.
	Execute(ctx context.Context, executablePath string, cmdArgs []string, environment fabricator.Environment) error
}

func NewCmdPlugin(streams fabricator.IOStreams, flagparser fabricator.FlagParser) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "plugin [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "Provides utilities for interacting with plugins.",
		Long:                  pluginLong,
		Run: func(cmd *cobra.Command, args []string) {
			util.DefaultSubCommandRun(streams.ErrOut)(cmd, args)
		},
	}

	cmd.AddCommand(NewCmdPluginList(streams, flagparser))
	return cmd
}

type Options struct {
	fabricator.RootOptions
	fabricator.IOStreams
	Verifier PathVerifier
	NameOnly bool

	PluginPaths []string
}

// NewOptions returns initialized Options
func NewOptions(ioStreams fabricator.IOStreams, flagset *pflag.FlagSet, flagparser fabricator.FlagParser) *Options {
	o := Options{
		IOStreams: ioStreams,
	}
	o.RootOptions.FlagParser = flagparser
	o.RootOptions.RegisterOptions(flagset)
	return &o
}

// NewCmdPluginList provides a way to list all plugin executables visible to fabricator
func NewCmdPluginList(streams fabricator.IOStreams, flagparser fabricator.FlagParser) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all visible plugin executables on a user's PATH",
		Long:  pluginListLong,
	}
	o := NewOptions(streams, cmd.Flags(), flagparser)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		util.CheckErr(o.Complete(cmd))
		util.CheckErr(o.Run())
	}
	cmd.Flags().BoolVar(&o.NameOnly, "name-only", o.NameOnly, "If true, display only the binary name of each plugin, rather than its full path")
	return cmd
}

func (o *Options) Complete(cmd *cobra.Command) error {
	err := o.FlagParser(cmd)
	if err != nil {
		return err
	}
	o.Verifier = &CommandOverrideVerifier{
		root:        cmd.Root(),
		seenPlugins: make(map[string]string),
	}

	o.PluginPaths = filepath.SplitList(o.PluginPath)
	o.PluginPaths = append(o.PluginPaths, filepath.SplitList(os.Getenv("PATH"))...)
	return nil
}

func (o *Options) Run() error {
	pluginsFound := false
	isFirstFile := true
	pluginErrors := []error{}
	pluginWarnings := 0

	for _, dir := range uniquePathsList(o.PluginPaths) {
		if len(strings.TrimSpace(dir)) == 0 {
			continue
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			if _, ok := err.(*os.PathError); ok {
				fmt.Fprintf(o.ErrOut, "Unable to read directory %q from your PATH: %v. Skipping...\n", dir, err)
				continue
			}

			pluginErrors = append(pluginErrors, fmt.Errorf("error: unable to read directory %q in your PATH: %v", dir, err))
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !hasValidPrefix(f.Name(), ValidPluginFilenamePrefixes) {
				continue
			}
			if runtime.GOOS != "windows" {
				// filter out windows executables
				fileExt := strings.ToLower(filepath.Ext(f.Name()))

				switch fileExt {
				case ".bat", ".cmd", ".com", ".exe", ".ps1":
					continue
				}
			}

			if isFirstFile {
				fmt.Fprintf(o.Out, "The following compatible plugins are available:\n\n")
				pluginsFound = true
				isFirstFile = false
			}

			pluginPath := f.Name()
			if !o.NameOnly {
				pluginPath = filepath.Join(dir, pluginPath)
			}

			fmt.Fprintf(o.Out, "%s\n", pluginPath)
			if errs := o.Verifier.Verify(filepath.Join(dir, f.Name())); len(errs) != 0 {
				for _, err := range errs {
					fmt.Fprintf(o.ErrOut, "  - %s\n", err)
					pluginWarnings++
				}
			}
		}
	}

	if !pluginsFound {
		pluginErrors = append(pluginErrors, fmt.Errorf("error: unable to find any fabricator plugins in your PATH"))
	}

	if pluginWarnings > 0 {
		if pluginWarnings == 1 {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: one plugin warning was found"))
		} else {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: %v plugin warnings were found", pluginWarnings))
		}
	}
	if len(pluginErrors) > 0 {
		errs := bytes.NewBuffer(nil)
		for _, e := range pluginErrors {
			fmt.Fprintln(errs, e)
		}
		return fmt.Errorf("%s", errs.String())
	}

	return nil
}

// pathVerifier receives a path and determines if it is valid or not
type PathVerifier interface {
	// Verify determines if a given path is valid
	Verify(path string) []error
}

type CommandOverrideVerifier struct {
	root        *cobra.Command
	seenPlugins map[string]string
}

// Verify implements PathVerifier and determines if a given path
// is valid depending on whether or not it overwrites an existing
// fabricator command path, or a previously seen plugin.
func (v *CommandOverrideVerifier) Verify(path string) []error {
	if v.root == nil {
		return []error{fmt.Errorf("unable to verify path with nil root")}
	}

	// extract the plugin binary name
	segs := strings.Split(path, "/")
	binName := segs[len(segs)-1]

	cmdPath := strings.Split(binName, "-")
	if len(cmdPath) > 1 {
		// the first argument is always "fabricator" for a plugin binary
		cmdPath = cmdPath[1:]
	}

	errors := []error{}

	if isExec, err := isExecutable(path); err == nil && !isExec {
		errors = append(errors, fmt.Errorf("warning: %s identified as a fabricator plugin, but it is not executable", path))
	} else if err != nil {
		errors = append(errors, fmt.Errorf("error: unable to identify %s as an executable file: %v", path, err))
	}

	if existingPath, ok := v.seenPlugins[binName]; ok {
		errors = append(errors, fmt.Errorf("warning: %s is overshadowed by a similarly named plugin: %s", path, existingPath))
	} else {
		v.seenPlugins[binName] = path
	}

	if cmd, _, err := v.root.Find(cmdPath); err == nil {
		errors = append(errors, fmt.Errorf("warning: %s overwrites existing command: %q", binName, cmd.CommandPath()))
	}

	return errors
}

func isExecutable(fullPath string) (bool, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}

	if runtime.GOOS == "windows" {
		fileExt := strings.ToLower(filepath.Ext(fullPath))

		switch fileExt {
		case ".bat", ".cmd", ".com", ".exe", ".ps1":
			return true, nil
		}
		return false, nil
	}

	if m := info.Mode(); !m.IsDir() && m&0111 != 0 {
		return true, nil
	}

	return false, nil
}

// uniquePathsList deduplicates a given slice of strings without
// sorting or otherwise altering its order in any way.
func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	newPaths := []string{}
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
}

func hasValidPrefix(filepath string, validPrefixes []string) bool {
	for _, prefix := range validPrefixes {
		if !strings.HasPrefix(filepath, prefix+"-") {
			continue
		}
		return true
	}
	return false
}

func NewPluginWrapper(ctx context.Context, streams fabricator.IOStreams, handler PluginHandler, flagparser fabricator.FlagParser, pluginPathPieces []string) *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		Short:                 "execute plugin",
		Long:                  "execute plugin",
	}
	// This cannot use the standard way of loading options, because we are trying to execute a plugin. But we will need to read the standard options to see if we have to use a plugin path.
	o := NewOptions(streams, cmd.Flags(), flagparser)
	// Most of the time FlagParser will complain about additional flags (since it cannot know what the plugin needs) so we ignore it here. The plugin is responsible to process the flags
	o.FlagParser(cmd)

	o.PluginPaths = filepath.SplitList(o.PluginPath)
	o.PluginPaths = append(o.PluginPaths, filepath.SplitList(os.Getenv("PATH"))...)
	fun, name, err := pluginCommandHandler(ctx, handler, pluginPathPieces, o.PluginPaths)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err != nil {
			return err
		}
		return fun()
	}
	cmd.Use = name
	return cmd
}

func pluginCommandHandler(ctx context.Context, pluginHandler PluginHandler, cmdArgs []string, paths []string) (func() error, string, error) {
	var remainingArgs []string // all "non-flag" arguments
	name := ""
	if len(cmdArgs) > 0 {
		name = cmdArgs[0]
	}
	for _, arg := range cmdArgs {
		if strings.HasPrefix(arg, "-") {
			break
		}
		remainingArgs = append(remainingArgs, strings.Replace(arg, "-", "_", -1))
	}

	if len(remainingArgs) == 0 {
		// the length of cmdArgs is at least 1
		err := fmt.Errorf("flags cannot be placed before plugin name: %s", cmdArgs[0])
		return func() error { return err }, name, err
	}

	foundBinaryPath := ""

	// attempt to find binary, starting at longest possible name with given cmdArgs
	for len(remainingArgs) > 0 {
		tentativeName := strings.Join(remainingArgs, "-")
		// try extended name with GOOS-ARCH first so developpers can work with their locally build plugins
		path, found := pluginHandler.Lookup(ctx, strings.Join([]string{tentativeName, buildinfo.ProvideBuildInfo().OS, buildinfo.ProvideBuildInfo().Platform}, "-"), paths)

		if !found {
			path, found = pluginHandler.Lookup(ctx, tentativeName, paths)
		}
		if !found {
			remainingArgs = remainingArgs[:len(remainingArgs)-1]
			continue
		}

		foundBinaryPath = path
		break
	}

	if len(foundBinaryPath) == 0 {
		err := errors.New("no plugin found")
		return func() error { return err }, name, err
	}

	exec := func() error {
		// invoke cmd binary relaying the current environment and args given
		if err := pluginHandler.Execute(ctx, foundBinaryPath, cmdArgs[len(remainingArgs):], make(fabricator.Environment)); err != nil {
			return err
		}

		return nil
	}

	return exec, name, nil
}

// HandlePluginCommand receives a pluginHandler and command-line arguments and attempts to find
// a plugin executable on the PATH that satisfies the given arguments.
func HandlePluginCommand(ctx context.Context, pluginHandler PluginHandler, cmdArgs []string, paths []string) error {
	var remainingArgs []string // all "non-flag" arguments
	for _, arg := range cmdArgs {
		if strings.HasPrefix(arg, "-") {
			break
		}
		remainingArgs = append(remainingArgs, strings.Replace(arg, "-", "_", -1))
	}

	if len(remainingArgs) == 0 {
		// the length of cmdArgs is at least 1
		return fmt.Errorf("flags cannot be placed before plugin name: %s", cmdArgs[0])
	}

	foundBinaryPath := ""

	// attempt to find binary, starting at longest possible name with given cmdArgs
	for len(remainingArgs) > 0 {
		path, found := pluginHandler.Lookup(ctx, strings.Join(remainingArgs, "-"), paths)
		if !found {
			remainingArgs = remainingArgs[:len(remainingArgs)-1]
			continue
		}

		foundBinaryPath = path
		break
	}

	if len(foundBinaryPath) == 0 {
		return nil
	}

	// invoke cmd binary relaying the current environment and args given
	if err := pluginHandler.Execute(ctx, foundBinaryPath, cmdArgs[len(remainingArgs):], make(fabricator.Environment)); err != nil {
		return err
	}

	return nil
}
