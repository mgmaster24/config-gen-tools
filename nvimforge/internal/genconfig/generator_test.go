package genconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
)

func TestRender_ProducesSortedRelPaths(t *testing.T) {
	files, err := Render(BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo}}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("Render produced no files")
	}
	for i := 1; i < len(files); i++ {
		if files[i-1].RelPath >= files[i].RelPath {
			t.Errorf("files not sorted by RelPath: %q >= %q", files[i-1].RelPath, files[i].RelPath)
		}
	}
	// Every file must come from a real, expected location.
	wantSome := map[string]bool{
		"init.lua":                    true,
		"lua/config/options.lua":      true,
		"lua/plugins/snacks.lua":      true,
		"lua/plugins/lsp.lua":         true,
		"lua/plugins/dap.lua":         true,
		"lua/plugins/colorscheme.lua": true,
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.RelPath] = true
	}
	for want := range wantSome {
		if !got[want] {
			t.Errorf("expected generated file %q, not found in %v", want, got)
		}
	}
}

func TestRender_IsDeterministic(t *testing.T) {
	data := BuildTemplateData(config.Config{Languages: config.AllLanguages})
	a, err := Render(data)
	if err != nil {
		t.Fatalf("Render (a): %v", err)
	}
	b, err := Render(data)
	if err != nil {
		t.Fatalf("Render (b): %v", err)
	}
	if len(a) != len(b) {
		t.Fatalf("got %d files then %d files", len(a), len(b))
	}
	for i := range a {
		if a[i].RelPath != b[i].RelPath || string(a[i].Content) != string(b[i].Content) {
			t.Errorf("Render is not deterministic at index %d: %q vs %q", i, a[i].RelPath, b[i].RelPath)
		}
	}
}

func TestWrite_FreshDeploy_NoBackup(t *testing.T) {
	dir := t.TempDir()
	deployPath := filepath.Join(dir, "nvim")

	files, err := Render(BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo}}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	backedUpTo, err := Write(files, deployPath, true)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if backedUpTo != "" {
		t.Errorf("backedUpTo = %q, want empty (nothing to back up)", backedUpTo)
	}

	if _, err := os.Stat(filepath.Join(deployPath, "init.lua")); err != nil {
		t.Errorf("init.lua not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(deployPath, markerName)); err != nil {
		t.Errorf("marker not written: %v", err)
	}
}

func TestWrite_BacksUpPreExistingConfig(t *testing.T) {
	dir := t.TempDir()
	deployPath := filepath.Join(dir, "nvim")

	if err := os.MkdirAll(deployPath, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deployPath, "init.lua"), []byte("-- hand written"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := Render(BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo}}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	backedUpTo, err := Write(files, deployPath, true)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if backedUpTo == "" {
		t.Fatal("expected a non-empty backedUpTo for a pre-existing hand-written config")
	}

	got, err := os.ReadFile(filepath.Join(backedUpTo, "init.lua"))
	if err != nil {
		t.Fatalf("reading backed-up init.lua: %v", err)
	}
	if string(got) != "-- hand written" {
		t.Errorf("backed-up content = %q, want the original hand-written content", got)
	}

	// New deployPath should hold the freshly generated content, not the
	// hand-written original.
	got, err = os.ReadFile(filepath.Join(deployPath, "init.lua"))
	if err != nil {
		t.Fatalf("reading new init.lua: %v", err)
	}
	if string(got) == "-- hand written" {
		t.Error("deployPath still has the old hand-written content after Write")
	}
}

func TestWrite_ReGeneration_SkipsBackup(t *testing.T) {
	dir := t.TempDir()
	deployPath := filepath.Join(dir, "nvim")

	files, err := Render(BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo}}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if _, err := Write(files, deployPath, true); err != nil {
		t.Fatalf("first Write: %v", err)
	}

	// Re-run with a different language selection: this is nvimforge
	// regenerating its own output, so no backup should be created even
	// though the content differs from before.
	files2, err := Render(BuildTemplateData(config.Config{Languages: []config.Language{config.LangRust}}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	backedUpTo, err := Write(files2, deployPath, true)
	if err != nil {
		t.Fatalf("second Write: %v", err)
	}
	if backedUpTo != "" {
		t.Errorf("backedUpTo = %q, want empty on re-generation of nvimforge's own output", backedUpTo)
	}
}

func TestWrite_BackupDisabled(t *testing.T) {
	dir := t.TempDir()
	deployPath := filepath.Join(dir, "nvim")

	if err := os.MkdirAll(deployPath, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deployPath, "init.lua"), []byte("-- hand written"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := Render(BuildTemplateData(config.Config{Languages: []config.Language{config.LangGo}}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	backedUpTo, err := Write(files, deployPath, false)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if backedUpTo != "" {
		t.Errorf("backedUpTo = %q, want empty when backup=false", backedUpTo)
	}
}
