package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		stdout   string
		stderr   string
		wantErr  *GitError
	}{
		{
			name:     "successful command",
			stdout:   "hello world",
			stderr:   "",
			exitCode: 0,
			wantErr:  nil,
		},
		{
			name:     "not a repo failure",
			stdout:   "",
			stderr:   "fatal: not a git repository (or any of the parent directories): .git",
			exitCode: 128,
			wantErr: &GitError{
				ExitCode: 128,
				Stderr:   "fatal: not a git repository (or any of the parent directories): .git",
				err:      &exec.ExitError{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cmd := Command{
				&exec.Cmd{
					Path: createMockExecutable(t, tt.stdout, tt.stderr, tt.exitCode),
				},
			}

			out, err := cmd.Output()
			if tt.wantErr != nil {
				require.Error(t, err)
				var gitError *GitError
				require.ErrorAs(t, err, &gitError)
				assert.Equal(t, tt.wantErr.ExitCode, gitError.ExitCode)
				assert.Equal(t, tt.wantErr.Stderr, gitError.Stderr)
				assert.Equal(t, tt.wantErr.Error(), gitError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.stdout, string(out))
		})
	}
}

func TestSetRepoDir(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "inserts repo dir after command",
			args: []string{"git", "status"},
			want: []string{"git", "-C", "/path/to/repo", "status"},
		},
		{
			name: "updates existing repo dir",
			args: []string{"git", "-C", "/old/path", "status"},
			want: []string{"git", "-C", "/path/to/repo", "status"},
		},
		{
			name: "completes dangling repo dir flag",
			args: []string{"git", "-C"},
			want: []string{"git", "-C", "/path/to/repo"},
		},
		{
			name: "handles short args",
			args: []string{"git"},
			want: []string{"git", "-C", "/path/to/repo"},
		},
		{
			name: "inserts repo dir after helper process args",
			args: []string{"testbin", "-test.run=TestCommandMocking", "--", "git", "status"},
			want: []string{"testbin", "-test.run=TestCommandMocking", "--", "git", "-C", "/path/to/repo", "status"},
		},
		{
			name: "handles incomplete helper process args",
			args: []string{"testbin", "-test.run=TestCommandMocking", "--"},
			want: []string{"testbin", "-test.run=TestCommandMocking", "--", "-C", "/path/to/repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{
				&exec.Cmd{
					Args: append([]string{}, tt.args...),
				},
			}

			cmd.setRepoDir("/path/to/repo")

			assert.Equal(t, tt.want, cmd.Args)
		})
	}
}

func createMockExecutable(t *testing.T, stdout string, stderr string, exitCode int) string {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "main.go")
	binaryPath := filepath.Join(tmpDir, "mockexec")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Create Go source
	source := buildCommandSourceCode(exitCode, stdout, stderr)

	// Write source file
	if err := os.WriteFile(sourcePath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}

	// Compile
	cmd := exec.Command("go", "build", "-o", binaryPath, sourcePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile: %v\n%s", err, out)
	}
	return binaryPath

}

func buildCommandSourceCode(exitCode int, stdout, stderr string) string {
	return fmt.Sprintf(`package main
	import (
		   "fmt"
		"os"
	)
	func main() {
		fmt.Printf(%q)
		fmt.Fprintf(os.Stderr, %q)
		os.Exit(%d)
	}`, stdout, stderr, exitCode)
}
