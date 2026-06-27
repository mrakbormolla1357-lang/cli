package git

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/cli/cli/v2/internal/run"
)

type commandCtx = func(ctx context.Context, name string, args ...string) *exec.Cmd

type Command struct {
	*exec.Cmd
}

func (gc *Command) Run() error {
	stderr := &bytes.Buffer{}
	if gc.Cmd.Stderr == nil {
		gc.Cmd.Stderr = stderr
	}
	// This is a hack in order to not break the hundreds of
	// existing tests that rely on `run.PrepareCmd` to be invoked.
	err := run.PrepareCmd(gc.Cmd).Run()
	if err != nil {
		ge := GitError{err: err, Stderr: stderr.String()}
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			ge.ExitCode = exitError.ExitCode()
		}
		return &ge
	}
	return nil
}

func (gc *Command) Output() ([]byte, error) {
	gc.Stdout = nil
	gc.Stderr = nil
	// This is a hack in order to not break the hundreds of
	// existing tests that rely on `run.PrepareCmd` to be invoked.
	out, err := run.PrepareCmd(gc.Cmd).Output()
	if err != nil {
		ge := GitError{err: err}

		// In real implementation, this should be an exec.ExitError, as below,
		// but the tests use a different type because exec.ExitError are difficult
		// to create. We want to get the exit code and stderr, but stderr
		// is not a method and so tests can't access it.
		// THIS MEANS THAT TESTS WILL NOT CORRECTLY HAVE STDERR SET,
		// but at least tests can get the exit code.
		var exitErrorWithExitCode errWithExitCode
		if errors.As(err, &exitErrorWithExitCode) {
			ge.ExitCode = exitErrorWithExitCode.ExitCode()
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			ge.Stderr = string(exitError.Stderr)
		}
		err = &ge
	}
	return out, err
}

func (gc *Command) setRepoDir(repoDir string) {
	for i, arg := range gc.Args {
		if arg == "-C" {
			if i+1 < len(gc.Args) {
				gc.Args[i+1] = repoDir
			} else {
				gc.Args = append(gc.Args, repoDir)
			}
			return
		}
	}

	commandAt := 0
	// Handle the test helper process separator without treating git's own
	// "--" pathspec separator as the command location.
	for i, arg := range gc.Args {
		if arg == "--" && hasTestFlag(gc.Args[:i]) {
			commandAt = i + 1
			break
		}
	}

	insertAt := commandAt + 1
	if insertAt > len(gc.Args) {
		insertAt = len(gc.Args)
	}

	gc.Args = append(gc.Args[:insertAt], append([]string{"-C", repoDir}, gc.Args[insertAt:]...)...)
}

func hasTestFlag(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

// Allow individual commands to be modified from the default client options.
type CommandModifier func(*Command)

func WithStderr(stderr io.Writer) CommandModifier {
	return func(gc *Command) {
		gc.Stderr = stderr
	}
}

func WithStdout(stdout io.Writer) CommandModifier {
	return func(gc *Command) {
		gc.Stdout = stdout
	}
}

func WithStdin(stdin io.Reader) CommandModifier {
	return func(gc *Command) {
		gc.Stdin = stdin
	}
}

func WithRepoDir(repoDir string) CommandModifier {
	return func(gc *Command) {
		gc.setRepoDir(repoDir)
	}
}
