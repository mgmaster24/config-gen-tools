# config-gen-tools

A monorepo of configuration-generation tools. Each tool is a self-contained
Go module in its own top-level directory, built, tested, and released
independently.

## Tools

| Tool | Description |
|------|-------------|
| [nvimforge](./nvimforge) | Installs Neovim and generates a minimal, language-aware `lazy.nvim` configuration. |

## Repository layout

```
<tool>/
  go.mod                 # one module per tool — this is what CI discovers
  .goreleaser.yaml       # required; the release workflow keys off its presence
  scripts/ci-smoke.sh    # optional end-to-end check run by CI
  cmd/<tool>/main.go
  internal/...
```

## Adding a new tool

1. Create `<tool>/` with its own `go.mod`.
2. Add `<tool>/.goreleaser.yaml` — copy `nvimforge/.goreleaser.yaml` and
   change `project_name`, `builds.main`, `builds.binary`, and the
   `buildinfo` ldflags paths.
3. Optionally add `<tool>/scripts/ci-smoke.sh` for an end-to-end check that
   doesn't fit inside `go test`.

No CI changes are needed. The `discover` job in `.github/workflows/ci.yml`
finds every directory containing a `go.mod` and fans the build, vet, test,
smoke, and lint jobs out over them.

## CI

`.github/workflows/ci.yml` runs on every push to `main` and every pull
request:

- **discover** — emits the list of tool directories as a JSON matrix.
- **test** — `go build`, `go vet`, `go test -race`, then `scripts/ci-smoke.sh`
  if present, across Linux, macOS, and Windows for each tool.
- **lint** — `golangci-lint` per tool.

## Releasing

Releases are per tool, using **tool-scoped tags**:

```sh
git tag nvimforge/v1.2.3
git push origin nvimforge/v1.2.3
```

`.github/workflows/release.yml` parses the tool name and version out of the
tag, runs GoReleaser inside that tool's directory to cross-compile and
archive, then publishes a GitHub release attached to the full prefixed tag.

Two tools can therefore both be at `v1.2.3` without colliding, and releasing
one never implies a release of the others.

> GoReleaser's native monorepo support (`monorepo.tag_prefix`) is a
> GoReleaser Pro feature. This repo uses OSS GoReleaser, so the workflow runs
> `release --clean --skip=publish,validate` to build and archive, and creates
> the GitHub release with `gh release create` against the prefixed tag. That
> split is deliberate: letting GoReleaser publish would require handing it the
> bare `v1.2.3`, which would create an unprefixed tag shared across every tool
> in the repo.
