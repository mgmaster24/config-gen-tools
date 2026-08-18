//go:build integration

// Package integration checks that the script shellforge generates is
// something the target shell can actually parse. The golden tests assert the
// expected *bytes*; these assert the bytes are valid shell.
//
//	go test -tags integration ./...
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/genshell"
)

// shellBin resolves the interpreter for a shell, skipping when it isn't
// installed — zsh in particular isn't present on a stock CI runner.
func shellBin(t *testing.T, shell config.Shell) string {
	t.Helper()
	bin, err := exec.LookPath(string(shell))
	if err != nil {
		t.Skipf("%s not on PATH; skipping", shell)
	}
	return bin
}

func TestGeneratedScriptParses(t *testing.T) {
	combos := map[string]config.Config{
		"zsh-defaults":  {Shell: config.ShellZsh, Integrations: config.DefaultIntegrations},
		"zsh-all":       {Shell: config.ShellZsh, Integrations: config.AllIntegrations},
		"bash-defaults": {Shell: config.ShellBash, Integrations: config.DefaultIntegrations},
		"bash-all":      {Shell: config.ShellBash, Integrations: config.AllIntegrations},
	}

	names := make([]string, 0, len(combos))
	for name := range combos {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cfg := combos[name]
		t.Run(name, func(t *testing.T) {
			bin := shellBin(t, cfg.Shell)

			files, err := genshell.Render(cfg.Shell, genshell.BuildTemplateData(cfg))
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			path := filepath.Join(t.TempDir(), files[0].RelPath)
			if err := os.WriteFile(path, files[0].Content, 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			// -n parses without executing: no hooks run, no network, no
			// mutation of the test environment.
			out, err := exec.Command(bin, "-n", path).CombinedOutput()
			if err != nil {
				t.Fatalf("generated script is not valid %s (%v):\n%s\n--- script ---\n%s",
					cfg.Shell, err, out, files[0].Content)
			}
		})
	}
}

// TestCheckerRejectsInvalidScript guards the guard: if `-n` silently accepted
// anything, TestGeneratedScriptParses would pass while proving nothing.
func TestCheckerRejectsInvalidScript(t *testing.T) {
	bin := shellBin(t, config.ShellBash)

	bad := filepath.Join(t.TempDir(), "bad.sh")
	if err := os.WriteFile(bad, []byte("if true; then\n  echo unterminated\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if out, err := exec.Command(bin, "-n", bad).CombinedOutput(); err == nil {
		t.Fatalf("bash -n accepted an unterminated if; output:\n%s", out)
	}
}
