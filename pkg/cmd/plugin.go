package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code.cestus.io/tools/fabricator"
	"code.cestus.io/tools/fabricator/pkg/helpers"
)

// DefaultPluginHandler implements PluginHandler
type DefaultPluginHandler struct {
	ValidPrefixes []string
	IO            fabricator.IOStreams
}

// NewDefaultPluginHandler instantiates the DefaultPluginHandler with a list of
// given filename prefixes used to identify valid plugin filenames.
func NewDefaultPluginHandler(validPrefixes []string, io fabricator.IOStreams) *DefaultPluginHandler {
	return &DefaultPluginHandler{
		ValidPrefixes: validPrefixes,
		IO:            io,
	}
}

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		fileExt := strings.ToLower(filepath.Ext(file))

		switch fileExt {
		case ".bat", ".cmd", ".com", ".exe", ".ps1":
			return nil
		}
		return errors.New("not an executable")
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return errors.New("not an executable")
}

// Lookup implements PluginHandler
func (h *DefaultPluginHandler) Lookup(ctx context.Context, filename string, paths []string) (string, bool) {
	for _, prefix := range h.ValidPrefixes {
		for _, dir := range paths {
			if dir == "" {
				// Unix shell semantics: path element "" means "."
				dir = "."
			}
			path := filepath.Join(dir, fmt.Sprintf("%s-%s", prefix, filename))
			if err := findExecutable(path); err == nil {
				return path, true
			}
			if runtime.GOOS == "windows" {
				ext := []string{".bat", ".cmd", ".com", ".exe", ".ps1"}
				for _, e := range ext {
					np := fmt.Sprintf("%s%s", path, e)
					if err := findExecutable(np); err == nil {
						return path, true
					}
				}
			}
		}
	}
	return "", false
}

// Execute implements PluginHandler
func (h *DefaultPluginHandler) Execute(ctx context.Context, executablePath string, cmdArgs []string, environment fabricator.Environment) error {
	executor := helpers.NewExecutor("", h.IO).WithEnvMap(environment)
	err := executor.Run(ctx, executablePath, cmdArgs...)
	return err
}
