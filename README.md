# wtm — worktree manager

Manages git worktrees as siblings of the repo you're in. New worktrees live in
`<repo>-worktrees/<branch>` next to the main checkout (`/` → `-`), overridable
with `WTM_ROOT`.

## Install

```sh
go install github.com/ilyes-kechidi/wtm@latest
# shell integration for `wtm go` (cd into worktrees):
eval "$(wtm init bash)"   # or zsh | fish — add to your rc
```

## Usage

```sh
wtm list [--json]
wtm add <branch> [--base <ref>] [--no-env] [--no-setup] [--copy-env <glob>] [--allow-dirty]
wtm go [<branch>]        # prints path; cds with shell integration
wtm remove [<branch>] [--force] [--with-branch]  # --force also deletes untracked + ignored files
wtm clean [--dry-run]    # prune stale + remove merged-into-default
wtm                      # switch picker (built-in, no fzf needed)
```

## Env files

`~/.config/wtm/config.yaml` holds global defaults, `.wtm.yaml` in the repo
overrides it:

```yaml
copy:
  - frontend/.env
  - backend/.env*
```

Files are copied (never overwritten) on `add`. `--no-env` skips, `--copy-env`
adds extra globs.

## Post-create setup commands

`setup` steps run sequentially after env copy, each via `sh -c`, output
streamed, stopping at the first failure (the worktree stays in place):

```yaml
copy:
  - front/app/.env*
setup:
  - run: npm install
    dir: front/app
  - npm --prefix web install   # plain string = worktree root
```

`--no-setup` skips. Steps live in the same files (`~/.config/wtm/config.yaml`
global, `.wtm.yaml` repo-local override, merged per field).

## Teardown commands

`teardown` steps run the same way as `setup`, but before `git worktree remove`
on `wtm remove` and `wtm clean`. Failure aborts removal so the worktree is left
in place. A step whose command is not found is skipped, so worktrees created
before that step existed can still be removed. `wtm clean --dry-run` does not
run teardown.

```yaml
teardown:
  - ./scripts/worktree-down.sh
  - run: docker compose down
    dir: .
```

## Environment variables for setup and teardown

Both `setup` and `teardown` commands receive:

| Variable       | Value                          |
|----------------|--------------------------------|
| `WTM_BRANCH`   | Branch name for the worktree   |
| `WTM_WORKTREE` | Absolute path to the worktree  |
