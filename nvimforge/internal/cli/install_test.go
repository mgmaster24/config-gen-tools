package cli

import (
	"path/filepath"
	"testing"

	"github.com/mgmaster24/nvimforge/internal/config"
)

func writeTestConfig(t *testing.T, path string, cfg config.Config) {
	t.Helper()
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func failIfPromptCalled(t *testing.T) func(config.Config) (config.Config, error) {
	t.Helper()
	return func(config.Config) (config.Config, error) {
		t.Fatal("Prompt should not have been called")
		return config.Config{}, nil
	}
}

func TestResolveInstallConfig_ExistingFile_NoPrompt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")
	writeTestConfig(t, path, config.Config{
		Languages:  []config.Language{config.LangPython},
		DeployPath: "/existing/path",
		Backup:     true,
		ShowBanner: true,
	})

	cfg, saveNeeded, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: path,
		Prompt:     failIfPromptCalled(t),
	})
	if err != nil {
		t.Fatalf("resolveInstallConfig: %v", err)
	}
	if saveNeeded {
		t.Error("saveNeeded should be false when loading an existing file")
	}
	if len(cfg.Languages) != 1 || cfg.Languages[0] != config.LangPython {
		t.Errorf("Languages = %v, want [python]", cfg.Languages)
	}
	if cfg.DeployPath != "/existing/path" {
		t.Errorf("DeployPath = %q, want /existing/path", cfg.DeployPath)
	}
}

func TestResolveInstallConfig_YesWithoutLangFails(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	_, _, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: nonexistent,
		Yes:        true,
		Prompt:     failIfPromptCalled(t),
	})
	if err == nil {
		t.Fatal("expected an error when --yes is set with no --lang and no config file")
	}
}

func TestResolveInstallConfig_YesWithLangs_NoFile(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	cfg, saveNeeded, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: nonexistent,
		LangFlags:  []string{"go", "rust"},
		Yes:        true,
		Prompt:     failIfPromptCalled(t),
	})
	if err != nil {
		t.Fatalf("resolveInstallConfig: %v", err)
	}
	if saveNeeded {
		t.Error("saveNeeded should be false for a --yes flag-driven run")
	}
	if len(cfg.Languages) != 2 || cfg.Languages[0] != config.LangGo || cfg.Languages[1] != config.LangRust {
		t.Errorf("Languages = %v, want [go rust]", cfg.Languages)
	}
	if cfg.DeployPath != config.DefaultDeployPath {
		t.Errorf("DeployPath = %q, want default %q", cfg.DeployPath, config.DefaultDeployPath)
	}
}

func TestResolveInstallConfig_NoFileNoYes_UsesPrompt(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	promptCalled := false
	cfg, saveNeeded, savePath, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: nonexistent,
		Prompt: func(defaults config.Config) (config.Config, error) {
			promptCalled = true
			defaults.Languages = []config.Language{config.LangLua}
			return defaults, nil
		},
	})
	if err != nil {
		t.Fatalf("resolveInstallConfig: %v", err)
	}
	if !promptCalled {
		t.Error("expected Prompt to be called when no file exists and --yes is not set")
	}
	if !saveNeeded {
		t.Error("saveNeeded should be true after a fresh interactive prompt")
	}
	if savePath != nonexistent {
		t.Errorf("savePath = %q, want %q", savePath, nonexistent)
	}
	if len(cfg.Languages) != 1 || cfg.Languages[0] != config.LangLua {
		t.Errorf("Languages = %v, want [lua]", cfg.Languages)
	}
}

func TestResolveInstallConfig_FlagsOverrideLoadedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")
	writeTestConfig(t, path, config.Config{
		Languages:  []config.Language{config.LangPython},
		DeployPath: "/original/path",
		Backup:     true,
		ShowBanner: true,
	})

	cfg, _, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath:     path,
		LangFlags:      []string{"go"},
		DeployPathFlag: "/overridden/path",
		BackupChanged:  true,
		BackupFlag:     false,
		NoBanner:       true,
		Prompt:         failIfPromptCalled(t),
	})
	if err != nil {
		t.Fatalf("resolveInstallConfig: %v", err)
	}
	if len(cfg.Languages) != 1 || cfg.Languages[0] != config.LangGo {
		t.Errorf("Languages = %v, want [go] (flag should override file)", cfg.Languages)
	}
	if cfg.DeployPath != "/overridden/path" {
		t.Errorf("DeployPath = %q, want /overridden/path", cfg.DeployPath)
	}
	if cfg.Backup {
		t.Error("Backup should be false: --backup=false was Changed")
	}
	if cfg.ShowBanner {
		t.Error("ShowBanner should be false: --no-banner was set")
	}
}

func TestResolveInstallConfig_BackupFlagNotChanged_KeepsFileValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nvimforge.toml")
	writeTestConfig(t, path, config.Config{
		Languages:  []config.Language{config.LangGo},
		DeployPath: "~/.config/nvim",
		Backup:     false,
	})

	cfg, _, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: path,
		// BackupChanged is false (zero value): the loaded file's Backup=false
		// should be preserved, not silently reset to the flag's zero value.
		BackupFlag: true,
		Prompt:     failIfPromptCalled(t),
	})
	if err != nil {
		t.Fatalf("resolveInstallConfig: %v", err)
	}
	if cfg.Backup {
		t.Error("Backup should remain false from the file when --backup wasn't explicitly passed")
	}
}

func TestResolveInstallConfig_InvalidLangFlagErrors(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	_, _, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: nonexistent,
		LangFlags:  []string{"cobol"},
		Yes:        true,
		Prompt:     failIfPromptCalled(t),
	})
	if err == nil {
		t.Fatal("expected an error for an invalid --lang value")
	}
}

func TestResolveInstallConfig_PromptReturningNoLanguages_FailsValidate(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.toml")

	_, _, _, err := resolveInstallConfig(installConfigOptions{
		ConfigPath: nonexistent,
		Prompt: func(defaults config.Config) (config.Config, error) {
			return defaults, nil // no languages selected
		},
	})
	if err == nil {
		t.Fatal("expected Validate() to reject a config with no languages")
	}
}

func TestParseLanguages(t *testing.T) {
	got, err := parseLanguages([]string{"go", "rust"})
	if err != nil {
		t.Fatalf("parseLanguages: %v", err)
	}
	if len(got) != 2 || got[0] != config.LangGo || got[1] != config.LangRust {
		t.Errorf("got %v, want [go rust]", got)
	}

	if _, err := parseLanguages([]string{"not-a-language"}); err == nil {
		t.Error("expected an error for an invalid language")
	}
}
