# Codebase Issue Proposals

This note records four small, actionable follow-up tasks found during a repository survey. The goal is to keep each task independently reviewable and low-risk.

## 1. Typo fix: acceptance token documentation

**Proposed task:** Fix the typo `authenticatin` to `authentication` in the acceptance test README.

**Evidence:** `acceptance/README.md` describes `GH_ACCEPTANCE_TOKEN` as "The token to use for authenticatin with the `GH_ACCEPTANCE_HOST`."

**Suggested scope:** One-line documentation-only change in `acceptance/README.md`.

## 2. Bug fix: harden `git.Command.setRepoDir`

**Proposed task:** Make `git.Command.setRepoDir` robust when `Command.Args` has an unexpected shape, and add focused unit coverage for short argument lists and commands that already include `-C`.

**Evidence:** `git/command.go` inserts `-C <repo>` by slicing around `index+3` after optionally locating `--`. That assumes enough elements exist after the insertion point and can panic if a caller constructs a shorter command than the current client helpers do.

**Suggested scope:** Replace the positional slice manipulation with a safer insertion helper, preserve the existing `--` handling behavior, and add tests in `git/command_test.go` or `git/client_test.go`.

## 3. Documentation discrepancy: `GH_HOST` setup

**Proposed task:** Correct the acceptance test README's `GH_HOST` description to say it is set from `GH_ACCEPTANCE_HOST`, not `GH_ACCEPTANCE_ORG`.

**Evidence:** `acceptance/README.md` says `GH_HOST` is set to the organization value, while `acceptance/acceptance_test.go` sets `GH_HOST` from `tsEnv.host` and separately sets `ORG` from `tsEnv.org`.

**Suggested scope:** One-line README clarification, ideally near the existing custom environment variable list.

## 4. Test improvement: scope-validation acceptance tests

**Proposed task:** Add tests or helper coverage for acceptance-test token scope validation so missing permissions fail before scripts create resources and schedule cleanup.

**Evidence:** `acceptance/README.md` already calls out scope validation as an unresolved TODO because cleanup can fail when a token lacks permissions such as `delete_repo`.

**Suggested scope:** Start with a small Go-level test around environment validation or a script metadata convention, then use that foundation to fail fast before running resource-lifecycle scripts.
