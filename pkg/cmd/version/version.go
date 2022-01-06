package version

import (
	"fmt"

	"code.cestus.io/tools/fabricator/internal/pkg/util"
	"code.cestus.io/tools/fabricator/pkg/fabricator"
	"code.cestus.io/tools/fabricator/pkg/genericclioptions"
	"github.com/spf13/cobra"
)

type options struct {
	fabricator.IOStreams
}

// NewOptions returns initialized Options
func NewOptions(ioStreams fabricator.IOStreams) *options {
	return &options{
		IOStreams: ioStreams,
	}
}

// Run executes version command
func (o *options) Run() error {
	version := genericclioptions.GetVersion()
	fmt.Fprintf(o.Out, "Name:       %s\n", version.Name)
	fmt.Fprintf(o.Out, "Version:    %s\n", version.Version)
	fmt.Fprintf(o.Out, "BuildDate:  %s\n", version.BuildDate)
	fmt.Fprintf(o.Out, "Go-Version: %s\n", version.GoVersion)
	fmt.Fprintf(o.Out, "Platform:   %s\n", version.Platform)
	fmt.Fprintf(o.Out, "OS:         %s\n", version.OS)
	return nil
}

func NewCmdVersion(ioStreams fabricator.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the version",
		Long:    "Print the version",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}
	return cmd
}
