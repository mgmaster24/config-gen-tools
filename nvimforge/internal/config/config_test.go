package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	c := Default()
	if len(c.Languages) != len(DefaultLanguages) {
		t.Errorf("Default().Languages = %v, want %v", c.Languages, DefaultLanguages)
	}
	if err := c.Validate(); err != nil {
		t.Errorf("Default() should be valid on its own, got %v", err)
	}
	// The returned slice must not alias the package-level default, or one
	// caller overwriting cfg.Languages would corrupt it for every other.
	c.Languages[0] = "mutated"
	if DefaultLanguages[0] == "mutated" {
		t.Error("Default() aliased DefaultLanguages instead of copying it")
	}
	if c.DeployPath != DefaultDeployPath {
		t.Errorf("Default().DeployPath = %q, want %q", c.DeployPath, DefaultDeployPath)
	}
	if !c.Backup {
		t.Error("Default().Backup should be true")
	}
	if !c.ShowBanner {
		t.Error("Default().ShowBanner should be true")
	}
}

func TestConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		c       Config
		wantErr bool
	}{
		{
			name:    "valid single language",
			c:       Config{Languages: []Language{LangGo}, DeployPath: "~/.config/nvim"},
			wantErr: false,
		},
		{
			name:    "valid multiple languages",
			c:       Config{Languages: []Language{LangGo, LangRust, LangPython}, DeployPath: "/tmp/nvim"},
			wantErr: false,
		},
		{
			name:    "no languages",
			c:       Config{Languages: nil, DeployPath: "~/.config/nvim"},
			wantErr: true,
		},
		{
			name:    "invalid language",
			c:       Config{Languages: []Language{"cobol"}, DeployPath: "~/.config/nvim"},
			wantErr: true,
		},
		{
			name:    "duplicate language",
			c:       Config{Languages: []Language{LangGo, LangGo}, DeployPath: "~/.config/nvim"},
			wantErr: true,
		},
		{
			name:    "empty deploy path",
			c:       Config{Languages: []Language{LangGo}, DeployPath: ""},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.c.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "nvimforge.toml")

	original := Config{
		Languages:  []Language{LangGo, LangRust, LangDockerYAML},
		DeployPath: "/custom/deploy/path",
		Backup:     false,
		ShowBanner: false,
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Languages) != len(original.Languages) {
		t.Fatalf("Languages = %v, want %v", loaded.Languages, original.Languages)
	}
	for i, l := range original.Languages {
		if loaded.Languages[i] != l {
			t.Errorf("Languages[%d] = %q, want %q", i, loaded.Languages[i], l)
		}
	}
	if loaded.DeployPath != original.DeployPath {
		t.Errorf("DeployPath = %q, want %q", loaded.DeployPath, original.DeployPath)
	}
	if loaded.Backup != original.Backup {
		t.Errorf("Backup = %v, want %v", loaded.Backup, original.Backup)
	}
	if loaded.ShowBanner != original.ShowBanner {
		t.Errorf("ShowBanner = %v, want %v", loaded.ShowBanner, original.ShowBanner)
	}
}

func TestLoad_RejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")

	content := `
languages = ["go"]
deploy_path = "~/.config/nvim"
backup = true
show_banner = true
not_a_real_field = "oops"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}

func TestLoad_RejectsInvalidLanguage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")

	content := `
languages = ["go", "cobol"]
deploy_path = "~/.config/nvim"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed (parsing is not validation): %v", err)
	}
	if err := c.Validate(); err == nil {
		t.Error("Validate() should reject the invalid language \"cobol\"")
	}
}

func TestExpandedDeployPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	c := Config{DeployPath: "~/.config/nvim"}
	got, err := c.ExpandedDeployPath()
	if err != nil {
		t.Fatalf("ExpandedDeployPath: %v", err)
	}
	want := filepath.Join(home, ".config/nvim")
	if got != want {
		t.Errorf("ExpandedDeployPath() = %q, want %q", got, want)
	}
}

func TestResolve_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.toml")

	// Explicit path that doesn't exist yet: returned as-is, existed=false.
	gotPath, existed, err := Resolve(path)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if gotPath != path {
		t.Errorf("path = %q, want %q", gotPath, path)
	}
	if existed {
		t.Error("existed should be false for a nonexistent explicit path")
	}

	// Now create it: existed should flip to true.
	if err := os.WriteFile(path, []byte("languages = []\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, existed, err = Resolve(path)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !existed {
		t.Error("existed should be true once the explicit path exists")
	}
}

func TestResolve_PrefersCwdFile(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.WriteFile(FileName, []byte("languages = []\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	path, existed, err := Resolve("")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !existed {
		t.Error("existed should be true for the cwd nvimforge.toml")
	}
	if !strings.HasSuffix(path, FileName) {
		t.Errorf("path = %q, want suffix %q", path, FileName)
	}
}
