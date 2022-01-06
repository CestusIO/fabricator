package helpers

import (
	"os"

	"code.cestus.io/tools/fabricator/pkg/fabricator"
	"code.cestus.io/tools/fabricator/pkg/ff"
	"code.cestus.io/tools/fabricator/pkg/ff/ffpflag"
	"github.com/spf13/cobra"
)

var DefaultFlagParser fabricator.FlagParser = func(cmd *cobra.Command) error {
	flagset := ffpflag.NewFlagSet(cmd.Flags())
	return ff.Parse(flagset, os.Args[1:],
		ff.WithEnvVarPrefix("fabricator"),
	)
}
