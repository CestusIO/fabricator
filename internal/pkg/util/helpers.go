package util

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	DefaultErrorExitCode = 1
)

var fatalErrHandler = fatal

// BehaviorOnFatal allows you to override the default behavior when a fatal
// error occurs, which is to call os.Exit(code). You can pass 'panic' as a function
// here if you prefer the panic() over os.Exit(1).
func BehaviorOnFatal(f func(string, int)) {
	fatalErrHandler = f
}

// DefaultBehaviorOnFatal allows you to undo any previous override.  Useful in
// tests.
func DefaultBehaviorOnFatal() {
	fatalErrHandler = fatal
}

// fatal prints the message (if provided) and then exits. If V(6) or greater,
// klog.Fatal is invoked for extended information.
func fatal(msg string, code int) {
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(code)
}

// ErrExit may be passed to CheckError to instruct it to output nothing but exit with
// status code 1.
var ErrExit = fmt.Errorf("exit")

// CheckErr prints a user friendly error to STDERR and exits with a non-zero
// exit code. Unrecognized errors will be printed with an "error: " prefix.
//
// This method is generic to the command in use and may be used by non-fabricator
// commands.
func CheckErr(err error) {
	checkErr(err, fatalErrHandler)
}

// checkErr formats a given error as a string and calls the passed handleErr
// func with that string and an kubectl exit code.
func checkErr(err error, handleErr func(string, int)) {

	if err == nil {
		return
	}
	fatalErrHandler(err.Error(), DefaultErrorExitCode)
}

// DefaultSubCommandRun prints a command's help string to the specified output if no
// arguments (sub-commands) are provided, or a usage error otherwise.
func DefaultSubCommandRun(out io.Writer) func(c *cobra.Command, args []string) {
	return func(c *cobra.Command, args []string) {
		c.SetOutput(out)
		RequireNoArguments(c, args)
		c.Help()
		CheckErr(ErrExit)
	}
}

// RequireNoArguments exits with a usage error if extra arguments are provided.
func RequireNoArguments(c *cobra.Command, args []string) {
	if len(args) > 0 {
		CheckErr(UsageErrorf(c, "unknown command %q", strings.Join(args, " ")))
	}
}

// UsageErrorf prints the usage error
func UsageErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s\nSee '%s -h' for help and examples", msg, cmd.CommandPath())
}
