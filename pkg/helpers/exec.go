package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"code.cestus.io/tools/fabricator/pkg/fabricator"
)

type Executor struct {
	root string
	io   fabricator.IOStreams
	env  fabricator.Environment
}

func NewExecutor(root string, io fabricator.IOStreams) *Executor {
	return &Executor{
		root: root,
		io:   io,
		env:  make(fabricator.Environment),
	}
}

func (e *Executor) WithRoot(root string) *Executor {
	return &Executor{
		root: root,
		io:   e.io,
		env:  e.env,
	}
}

func (e *Executor) WithEnv(key, value string) *Executor {
	env := fabricator.Environment{}
	env[key] = value

	return e.WithEnvMap(env)
}

func (e *Executor) WithEnvMap(env fabricator.Environment) *Executor {
	newEnv := fabricator.Environment{}

	if e.env != nil {
		for k, v := range e.env {
			newEnv[k] = v
		}
	}

	for k, v := range env {
		newEnv[k] = v
	}

	return &Executor{
		root: e.root,
		io:   e.io,
		env:  newEnv,
	}
}

func (e *Executor) setEnv(cmd *exec.Cmd) {
	if e.env != nil {
		env := append(os.Environ())

		for key, value := range e.env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}

		cmd.Env = env
	}
}

func (e *Executor) Output(ctx context.Context, path string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = e.root
	cmd.Stderr = e.io.ErrOut
	e.setEnv(cmd)

	data, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (e *Executor) JSONOutput(ctx context.Context, target interface{}, path string, args ...string) error {
	jsonData, err := e.Output(ctx, path, args...)

	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(jsonData), target); err != nil {
		return fmt.Errorf("failed to JSON-decode the program's output: %s", err)
	}

	return nil
}

func (e *Executor) Run(ctx context.Context, path string, args ...string) error {
	if e.io.Out != nil {
		fmt.Fprintf(e.io.Out, "executing %s %s\n", path, strings.Join(args, " "))
	}

	cmd := exec.Command(path, args...)
	cmd.Dir = e.root
	cmd.Stdout = e.io.Out
	cmd.Stderr = e.io.ErrOut
	e.setEnv(cmd)

	go func() {
		<-ctx.Done()
		// Do not send signals on a terminated process
		if cmd.Process == nil {
			return
		}

		if runtime.GOOS == "windows" {
			_ = cmd.Process.Signal(os.Kill)
			return
		}

		go func() {
			time.Sleep(5 * time.Second)
			_ = cmd.Process.Signal(os.Kill)
		}()

		cmd.Process.Signal(os.Interrupt)
	}()

	return cmd.Run()
}
