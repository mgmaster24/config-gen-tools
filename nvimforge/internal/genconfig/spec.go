// Package genconfig generates a brand-new, minimal lazy.nvim-based Neovim
// configuration from a set of selected languages. LanguageSpecs is the
// single source of truth for what each language contributes; everything
// else in this package aggregates and renders that data.
package genconfig

import "github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"

// FormatterSpec is one conform.nvim formatter entry: a mason package name
// and the filetypes it formats.
type FormatterSpec struct {
	MasonName string
	Filetypes []string
}

// LanguageSpec is everything one Language contributes to the generated
// config. An unselected language's LanguageSpec is simply never consulted
// by BuildTemplateData, which is what guarantees it can't leak into
// generated output.
type LanguageSpec struct {
	TreesitterParsers []string
	// LSPServers are mason package names for LSP servers. They don't need
	// individual lspconfig.setup{} calls: mason-lspconfig's
	// automatic_enable picks up anything mason has installed.
	LSPServers []string
	Formatters []FormatterSpec
	// DAPAdapters are mason package names for debug adapters. Only
	// languages listed here get a hand-written dap.adapters/configurations
	// block in the generated dap.lua (DAP setup isn't uniform enough
	// across debuggers to be purely table-driven).
	DAPAdapters []string
}

var LanguageSpecs = map[config.Language]LanguageSpec{
	config.LangRust: {
		TreesitterParsers: []string{"rust", "toml"},
		LSPServers:        []string{"rust-analyzer"},
		Formatters:        []FormatterSpec{{MasonName: "rustfmt", Filetypes: []string{"rust"}}},
		DAPAdapters:       []string{"codelldb"},
	},
	config.LangGo: {
		TreesitterParsers: []string{"go", "gomod", "gowork", "gosum"},
		LSPServers:        []string{"gopls"},
		Formatters:        []FormatterSpec{{MasonName: "goimports", Filetypes: []string{"go"}}},
		DAPAdapters:       []string{"delve"},
	},
	config.LangPython: {
		TreesitterParsers: []string{"python"},
		LSPServers:        []string{"pyright", "ruff"},
		Formatters:        []FormatterSpec{{MasonName: "ruff", Filetypes: []string{"python"}}},
		DAPAdapters:       []string{"debugpy"},
	},
	config.LangTypeScript: {
		TreesitterParsers: []string{"typescript", "tsx", "javascript"},
		LSPServers:        []string{"typescript-language-server"},
		Formatters: []FormatterSpec{{
			MasonName: "prettierd",
			Filetypes: []string{"typescript", "typescriptreact", "javascript", "javascriptreact"},
		}},
	},
	config.LangLua: {
		TreesitterParsers: []string{"lua"},
		LSPServers:        []string{"lua-language-server"},
		Formatters:        []FormatterSpec{{MasonName: "stylua", Filetypes: []string{"lua"}}},
	},
	config.LangCCpp: {
		TreesitterParsers: []string{"c", "cpp"},
		LSPServers:        []string{"clangd"},
		Formatters:        []FormatterSpec{{MasonName: "clang-format", Filetypes: []string{"c", "cpp"}}},
		DAPAdapters:       []string{"codelldb"},
	},
	config.LangCSharp: {
		// The parser is "c_sharp" (underscore) while the filetype is "cs" —
		// the one language here where the two names don't line up.
		TreesitterParsers: []string{"c_sharp"},
		// Both of these are NuGet packages: mason installs them by shelling
		// out to `dotnet tool`, which is why LangCSharp's prereq check on the
		// .NET SDK matters more than the other languages' toolchain checks.
		LSPServers:  []string{"roslyn-language-server"},
		Formatters:  []FormatterSpec{{MasonName: "csharpier", Filetypes: []string{"cs"}}},
		DAPAdapters: []string{"netcoredbg"},
	},
	config.LangBash: {
		TreesitterParsers: []string{"bash"},
		LSPServers:        []string{"bash-language-server"},
		Formatters:        []FormatterSpec{{MasonName: "shfmt", Filetypes: []string{"sh", "bash"}}},
	},
	config.LangDockerYAML: {
		TreesitterParsers: []string{"dockerfile", "yaml"},
		LSPServers:        []string{"dockerfile-language-server", "yaml-language-server"},
		Formatters:        []FormatterSpec{{MasonName: "yamlfmt", Filetypes: []string{"yaml"}}},
	},
}
