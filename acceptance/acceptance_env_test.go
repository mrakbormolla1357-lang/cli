//go:build acceptance

package acceptance_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestScriptEnvFromEnv(t *testing.T) {
	t.Setenv("GH_ACCEPTANCE_HOST", "github.example.com")
	t.Setenv("GH_ACCEPTANCE_ORG", "example-org")
	t.Setenv("GH_ACCEPTANCE_TOKEN", "token")
	t.Setenv("GH_ACCEPTANCE_SCRIPT", "pr-view.txtar")
	t.Setenv("GH_ACCEPTANCE_PRESERVE_WORK_DIR", "true")
	t.Setenv("GH_ACCEPTANCE_SKIP_DEFER", "true")

	var env testScriptEnv
	err := env.fromEnv()

	require.NoError(t, err)
	require.Equal(t, "github.example.com", env.host)
	require.Equal(t, "example-org", env.org)
	require.Equal(t, "token", env.token)
	require.Equal(t, "pr-view.txtar", env.script)
	require.True(t, env.preserveWorkDir)
	require.True(t, env.skipDefer)
}

func TestTestScriptEnvFromEnvMissingRequiredEnv(t *testing.T) {
	t.Setenv("GH_ACCEPTANCE_HOST", "github.example.com")
	t.Setenv("GH_ACCEPTANCE_ORG", "")
	t.Setenv("GH_ACCEPTANCE_TOKEN", "token")

	var env testScriptEnv
	err := env.fromEnv()

	require.EqualError(t, err, "environment variable(s) GH_ACCEPTANCE_ORG must be set and non-empty")
}

func TestTestScriptEnvFromEnvRejectsReservedOrgs(t *testing.T) {
	tests := []string{"github", "cli"}

	for _, org := range tests {
		t.Run(org, func(t *testing.T) {
			t.Setenv("GH_ACCEPTANCE_HOST", "github.example.com")
			t.Setenv("GH_ACCEPTANCE_ORG", org)
			t.Setenv("GH_ACCEPTANCE_TOKEN", "token")

			var env testScriptEnv
			err := env.fromEnv()

			require.EqualError(t, err, "GH_ACCEPTANCE_ORG cannot be 'github' or 'cli'")
		})
	}
}
