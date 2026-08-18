package genconfig

import (
	"reflect"
	"testing"

	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
)

func TestBuildTemplateData_NoLanguages_OnlyBaseline(t *testing.T) {
	data := BuildTemplateData(config.Config{})

	wantParsers := sortedKeys(toSet(baselineTreesitterParsers))
	if !reflect.DeepEqual(data.TreesitterParsers, wantParsers) {
		t.Errorf("TreesitterParsers = %v, want %v", data.TreesitterParsers, wantParsers)
	}
	if len(data.MasonEnsureInstalled) != 0 {
		t.Errorf("MasonEnsureInstalled should be empty with no languages, got %v", data.MasonEnsureInstalled)
	}
	if len(data.Formatters) != 0 {
		t.Errorf("Formatters should be empty with no languages, got %v", data.Formatters)
	}
	if data.HasRust || data.HasGo || data.HasPython || data.HasTypeScript || data.HasLua || data.HasCCpp || data.HasCSharp || data.HasBash || data.HasDockerYAML {
		t.Error("no Has* flag should be true with no languages selected")
	}
}

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, i := range items {
		set[i] = true
	}
	return set
}

func TestBuildTemplateData_UnselectedLanguageLeaksNothing(t *testing.T) {
	data := BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo}})

	for _, name := range []string{"rust-analyzer", "pyright", "typescript-language-server", "clangd"} {
		for _, got := range data.MasonEnsureInstalled {
			if got == name {
				t.Errorf("MasonEnsureInstalled contains %q, but only Go was selected", name)
			}
		}
	}
	if data.HasRust || data.HasPython || data.HasTypeScript || data.HasCCpp {
		t.Error("only HasGo should be true when only Go is selected")
	}
	if !data.HasGo {
		t.Error("HasGo should be true")
	}
}

func TestBuildTemplateData_DedupesSharedMasonPackage(t *testing.T) {
	// Both Rust and C/C++ contribute "codelldb" as a DAP adapter — it
	// must appear exactly once in the aggregated list.
	data := BuildTemplateData(config.Config{Languages: []config.Language{config.LangRust, config.LangCCpp}})

	count := 0
	for _, name := range data.MasonEnsureInstalled {
		if name == "codelldb" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("codelldb appears %d times in MasonEnsureInstalled, want exactly 1", count)
	}
}

func TestBuildTemplateData_IsOrderIndependent(t *testing.T) {
	a := BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo, config.LangRust, config.LangPython}})
	b := BuildTemplateData(config.Config{Languages: []config.Language{config.LangPython, config.LangRust, config.LangGo}})

	if !reflect.DeepEqual(a, b) {
		t.Errorf("BuildTemplateData should be independent of language selection order:\na = %+v\nb = %+v", a, b)
	}
}

func TestBuildTemplateData_FormattersGroupedByFiletypeAndSorted(t *testing.T) {
	data := BuildTemplateData(config.Config{Languages: []config.Language{config.LangTypeScript}})

	if len(data.Formatters) == 0 {
		t.Fatal("expected at least one filetype formatter entry for TypeScript")
	}
	// Filetypes must be in sorted order.
	for i := 1; i < len(data.Formatters); i++ {
		if data.Formatters[i-1].Filetype >= data.Formatters[i].Filetype {
			t.Errorf("Formatters not sorted by Filetype: %v", data.Formatters)
			break
		}
	}
}
