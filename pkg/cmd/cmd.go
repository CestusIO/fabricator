package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"code.cestus.io/tools/fabricator/pkg/cmd/help"
	"code.cestus.io/tools/fabricator/pkg/cmd/plugin"
	"code.cestus.io/tools/fabricator/pkg/cmd/version"
	"code.cestus.io/tools/fabricator/pkg/fabricator"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewDefaultFabricatorCommand creates the `fabricator` command with default arguments
func NewDefaultFabricatorCommand(ctx context.Context, io fabricator.IOStreams, flagParser fabricator.FlagParser) *cobra.Command {
	return NewDefaultFabricatorCommandWithArgs(ctx, NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes, io), os.Args, io, flagParser)
}

// NewDefaultFabricatorCommandWithArgs creates the `fabricator` command with arguments
func NewDefaultFabricatorCommandWithArgs(ctx context.Context, pluginHandler plugin.PluginHandler, args []string, io fabricator.IOStreams, flagparser fabricator.FlagParser) *cobra.Command {
	cmd := NewFabricatorCommand(io, flagparser)

	if pluginHandler == nil {
		return cmd
	}

	if len(args) > 1 {
		cmdPathPieces := args[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if _, _, err := cmd.Find(cmdPathPieces); err != nil {
			cmd.AddCommand(plugin.NewPluginWrapper(ctx, io, pluginHandler, flagparser, cmdPathPieces))
		}
	}

	return cmd
}

type options struct {
	fabricator.RootOptions
	fabricator.IOStreams
}

// NewOptions returns initialized Options
func NewOptions(ioStreams fabricator.IOStreams, flagset *pflag.FlagSet) *options {
	o := options{
		IOStreams: ioStreams,
	}
	o.RootOptions.RegisterOptions(flagset)
	return &o
}

// NewFabricatorCommand creates the `fabricator` command and its nested children.
func NewFabricatorCommand(io fabricator.IOStreams, flagparser fabricator.FlagParser) *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "fabricator",
		Short: "fabricator is the swiss army knive for code generation",
		Long:  `"fabricator is the swiss army knive for code generation"`,
		Run:   runHelp,
		// Hook before and after Run
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return nil
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return nil
		},
	}

	flags := cmds.PersistentFlags()
	_ = NewOptions(io, cmds.LocalFlags())
	flags.SetNormalizeFunc(WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(WordSepNormalizeFunc)

	//cmds.PersistentFlags().AddGoFlagSet(flagset)

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(WarnWordSepNormalizeFunc)

	cmds.AddCommand(plugin.NewCmdPlugin(io, flagparser))
	cmds.AddCommand(version.NewCmdVersion(io))
	help := help.NewHelpCommand(io)
	cmds.AddCommand(help)
	cmds.SetHelpCommand(help)

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// WarnWordSepNormalizeFunc changes and warns for flags that contain "_" separators
func WarnWordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		nname := strings.Replace(name, "_", "-", -1)
		if _, alreadyWarned := underscoreWarnings[name]; !alreadyWarned {
			fmt.Printf("using an underscore in a flag name is not supported. %s has been converted to %s.", name, nname)
			underscoreWarnings[name] = true
		}

		return pflag.NormalizedName(nname)
	}
	return pflag.NormalizedName(name)
}

// WordSepNormalizeFunc changes all flags that contain "_" separators
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}

var underscoreWarnings = make(map[string]bool)
