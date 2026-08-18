package genshell

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
)

// Run with -update to rewrite the fixtures after a deliberate template change:
//
//	go test ./internal/genshell/... -run TestRender_Golden -update
var updateGolden = flag.Bool("update", false, "update golden files instead of comparing against them")

func TestRender_Golden(t *testing.T) {
	combos := map[string]config.Config{
		"zsh-defaults": {Shell: config.ShellZsh, Integrations: config.DefaultIntegrations},
		"bash-defaults": {
			Shell: config.ShellBash, Integrations: config.DefaultIntegrations,
		},
		"zsh-all": {Shell: config.ShellZsh, Integrations: config.AllIntegrations},
		// A single PATH-phase integration: proves the file is still valid when
		// the later phases are all empty.
		"zsh-mise-only": {Shell: config.ShellZsh, Integrations: []config.Integration{config.IntegrationMise}},
	}

	names := make([]string, 0, len(combos))
	for name := range combos {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cfg := combos[name]
		t.Run(name, func(t *testing.T) {
			files, err := Render(cfg.Shell, BuildTemplateData(cfg))
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if len(files) != 1 {
				t.Fatalf("got %d files, want 1", len(files))
			}

			goldenPath := filepath.Join("testdata", "golden", name+".sh")
			if *updateGolden {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				if err := os.WriteFile(goldenPath, files[0].Content, 0o644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("reading golden file (run with -update to create it): %v", err)
			}
			if string(want) != string(files[0].Content) {
				t.Errorf("rendered output differs from %s\n--- got ---\n%s\n--- want ---\n%s",
					goldenPath, files[0].Content, want)
			}
		})
	}
}
