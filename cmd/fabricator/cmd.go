package main

import (
	"context"
	"os"

	"code.cestus.io/tools/fabricator/pkg/cmd"
	"code.cestus.io/tools/fabricator/pkg/fabricator"
	"code.cestus.io/tools/fabricator/pkg/helpers"
)

func main() {
	ctx := context.Background()

	io := fabricator.NewStdIOStreams()
	ctx, cancel := helpers.WithCancelOnSignal(ctx, io, fabricator.TerminationSignals...)
	defer cancel()
	rootCmd := cmd.NewDefaultFabricatorCommand(ctx, io, helpers.DefaultFlagParser)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
