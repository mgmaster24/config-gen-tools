// Package checks holds nvimforge's prerequisite data: which tools a given
// language selection needs on the host. The detection machinery itself
// lives in forge/prereq; this package only supplies the check list.
package checks

import (
	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
)

// UniversalChecks run regardless of which languages are selected.
var UniversalChecks = []prereq.Check{
	{
		Name:        "git",
		Description: "version control, required by lazy.nvim to clone plugins",
		Severity:    prereq.SeverityRecommended,
		Binary:      "git",
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMBrew, Command: "brew install git"},
			{Manager: prereq.PMApt, Command: "sudo apt install git"},
			{Manager: prereq.PMDnf, Command: "sudo dnf install git"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S git"},
			{Manager: prereq.PMWinget, Command: "winget install Git.Git"},
			{Manager: prereq.PMScoop, Command: "scoop install git"},
		},
	},
	{
		Name:        "c-compiler",
		Description: "C compiler, required to build treesitter parsers",
		Severity:    prereq.SeverityRecommended,
		Detect:      prereq.DetectFirstOf("cc", "gcc", "clang"),
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMBrew, Command: "xcode-select --install"},
			{Manager: prereq.PMApt, Command: "sudo apt install build-essential"},
			{Manager: prereq.PMDnf, Command: "sudo dnf groupinstall 'Development Tools'"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S base-devel"},
			{Manager: prereq.PMWinget, Command: "winget install Microsoft.VisualStudio.2022.BuildTools"},
		},
	},
	{
		Name:        "make",
		Description: "used alongside the C compiler to build treesitter parsers",
		Severity:    prereq.SeverityRecommended,
		Binary:      "make",
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMBrew, Command: "xcode-select --install"},
			{Manager: prereq.PMApt, Command: "sudo apt install make"},
			{Manager: prereq.PMDnf, Command: "sudo dnf install make"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S make"},
		},
	},
	{
		Name:        "ripgrep",
		Description: "used by snacks.nvim's picker for live grep",
		Severity:    prereq.SeverityRecommended,
		Binary:      "rg",
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMBrew, Command: "brew install ripgrep"},
			{Manager: prereq.PMApt, Command: "sudo apt install ripgrep"},
			{Manager: prereq.PMDnf, Command: "sudo dnf install ripgrep"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S ripgrep"},
			{Manager: prereq.PMWinget, Command: "winget install BurntSushi.ripgrep.MSVC"},
			{Manager: prereq.PMScoop, Command: "scoop install ripgrep"},
		},
	},
	{
		Name:        "fd",
		Description: "used by snacks.nvim's picker for faster file finding",
		Severity:    prereq.SeverityRecommended,
		Detect:      prereq.DetectFirstOf("fd", "fdfind"),
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMBrew, Command: "brew install fd"},
			{Manager: prereq.PMApt, Command: "sudo apt install fd-find"},
			{Manager: prereq.PMDnf, Command: "sudo dnf install fd-find"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S fd"},
			{Manager: prereq.PMWinget, Command: "winget install sharkdp.fd"},
			{Manager: prereq.PMScoop, Command: "scoop install fd"},
		},
	},
	{
		Name:        "downloader",
		Description: "nvimforge needs curl or wget to download Neovim releases",
		Severity:    prereq.SeverityRequired,
		Detect:      prereq.DetectFirstOf("curl", "wget"),
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMBrew, Command: "brew install curl"},
			{Manager: prereq.PMApt, Command: "sudo apt install curl"},
			{Manager: prereq.PMDnf, Command: "sudo dnf install curl"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S curl"},
		},
	},
	{
		Name:        "archiver",
		Description: "nvimforge needs tar or unzip to extract downloaded Neovim releases",
		Severity:    prereq.SeverityRequired,
		Detect:      prereq.DetectFirstOf("tar", "unzip"),
		Hints: []prereq.InstallHint{
			{Manager: prereq.PMApt, Command: "sudo apt install tar unzip"},
			{Manager: prereq.PMDnf, Command: "sudo dnf install tar unzip"},
			{Manager: prereq.PMPacman, Command: "sudo pacman -S tar unzip"},
		},
	},
}

// LanguageChecks lists the additional checks that apply only when a given
// language is selected. A language with no host-toolchain requirement
// (its LSP/formatter/DAP tooling is a standalone binary mason fetches, or
// it's already covered by UniversalChecks) maps to an empty slice.
var LanguageChecks = map[config.Language][]prereq.Check{
	config.LangRust: {
		{
			Name: "rustc", Description: "Rust compiler", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangRust), ScopeLabel: config.LangRust.DisplayName(), Binary: "rustc",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install rustup && rustup-init"},
				{Manager: prereq.PMWinget, Command: "winget install Rustlang.Rustup"},
			},
		},
		{
			Name: "cargo", Description: "Rust package manager, needed for the rustfmt formatter component", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangRust), ScopeLabel: config.LangRust.DisplayName(), Binary: "cargo",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install rustup && rustup-init"},
				{Manager: prereq.PMWinget, Command: "winget install Rustlang.Rustup"},
			},
		},
	},
	config.LangGo: {
		{
			Name: "go", Description: "Go toolchain, used by gopls and dap-go", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangGo), ScopeLabel: config.LangGo.DisplayName(), Binary: "go",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install go"},
				{Manager: prereq.PMApt, Command: "sudo apt install golang-go"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install golang"},
				{Manager: prereq.PMPacman, Command: "sudo pacman -S go"},
				{Manager: prereq.PMWinget, Command: "winget install GoLang.Go"},
			},
		},
	},
	config.LangPython: {
		{
			Name: "python3", Description: "Python interpreter, used by debugpy", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangPython), ScopeLabel: config.LangPython.DisplayName(), Binary: "python3",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install python3"},
				{Manager: prereq.PMApt, Command: "sudo apt install python3"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install python3"},
				{Manager: prereq.PMPacman, Command: "sudo pacman -S python"},
				{Manager: prereq.PMWinget, Command: "winget install Python.Python.3"},
			},
		},
		{
			Name: "pip", Description: "Python package installer, used to set up debugpy's virtualenv", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangPython), ScopeLabel: config.LangPython.DisplayName(), Detect: prereq.DetectFirstOf("pip3", "pip"),
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMApt, Command: "sudo apt install python3-pip"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install python3-pip"},
			},
		},
	},
	config.LangTypeScript: {
		{
			Name: "node", Description: "Node.js runtime, required by typescript-language-server", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangTypeScript), ScopeLabel: config.LangTypeScript.DisplayName(), Binary: "node",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install node"},
				{Manager: prereq.PMApt, Command: "sudo apt install nodejs"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install nodejs"},
				{Manager: prereq.PMPacman, Command: "sudo pacman -S nodejs npm"},
				{Manager: prereq.PMWinget, Command: "winget install OpenJS.NodeJS"},
			},
		},
		{
			Name: "npm", Description: "Node package manager, used by mason to install TS tooling", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangTypeScript), ScopeLabel: config.LangTypeScript.DisplayName(), Binary: "npm",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install node"},
				{Manager: prereq.PMApt, Command: "sudo apt install npm"},
			},
		},
	},
	config.LangLua:  {},
	config.LangCCpp: {},
	config.LangCSharp: {
		{
			// Unlike the other language checks, this one isn't just about
			// having a usable toolchain at the end: mason installs both
			// roslyn-language-server and csharpier as NuGet packages by
			// running `dotnet tool`, so with no dotnet on PATH the generated
			// config's C# support fails to install rather than degrading.
			Name:          "dotnet",
			Description:   "the .NET SDK; mason installs roslyn-language-server and csharpier as dotnet tools, so both fail without it",
			Severity:      prereq.SeverityRecommended,
			BlocksTooling: true,
			Scope:         string(config.LangCSharp), ScopeLabel: config.LangCSharp.DisplayName(), Binary: "dotnet",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install --cask dotnet-sdk"},
				{Manager: prereq.PMApt, Command: "sudo apt install dotnet-sdk-10.0"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install dotnet-sdk-10.0"},
				{Manager: prereq.PMPacman, Command: "sudo pacman -S dotnet-sdk"},
				{Manager: prereq.PMWinget, Command: "winget install Microsoft.DotNet.SDK.10"},
				{Manager: prereq.PMScoop, Command: "scoop install dotnet-sdk"},
			},
		},
	},
	config.LangBash: {
		{
			Name: "bash", Description: "Bash shell, for editing/running .sh files", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangBash), ScopeLabel: config.LangBash.DisplayName(), Binary: "bash",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMApt, Command: "sudo apt install bash"},
			},
		},
	},
	config.LangDockerYAML: {
		{
			Name: "docker", Description: "Docker CLI, useful for running the Dockerfiles/compose files you edit", Severity: prereq.SeverityRecommended,
			Scope: string(config.LangDockerYAML), ScopeLabel: config.LangDockerYAML.DisplayName(), Binary: "docker",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install --cask docker"},
				{Manager: prereq.PMWinget, Command: "winget install Docker.DockerDesktop"},
			},
		},
	},
}

// ForLanguages returns the checks that apply to a given language selection:
// the universal ones plus each selected language's own. Assembling the list
// here rather than inside forge/prereq keeps that package free of any
// knowledge of languages.
func ForLanguages(langs []config.Language) []prereq.Check {
	out := make([]prereq.Check, 0, len(UniversalChecks))
	out = append(out, UniversalChecks...)
	for _, l := range langs {
		out = append(out, LanguageChecks[l]...)
	}
	return out
}
