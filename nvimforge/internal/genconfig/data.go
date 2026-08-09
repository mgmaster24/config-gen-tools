package genconfig

import (
	"sort"

	"github.com/mgmaster24/nvimforge/internal/config"
)

// baselineTreesitterParsers install regardless of selected languages: they
// cover editing the generated config itself (lua/vim/vimdoc/query) and
// markdown, which nvimforge assumes every user wants.
var baselineTreesitterParsers = []string{"lua", "vim", "vimdoc", "query", "markdown", "markdown_inline"}

// FiletypeFormatters is one conform.nvim formatters_by_ft entry.
type FiletypeFormatters struct {
	Filetype   string
	Formatters []string
}

// TemplateData is everything the embedded Lua templates render from. It is
// built once per generation by BuildTemplateData and contains no
// non-deterministic content (no timestamps, no map iteration leaking
// through), which golden-file tests depend on.
type TemplateData struct {
	HasRust       bool
	HasGo         bool
	HasPython     bool
	HasTypeScript bool
	HasLua        bool
	HasCCpp       bool
	HasBash       bool
	HasDockerYAML bool

	TreesitterParsers    []string
	MasonEnsureInstalled []string
	Formatters           []FiletypeFormatters
}

// BuildTemplateData aggregates the LanguageSpecs for cfg.Languages into one
// deterministic TemplateData: every slice is deduplicated and sorted, so
// Render's output doesn't depend on the order languages were selected in.
func BuildTemplateData(cfg config.Config) TemplateData {
	data := TemplateData{}

	parserSet := make(map[string]bool)
	for _, p := range baselineTreesitterParsers {
		parserSet[p] = true
	}
	masonSet := make(map[string]bool)
	formatterByFiletype := make(map[string]map[string]bool)

	for _, lang := range cfg.Languages {
		switch lang {
		case config.LangRust:
			data.HasRust = true
		case config.LangGo:
			data.HasGo = true
		case config.LangPython:
			data.HasPython = true
		case config.LangTypeScript:
			data.HasTypeScript = true
		case config.LangLua:
			data.HasLua = true
		case config.LangCCpp:
			data.HasCCpp = true
		case config.LangBash:
			data.HasBash = true
		case config.LangDockerYAML:
			data.HasDockerYAML = true
		}

		spec, ok := LanguageSpecs[lang]
		if !ok {
			continue
		}

		for _, p := range spec.TreesitterParsers {
			parserSet[p] = true
		}
		for _, name := range spec.LSPServers {
			masonSet[name] = true
		}
		for _, f := range spec.Formatters {
			masonSet[f.MasonName] = true
			for _, ft := range f.Filetypes {
				if formatterByFiletype[ft] == nil {
					formatterByFiletype[ft] = make(map[string]bool)
				}
				formatterByFiletype[ft][f.MasonName] = true
			}
		}
		for _, name := range spec.DAPAdapters {
			masonSet[name] = true
		}
	}

	data.TreesitterParsers = sortedKeys(parserSet)
	data.MasonEnsureInstalled = sortedKeys(masonSet)

	ftKeys := make([]string, 0, len(formatterByFiletype))
	for ft := range formatterByFiletype {
		ftKeys = append(ftKeys, ft)
	}
	sort.Strings(ftKeys)
	for _, ft := range ftKeys {
		data.Formatters = append(data.Formatters, FiletypeFormatters{
			Filetype:   ft,
			Formatters: sortedKeys(formatterByFiletype[ft]),
		})
	}

	return data
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
