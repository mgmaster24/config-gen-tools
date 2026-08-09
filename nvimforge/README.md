# nvimforge

A cross-platform CLI that installs Neovim and generates a minimal,
language-aware Neovim configuration — a fresh `lazy.nvim` setup built
around `snacks.nvim`, `blink.cmp`, and `mason.nvim`, not a fork of any
existing distribution.

## Install

macOS / Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/mgmaster24/nvimforge/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/mgmaster24/nvimforge/main/install.ps1 | iex
```

Either script only downloads and places the `nvimforge` binary for your
platform from the latest GitHub release — no admin/sudo privileges
required.

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
```

### Supported languages

`rust`, `go`, `python`, `typescript`, `lua`, `c-cpp`, `bash`, `docker-yaml`.
Each is independently selectable; only the LSP servers, treesitter
parsers, formatters, and (where applicable) debug adapters for the
languages you pick get wired into the generated config. LSP/formatter/DAP
binaries themselves are installed by `mason.nvim` on Neovim's first
launch, not by `nvimforge`.

### Configuration

`nvimforge install` looks for `./nvimforge.toml`, then
`~/.config/nvimforge/config.toml`. Example:

```toml
languages = ["go", "rust", "lua"]
deploy_path = "~/.config/nvim"
backup = true
show_banner = true
```

Flags always override the config file; `--lang` alone (with `--yes`) can
drive a fully non-interactive first run without a config file at all.

### Prerequisites

`nvimforge doctor` (also run automatically as part of `install`) reports
what's missing — git, a C compiler + make, ripgrep, fd, and per-language
toolchains — and how to install each one for your OS's package manager. It
never installs anything itself; Neovim is the one thing `nvimforge`
actively installs/updates.

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
