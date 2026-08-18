# gitforge

Generates an includable gitconfig with **directory-scoped identities** — a
work email inside `~/work/`, a personal one everywhere else — plus a set of
opinionated defaults.

The reason to generate this rather than hand-write it is one specific
footgun: a `gitdir:` condition without a trailing slash matches only that
exact path, **not the repositories beneath it**. `gitdir:~/work` silently
fails to apply to `~/work/some-repo`, and nothing warns you — you just
discover months of commits authored with the wrong address.

gitforge normalizes every directory to a trailing slash, so it cannot be got
wrong. There's an integration test asserting that real git genuinely does
*not* match subdirectories without it, so the rule stays honest if git ever
changes.

## Usage

```sh
gitforge generate            # write the config files
gitforge generate --dry-run  # print them, write nothing
gitforge doctor              # report missing prerequisites
```

gitforge **never writes your `~/.gitconfig`.** It generates files under
`~/.config/gitforge/` and prints the include to add:

```ini
[include]
	path = ~/.config/gitforge/gitconfig
```

## Configuration

`gitforge` reads `./gitforge.toml`, then `~/.config/gitforge/config.toml`.

```toml
deploy_path = "~/.config/gitforge"
backup = true
features = ["rerere", "autostash", "prune", "rebase-on-pull", "default-branch-main", "zdiff3"]

# Exactly one identity must omit `dir` — that's the default.
[[identities]]
name = "default"
user_name = "Your Name"
email = "you@personal.example"

[[identities]]
name = "work"
user_name = "Your Name"
email = "you@work.example"
dir = "~/work"                    # trailing slash added for you
signing_key = "~/.ssh/id_ed25519.pub"
ssh_sign = true
```

Unlike the other tools there is no useful default identity — no sensible
guess exists for a name and email — so a first run requires a config file.

Validation rejects the failure modes that are otherwise silent: no default
identity, two defaults, two identities matching the same directory (which
would make the winner depend on file order), and `ssh_sign` without a key.

### Features

| Key | Effect |
|-----|--------|
| `delta` | use delta as the diff pager |
| `rerere` | remember and reuse conflict resolutions |
| `autostash` | auto-stash local changes when rebasing |
| `prune` | prune deleted remote branches on fetch |
| `rebase-on-pull` | rebase instead of merge on pull |
| `default-branch-main` | name new repos' first branch `main` |
| `zdiff3` | use the zdiff3 conflict style |

`delta` is off by default and is the one **blocking** prerequisite: unlike
shellforge's guarded snippets, `core.pager = delta` is unconditional once
written, so a missing `delta` breaks every `git diff` rather than degrading
quietly. `gitforge generate` warns and confirms before writing it.

## Development

```sh
go build ./...
go vet ./...
go test ./... -race
go test -tags integration ./...   # drives real git
```

The integration tests initialise throwaway repositories and assert that
`includeIf` resolves the expected identity, with `GIT_CONFIG_GLOBAL` pointing
at the generated file so your real gitconfig is never involved.
