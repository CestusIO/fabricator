package help

import (
	"errors"

	"code.cestus.io/tools/fabricator"
	"code.cestus.io/tools/fabricator/internal/pkg/util"
	"github.com/spf13/cobra"
)

type options struct {
	fabricator.IOStreams
	cmd *cobra.Command
}

// NewOptions returns initialized Options
func NewOptions(ioStreams fabricator.IOStreams) *options {
	return &options{
		IOStreams: ioStreams,
	}
}

// Run executes help command of the root command
func (o *options) Run() error {
	if o.cmd.Root() == nil {
		return errors.New("not a subcommand")
	}
	return o.cmd.Root().Help()
}

// NewHelpCommand creaters a new help command
func NewHelpCommand(ioStreams fabricator.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "help",
		Short:   "prints help",
		Long:    "prints help",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}
	o.cmd = cmd
	return cmd
}
