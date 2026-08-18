# shellforge

Generates a zsh or bash init script from a set of tool integrations, with the
hooks emitted in **dependency order**.

That ordering is the point. Several common shell hooks are position-sensitive
and break quietly when a hand-edited rc file drifts:

- Version managers (`mise`, `fnm`) rewrite `PATH`; anything resolving binaries
  after them must see the result.
- The prompt must initialise after `PATH` is final.
- `zoxide` must be **last**. Its own docs require it, and when it isn't it
  prints a "configuration issue" warning on every single shell startup.

shellforge assigns each integration a phase and sorts by it. Selection order
is discarded — placement is the generator's decision, not yours.

## Installation

macOS / Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/mgmaster24/config-gen-tools/main/shellforge/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/mgmaster24/config-gen-tools/main/shellforge/install.ps1 | iex
```

The script downloads, checksum-verifies, and places the `shellforge` binary
for your platform — no admin/sudo required. It resolves releases tagged
`shellforge/vX.Y.Z`, so nvimforge or gitforge releasing more recently never
shadows it.

| Variable | Effect | Default |
|----------|--------|---------|
| `SHELLFORGE_VERSION` | Pin a release. Accepts `v1.2.3` or `shellforge/v1.2.3`. | latest |
| `SHELLFORGE_INSTALL_DIR` | Where to place the binary. | `~/.local/bin` (Unix), `%LOCALAPPDATA%\shellforge\bin` (Windows) |

From source:

```sh
go install github.com/mgmaster24/config-gen-tools/shellforge/cmd/shellforge@latest
```

## Usage

```sh
shellforge generate                       # zsh, default integrations
shellforge generate --shell bash
shellforge generate --integration zoxide --integration starship
shellforge generate --dry-run             # print the script, write nothing
shellforge doctor                         # which integration binaries are missing
```

shellforge **never writes your `~/.zshrc`.** It generates a self-contained
script under `~/.config/shellforge/` and prints the one line to source:

```sh
[ -f "$HOME/.config/shellforge/init.zsh" ] && . "$HOME/.config/shellforge/init.zsh"
```

Add that as the last line of your rc file.

## Configuration

`shellforge generate` reads `./shellforge.toml`, then
`~/.config/shellforge/config.toml`. Flags override the file.

```toml
shell = "zsh"
integrations = ["fzf", "eza", "bat", "starship", "zoxide"]
deploy_path = "~/.config/shellforge"
backup = true
```

The file is decoded strictly — an unrecognised key is an error, not a
silently ignored typo.

### Integrations

| Key | Phase | Enabled by default |
|-----|-------|--------------------|
| `mise` | path | no |
| `fnm` | path | no |
| `fzf` | tool | yes |
| `direnv` | tool | no |
| `eza` | tool | yes |
| `bat` | tool | yes |
| `starship` | prompt | yes |
| `zoxide` | last | yes |

Version managers stay opt-in because they take over `PATH` resolution.

Every emitted block is wrapped in a `command -v` guard, so uninstalling a tool
without regenerating costs a no-op rather than an error on every new shell.
That's why no shellforge prereq is treated as blocking.

## Development

```sh
go build ./...
go vet ./...
go test ./... -race
```

Template changes should update the golden fixtures:

```sh
go test ./internal/genshell/... -run TestRender_Golden -update
```

Integration tests parse the generated script with the real shell (`zsh -n`,
`bash -n`) — offline, and skipped when the shell isn't installed:

```sh
go test -tags integration ./...
```
