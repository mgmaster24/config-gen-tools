package prereq

import "github.com/mgmaster24/nvimforge/internal/config"

// UniversalChecks run regardless of which languages are selected.
var UniversalChecks = []Check{
	{
		Name:        "git",
		Description: "version control, required by lazy.nvim to clone plugins",
		Severity:    SeverityRecommended,
		Binary:      "git",
		Hints: []InstallHint{
			{PMBrew, "brew install git"},
			{PMApt, "sudo apt install git"},
			{PMDnf, "sudo dnf install git"},
			{PMPacman, "sudo pacman -S git"},
			{PMWinget, "winget install Git.Git"},
			{PMScoop, "scoop install git"},
		},
	},
	{
		Name:        "c-compiler",
		Description: "C compiler, required to build treesitter parsers",
		Severity:    SeverityRecommended,
		Detect:      detectFirstOf("cc", "gcc", "clang"),
		Hints: []InstallHint{
			{PMBrew, "xcode-select --install"},
			{PMApt, "sudo apt install build-essential"},
			{PMDnf, "sudo dnf groupinstall 'Development Tools'"},
			{PMPacman, "sudo pacman -S base-devel"},
			{PMWinget, "winget install Microsoft.VisualStudio.2022.BuildTools"},
		},
	},
	{
		Name:        "make",
		Description: "used alongside the C compiler to build treesitter parsers",
		Severity:    SeverityRecommended,
		Binary:      "make",
		Hints: []InstallHint{
			{PMBrew, "xcode-select --install"},
			{PMApt, "sudo apt install make"},
			{PMDnf, "sudo dnf install make"},
			{PMPacman, "sudo pacman -S make"},
		},
	},
	{
		Name:        "ripgrep",
		Description: "used by snacks.nvim's picker for live grep",
		Severity:    SeverityRecommended,
		Binary:      "rg",
		Hints: []InstallHint{
			{PMBrew, "brew install ripgrep"},
			{PMApt, "sudo apt install ripgrep"},
			{PMDnf, "sudo dnf install ripgrep"},
			{PMPacman, "sudo pacman -S ripgrep"},
			{PMWinget, "winget install BurntSushi.ripgrep.MSVC"},
			{PMScoop, "scoop install ripgrep"},
		},
	},
	{
		Name:        "fd",
		Description: "used by snacks.nvim's picker for faster file finding",
		Severity:    SeverityRecommended,
		Detect:      detectFirstOf("fd", "fdfind"),
		Hints: []InstallHint{
			{PMBrew, "brew install fd"},
			{PMApt, "sudo apt install fd-find"},
			{PMDnf, "sudo dnf install fd-find"},
			{PMPacman, "sudo pacman -S fd"},
			{PMWinget, "winget install sharkdp.fd"},
			{PMScoop, "scoop install fd"},
		},
	},
	{
		Name:        "downloader",
		Description: "nvimforge needs curl or wget to download Neovim releases",
		Severity:    SeverityRequired,
		Detect:      detectFirstOf("curl", "wget"),
		Hints: []InstallHint{
			{PMBrew, "brew install curl"},
			{PMApt, "sudo apt install curl"},
			{PMDnf, "sudo dnf install curl"},
			{PMPacman, "sudo pacman -S curl"},
		},
	},
	{
		Name:        "archiver",
		Description: "nvimforge needs tar or unzip to extract downloaded Neovim releases",
		Severity:    SeverityRequired,
		Detect:      detectFirstOf("tar", "unzip"),
		Hints: []InstallHint{
			{PMApt, "sudo apt install tar unzip"},
			{PMDnf, "sudo dnf install tar unzip"},
			{PMPacman, "sudo pacman -S tar unzip"},
		},
	},
}

// LanguageChecks lists the additional checks that apply only when a given
// language is selected. A language with no host-toolchain requirement
// (its LSP/formatter/DAP tooling is a standalone binary mason fetches, or
// it's already covered by UniversalChecks) maps to an empty slice.
var LanguageChecks = map[config.Language][]Check{
	config.LangRust: {
		{
			Name: "rustc", Description: "Rust compiler", Severity: SeverityRecommended,
			Language: config.LangRust, Binary: "rustc",
			Hints: []InstallHint{
				{PMBrew, "brew install rustup && rustup-init"},
				{PMWinget, "winget install Rustlang.Rustup"},
			},
		},
		{
			Name: "cargo", Description: "Rust package manager, needed for the rustfmt formatter component", Severity: SeverityRecommended,
			Language: config.LangRust, Binary: "cargo",
			Hints: []InstallHint{
				{PMBrew, "brew install rustup && rustup-init"},
				{PMWinget, "winget install Rustlang.Rustup"},
			},
		},
	},
	config.LangGo: {
		{
			Name: "go", Description: "Go toolchain, used by gopls and dap-go", Severity: SeverityRecommended,
			Language: config.LangGo, Binary: "go",
			Hints: []InstallHint{
				{PMBrew, "brew install go"},
				{PMApt, "sudo apt install golang-go"},
				{PMDnf, "sudo dnf install golang"},
				{PMPacman, "sudo pacman -S go"},
				{PMWinget, "winget install GoLang.Go"},
			},
		},
	},
	config.LangPython: {
		{
			Name: "python3", Description: "Python interpreter, used by debugpy", Severity: SeverityRecommended,
			Language: config.LangPython, Binary: "python3",
			Hints: []InstallHint{
				{PMBrew, "brew install python3"},
				{PMApt, "sudo apt install python3"},
				{PMDnf, "sudo dnf install python3"},
				{PMPacman, "sudo pacman -S python"},
				{PMWinget, "winget install Python.Python.3"},
			},
		},
		{
			Name: "pip", Description: "Python package installer, used to set up debugpy's virtualenv", Severity: SeverityRecommended,
			Language: config.LangPython, Detect: detectFirstOf("pip3", "pip"),
			Hints: []InstallHint{
				{PMApt, "sudo apt install python3-pip"},
				{PMDnf, "sudo dnf install python3-pip"},
			},
		},
	},
	config.LangTypeScript: {
		{
			Name: "node", Description: "Node.js runtime, required by typescript-language-server", Severity: SeverityRecommended,
			Language: config.LangTypeScript, Binary: "node",
			Hints: []InstallHint{
				{PMBrew, "brew install node"},
				{PMApt, "sudo apt install nodejs"},
				{PMDnf, "sudo dnf install nodejs"},
				{PMPacman, "sudo pacman -S nodejs npm"},
				{PMWinget, "winget install OpenJS.NodeJS"},
			},
		},
		{
			Name: "npm", Description: "Node package manager, used by mason to install TS tooling", Severity: SeverityRecommended,
			Language: config.LangTypeScript, Binary: "npm",
			Hints: []InstallHint{
				{PMBrew, "brew install node"},
				{PMApt, "sudo apt install npm"},
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
			Severity:      SeverityRecommended,
			BlocksTooling: true,
			Language:      config.LangCSharp, Binary: "dotnet",
			Hints: []InstallHint{
				{PMBrew, "brew install --cask dotnet-sdk"},
				{PMApt, "sudo apt install dotnet-sdk-10.0"},
				{PMDnf, "sudo dnf install dotnet-sdk-10.0"},
				{PMPacman, "sudo pacman -S dotnet-sdk"},
				{PMWinget, "winget install Microsoft.DotNet.SDK.10"},
				{PMScoop, "scoop install dotnet-sdk"},
			},
		},
	},
	config.LangBash: {
		{
			Name: "bash", Description: "Bash shell, for editing/running .sh files", Severity: SeverityRecommended,
			Language: config.LangBash, Binary: "bash",
			Hints: []InstallHint{
				{PMApt, "sudo apt install bash"},
			},
		},
	},
	config.LangDockerYAML: {
		{
			Name: "docker", Description: "Docker CLI, useful for running the Dockerfiles/compose files you edit", Severity: SeverityRecommended,
			Language: config.LangDockerYAML, Binary: "docker",
			Hints: []InstallHint{
				{PMBrew, "brew install --cask docker"},
				{PMWinget, "winget install Docker.DockerDesktop"},
			},
		},
	},
}
