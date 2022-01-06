package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"code.cestus.io/tools/fabricator/internal/pkg/util"
	"code.cestus.io/tools/fabricator/pkg/fabricator"
	"code.cestus.io/tools/fabricator/pkg/helpers"
)

func TestFabricatorCommandHandlesPlugins(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectPlugin     string
		expectPluginArgs []string
		expectError      string
	}{
		{
			name:             "test that a plugin executable is found based on command args",
			args:             []string{"fabricator", "foo", "--bar"},
			expectPlugin:     "plugin/testdata/fabricator-foo",
			expectPluginArgs: []string{"--bar"},
		},
		{
			name: "test that a plugin does not execute over an existing command by the same name",
			args: []string{"fabricator", "version"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			pluginsHandler := &testPluginHandler{
				pluginsDirectory: "plugin/testdata",
			}
			io, _, _, errOut := fabricator.NewTestIOStreams()

			util.BehaviorOnFatal(func(str string, code int) {
				errOut.Write([]byte(str))
			})

			root := NewDefaultFabricatorCommandWithArgs(ctx, pluginsHandler, test.args, io, helpers.DefaultFlagParser)

			root.SetArgs(test.args[1:])
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if pluginsHandler.err != nil && pluginsHandler.err.Error() != test.expectError {
				t.Fatalf("unexpected error: expected %q to occur, but got %q", test.expectError, pluginsHandler.err)
			}

			if pluginsHandler.executedPlugin != test.expectPlugin {
				t.Fatalf("unexpected plugin execution: expected %q, got %q", test.expectPlugin, pluginsHandler.executedPlugin)
			}

			if len(pluginsHandler.withArgs) != len(test.expectPluginArgs) {
				t.Fatalf("unexpected plugin execution args: expected %q, got %q", test.expectPluginArgs, pluginsHandler.withArgs)
			}
		})
	}
}

type testPluginHandler struct {
	pluginsDirectory string

	// execution results
	executedPlugin string
	withArgs       []string
	withEnv        fabricator.Environment

	err error
}

func (h *testPluginHandler) Lookup(ctx context.Context, filename string, paths []string) (string, bool) {
	// append supported plugin prefix to the filename
	filename = fmt.Sprintf("%s-%s", "fabricator", filename)

	dir, err := os.Stat(h.pluginsDirectory)
	if err != nil {
		h.err = err
		return "", false
	}

	if !dir.IsDir() {
		h.err = fmt.Errorf("expected %q to be a directory", h.pluginsDirectory)
		return "", false
	}

	plugins, err := os.ReadDir(h.pluginsDirectory)
	if err != nil {
		h.err = err
		return "", false
	}

	for _, p := range plugins {
		if p.Name() == filename {
			h.err = nil
			return fmt.Sprintf("%s/%s", h.pluginsDirectory, p.Name()), true
		}
	}

	h.err = fmt.Errorf("unable to find a plugin executable %q", filename)
	return "", false
}

func (h *testPluginHandler) Execute(ctx context.Context, executablePath string, cmdArgs []string, env fabricator.Environment) error {
	h.executedPlugin = executablePath
	h.withArgs = cmdArgs
	h.withEnv = env
	return nil
}
