package cli

import (
	"path/filepath"
	"testing"

	"github.com/mgmaster24/nvimforge/internal/config"
)

func TestResolveDoctorLanguages_LangFlagsWin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")
	writeTestConfig(t, path, config.Config{Languages: []config.Language{config.LangPython}})

	got, err := resolveDoctorLanguages(path, []string{"go", "rust"})
	if err != nil {
		t.Fatalf("resolveDoctorLanguages: %v", err)
	}
	if len(got) != 2 || got[0] != config.LangGo || got[1] != config.LangRust {
		t.Errorf("got %v, want [go rust] (flags should win over file)", got)
	}
}

func TestResolveDoctorLanguages_FallsBackToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")
	writeTestConfig(t, path, config.Config{Languages: []config.Language{config.LangBash, config.LangDockerYAML}})

	got, err := resolveDoctorLanguages(path, nil)
	if err != nil {
		t.Fatalf("resolveDoctorLanguages: %v", err)
	}
	if len(got) != 2 || got[0] != config.LangBash || got[1] != config.LangDockerYAML {
		t.Errorf("got %v, want [bash docker-yaml]", got)
	}
}

func TestResolveDoctorLanguages_NoFileNoFlags_ReturnsNil(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	got, err := resolveDoctorLanguages(nonexistent, nil)
	if err != nil {
		t.Fatalf("resolveDoctorLanguages: %v", err)
	}
	if got != nil {
		t.Errorf("got %v, want nil (universal checks only)", got)
	}
}

func TestResolveDoctorLanguages_InvalidFlagErrors(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	_, err := resolveDoctorLanguages(nonexistent, []string{"cobol"})
	if err == nil {
		t.Fatal("expected an error for an invalid --lang value")
	}
}
