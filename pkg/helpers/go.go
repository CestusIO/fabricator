package helpers

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"code.cestus.io/tools/fabricator/pkg/fabricator"
)

// GuessGoImportPath guesses to Go import path for the specified folder.
func GuessGoImportPath(ctx context.Context, io fabricator.IOStreams, root string) (goImportPath string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("guessing Go import path at `%s`: %s", root, err)
		}
	}()

	executor := NewExecutor(root, io).WithEnv("GOWORK", "off")
	// If a Go import path is defined, use it.
	if goImportPath, err = executor.Output(ctx, "go", "list", "-f", "{{ .ImportPath }}"); err == nil {
		return strings.TrimSpace(goImportPath), nil
	}

	mod, err := GetGoModule(ctx, io, root)

	if err != nil {
		return "", err
	}

	return mod.GetRelativeImportPath(root)
}

type GoModule struct {
	Path string `json:"Path"`
	Dir  string `json:"Dir"`
}

func (m GoModule) String() string {
	return m.Path
}

func (m GoModule) GetRelativeImportPath(dir string) (_ string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getting relative import path for `%s`: %s", dir, err)
		}
	}()

	relPath, err := filepath.Rel(m.Dir, dir)

	if err != nil {
		return "", fmt.Errorf("unable to determine relative path: %s", err)
	}

	return path.Join(m.Path, filepath.ToSlash(relPath)), nil
}

func GetGoModule(ctx context.Context, io fabricator.IOStreams, root string) (_ *GoModule, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getting Go module at `%s`: %s", root, err)
		}
	}()

	executor := NewExecutor(root, io).WithEnv("GOWORK", "off")

	var result GoModule
	err = executor.JSONOutput(ctx, &result, "go", "list", "-m", "-json")

	return &result, nil
}

func GetGoPackage(ctx context.Context, io fabricator.IOStreams, pkg string) (_ *GoModule, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getting Go package for `%s`: %s", pkg, err)
		}
	}()

	wd, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("could not determine working directory")
	}

	executor := NewExecutor(wd, io).WithEnv("GOWORK", "off")
	var result GoModule
	err = executor.JSONOutput(ctx, &result, "go", "list", "-json", pkg)

	return &result, nil
}

func GetGoPackageNameFromGoImportPath(goImportPath string) string {
	return strings.Replace(path.Base(goImportPath), "-", "", -1)
}
