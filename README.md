# config-gen-tools

A monorepo of configuration-generation tools. Each tool is a self-contained
Go module in its own top-level directory, built, tested, and released
independently, on top of a shared `forge` core.

## Tools

| Tool | Description |
|------|-------------|
| [nvimforge](./nvimforge) | Installs Neovim and generates a minimal, language-aware `lazy.nvim` configuration. |
| [shellforge](./shellforge) | Generates a zsh/bash init script with tool hooks emitted in dependency order. |
| [gitforge](./gitforge) | Generates an includable gitconfig with directory-scoped identities. |
| [forge](./forge) | Shared library: filesystem helpers, prerequisite detection, command runner. Not a CLI. |

Each generator follows the same contract: it **never edits a file you own**.
It writes self-contained output under `~/.config/<tool>/` and prints the one
line to include or source from your own rc file.

## The shared core

`forge` is where the second tool stopped being a copy-paste of the first:

| Package | Provides |
|---|---|
| `forge/fsutil` | atomic writes, backup-if-not-ours, per-tool generated markers, `~` expansion |
| `forge/prereq` | check/detect/report with package-manager-aware install hints |
| `forge/runner` | the `os/exec` seam that makes detection unit-testable |

`forge/prereq` knows nothing about languages, shells, or identities — each
tool supplies its own check list via a `Scope` string, so the framework never
grows a dependency on any one tool's domain types.

## Repository layout

```
forge/                   # shared library module (no cmd/)
<tool>/
  go.mod                 # one module per tool — this is what CI discovers
  .goreleaser.yaml       # required to be releasable; the release workflow keys off it
  .golangci.yml
  scripts/ci-smoke.sh    # optional end-to-end check run by CI
  cmd/<tool>/main.go
  internal/
    config/              # the tool's own TOML config
    gen*/                # templates + render/write, with golden tests
    checks/              # prereq data on top of forge/prereq
    integration/         # //go:build integration — validates generated output
```

## Module paths and versioning

Modules are named for their location in the repo:

```
github.com/mgmaster24/config-gen-tools/<tool>
```

This is required, not cosmetic. Go resolves a module in subdirectory `foo/`
from the tag `foo/v1.2.3` — which is exactly the tool-scoped tagging the
release workflow uses. It also means `go install` works:

```sh
go install github.com/mgmaster24/config-gen-tools/nvimforge/cmd/nvimforge@latest
```

`forge` is published at `v0.1.0`, and each tool requires it by version. There
are **no `replace` directives** — a local `replace` would make `go install`
fail for anyone outside the repo, so the tools are verified to build with
`GOWORK=off` against the published module.

`go.work` at the repo root is for local development only. It lets edits
across modules build together, but nothing depends on it: CI and `go install`
resolve `forge` from its tag.

### Releasing a forge change

`forge` is a library, so it ships no binary — its "release" is the tag alone:

```sh
git tag forge/v0.2.0 && git push origin forge/v0.2.0
```

Then bump `github.com/mgmaster24/config-gen-tools/forge` in each dependent
tool's `go.mod` and run `go mod tidy`. Because `go.work` shadows the required
version locally, verify with `GOWORK=off go build ./...` before pushing —
that's the only way to catch a tool still pinned to an older forge.

The release workflow recognizes library tags: a module with no
`.goreleaser.yaml` is reported as a notice and skipped rather than failing
the run.

## Adding a new tool

1. Create `<tool>/` with `go.mod` declaring
   `github.com/mgmaster24/config-gen-tools/<tool>`, requiring
   `github.com/mgmaster24/config-gen-tools/forge` at its current version.
   Don't add a `replace` — see above.
2. Add `.goreleaser.yaml` and `.golangci.yml` (copy an existing pair).
3. Add `<tool>` to `go.work`.
4. Optionally add `scripts/ci-smoke.sh` and `internal/integration`.
5. Optionally add `install.sh` / `install.ps1` (copy an existing pair and
   change `TOOL`, the env var prefixes, and the closing hint).

No CI changes needed — `discover` finds every directory containing a
`go.mod` and fans the jobs out over it.

## CI

`.github/workflows/ci.yml` runs on every push to `main` and every PR:

- **discover** — emits the tool directories as a JSON matrix.
- **test** — `go build`, `go vet`, `go test -race`, then `scripts/ci-smoke.sh`
  if present, across Linux, macOS, and Windows.
- **integration** — tier-1 checks that generated output is *valid*: Neovim
  parses the Lua, the shell parses the script, git parses the gitconfig.
  Offline and fast, so it belongs on the PR gate.
- **lint** — `golangci-lint` per module, pinned to a version built with a Go
  release at least as new as the modules target.

## Releasing

Per tool, using tool-scoped tags:

```sh
git tag nvimforge/v1.2.3
git push origin nvimforge/v1.2.3
```

`.github/workflows/release.yml` parses the tool and version out of the tag,
runs GoReleaser in that tool's directory to cross-compile and archive, then
publishes a GitHub release attached to the full prefixed tag. Two tools can
both sit at `v1.2.3` without colliding, and releasing one never implies
releasing the others.

> GoReleaser's native monorepo support (`monorepo.tag_prefix`) is Pro-only.
> This repo uses OSS GoReleaser, so the workflow runs
> `release --clean --skip=publish,validate` to build and archive, and creates
> the release with `gh release create` against the prefixed tag. Letting
> GoReleaser publish would require handing it the bare `v1.2.3`, which would
> create an unprefixed tag shared across every tool in the repo.
