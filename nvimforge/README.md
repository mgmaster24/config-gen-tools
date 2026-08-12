# nvimforge

A cross-platform CLI that installs Neovim and generates a minimal,
language-aware Neovim configuration — a fresh `lazy.nvim` setup built
around `snacks.nvim`, `blink.cmp`, and `mason.nvim`, not a fork of any
existing distribution.

## Installation

### Install script (recommended)

macOS / Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/mgmaster24/config-gen-tools/main/nvimforge/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/mgmaster24/config-gen-tools/main/nvimforge/install.ps1 | iex
```

Either script only downloads, checksum-verifies, and places the `nvimforge`
binary for your platform from the latest GitHub release — no admin/sudo
privileges required.

Both accept two environment variables:

| Variable | Effect | Default |
|----------|--------|---------|
| `NVIMFORGE_VERSION` | Pin a release instead of taking the latest. Accepts `v1.2.3` or `nvimforge/v1.2.3`. | latest |
| `NVIMFORGE_INSTALL_DIR` | Where to place the binary. | `~/.local/bin` (Unix), `%LOCALAPPDATA%\nvimforge\bin` (Windows) |

### From source

```sh
go install github.com/mgmaster24/nvimforge/cmd/nvimforge@latest
```

This builds without release ldflags, so `nvimforge version` will report
`dev` rather than a real version. Prefer the install script unless you're
working on nvimforge itself.

### Releases

nvimforge lives in the [config-gen-tools](https://github.com/mgmaster24/config-gen-tools)
monorepo and is released under tool-scoped tags (`nvimforge/v1.2.3`), so its
version line is independent of the other tools in that repo. Archives and a
`checksums.txt` are attached to each release for `linux`, `darwin`
(amd64/arm64), and `windows` (amd64).

## Usage

```sh
nvimforge install      # verify prerequisites, install/update Neovim, generate the config
nvimforge doctor        # report missing prerequisites only (never installs anything)
nvimforge version
```

Run `nvimforge install` with no arguments for an interactive prompt, or
drive it non-interactively:

```sh
nvimforge install --lang go --lang rust --lang python --yes
nvimforge install --yes                  # the default languages
nvimforge install --dry-run              # print what would happen; write nothing
```

## Configuration

`nvimforge install` resolves its configuration in this order, with each
step overriding the one before it:

1. Built-in defaults (see below).
2. `./nvimforge.toml`, then `~/.config/nvimforge/config.toml`.
3. Flags (`--lang`, `--deploy-path`, `--backup`, `--no-banner`).

```toml
languages = ["go", "python", "typescript", "lua", "bash"]
deploy_path = "~/.config/nvim"
backup = true
show_banner = true
```

The file is decoded strictly — an unrecognized key is an error rather than
a silently ignored typo.

### Supported languages

| Language | Key | In defaults |
|----------|-----|-------------|
| Go | `go` | yes |
| Python | `python` | yes |
| TypeScript | `typescript` | yes |
| Lua | `lua` | yes |
| Bash | `bash` | yes |
| Rust | `rust` | no |
| C/C++ | `c-cpp` | no |
| C# | `csharp` | no |
| Docker/YAML | `docker-yaml` | no |

Each is independently selectable. Only the LSP servers, treesitter parsers,
formatters, and (where applicable) debug adapters for the languages you pick
get wired into the generated config. Those binaries are installed by
`mason.nvim` on Neovim's first launch, not by `nvimforge` itself.

### Defaults

With no config file and no flags, nvimforge starts from `go`, `python`,
`typescript`, `lua`, and `bash` — broad enough to be useful immediately,
while leaving the languages with heavier or host-toolchain-dependent tooling
opt-in.

The defaults are a starting point, not a floor:

- **Interactively**, they come pre-selected in the language prompt and can be
  deselected freely. Deselecting everything is an error — a config with no
  languages is rejected.
- **Non-interactively**, `nvimforge install --yes` with no config file
  installs exactly the defaults. Passing `--lang` replaces them outright
  rather than adding to them:

  ```sh
  nvimforge install --lang csharp --lang go --yes   # C# and Go only
  ```

### Prerequisites

`nvimforge doctor` (also run automatically as part of `install`) reports
what's missing — git, a C compiler + make, ripgrep, fd, and per-language
toolchains — and how to install each one for your OS's package manager. It
never installs anything itself; Neovim is the one thing `nvimforge`
actively installs/updates.

Prerequisites are report-only, with one exception. A few of them *block*
their language's tooling rather than merely degrading it: C#, for example,
needs the .NET SDK on PATH because mason installs `roslyn-language-server`
and `csharpier` as `dotnet` tools. Without it, mason can't install them at
all.

When a blocking prerequisite is missing, `nvimforge install` says so
explicitly and asks before generating, rather than silently writing a config
that cannot finish installing itself:

```
Warning: dotnet is missing, so mason cannot install the C# tooling.
The generated config will be written, but C# support will fail to install
until dotnet is on your PATH.

Missing dotnet. Generate the config anyway? (y/N)
```

Under `--yes` the warning is printed and the run proceeds — nvimforge never
refuses to generate on account of a language prerequisite.

## Development

```sh
go build ./...
go vet ./...
go test ./... -race
```

Template changes to the generated Neovim config should update the golden
fixtures:

```sh
go test ./internal/genconfig/... -run TestRender_Golden -update
```

Integration tests run a real Neovim against the generated config. They need
`nvim` on PATH and are behind a build tag, so a normal `go test ./...` never
picks them up:

```sh
go test -tags integration ./internal/integration/...
```

Tier 1 (the only tier implemented today) asserts every generated `.lua` file
parses. It's offline and takes about a second, so CI runs it on every pull
request. Deeper tiers — resolving plugin specs with `lazy.nvim`, installing
mason tooling, asserting an LSP client attaches — are network-bound and
belong on a schedule rather than the PR gate.

`scripts/ci-smoke.sh` is the end-to-end check CI runs after the unit tests.
It exercises a full `install --dry-run` and is the one place that touches the
network (the GitHub API, for Neovim release metadata):

```sh
bash scripts/ci-smoke.sh
```

### Adding a language

`internal/genconfig/spec.go` is the single source of truth. Adding a language
means: a `config.Language` constant plus an `AllLanguages` entry, a
`LanguageSpecs` entry, a `Has*` field and switch case in
`internal/genconfig/data.go`, and a `LanguageChecks` entry (even an empty
one — a test enforces this). DAP is the one part that isn't table-driven and
needs a hand-written block in `templates/lua/plugins/dap.lua.tmpl`.
