package genconfig

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mgmaster24/nvimforge/internal/config"
)

// Run with -update to (re)write the golden fixtures under testdata/golden/
// after a deliberate template change, e.g.:
//
//	go test ./internal/genconfig/... -run TestRender_Golden -update
var updateGolden = flag.Bool("update", false, "update golden files instead of comparing against them")

func TestRender_Golden(t *testing.T) {
	combos := map[string][]config.Language{
		"none":          nil,
		"go-only":       {config.LangGo},
		"rust-and-ccpp": {config.LangRust, config.LangCCpp}, // exercises codelldb dedup
		"all-languages": config.AllLanguages,
	}

	names := make([]string, 0, len(combos))
	for name := range combos {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		langs := combos[name]
		t.Run(name, func(t *testing.T) {
			files, err := Render(BuildTemplateData(config.Config{Languages: langs}))
			if err != nil {
				t.Fatalf("Render: %v", err)
			}

			goldenDir := filepath.Join("testdata", "golden", name)

			if *updateGolden {
				if err := os.RemoveAll(goldenDir); err != nil {
					t.Fatalf("RemoveAll: %v", err)
				}
				for _, f := range files {
					path := filepath.Join(goldenDir, filepath.FromSlash(f.RelPath))
					if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
						t.Fatalf("MkdirAll: %v", err)
					}
					if err := os.WriteFile(path, f.Content, 0o644); err != nil {
						t.Fatalf("WriteFile: %v", err)
					}
				}
				return
			}

			for _, f := range files {
				wantPath := filepath.Join(goldenDir, filepath.FromSlash(f.RelPath))
				want, err := os.ReadFile(wantPath)
				if err != nil {
					t.Fatalf("reading golden file %s (run with -update to create it): %v", wantPath, err)
				}
				if string(want) != string(f.Content) {
					t.Errorf("%s: rendered content does not match golden file %s\n--- got ---\n%s\n--- want ---\n%s",
						f.RelPath, wantPath, f.Content, want)
				}
			}
		})
	}
}
