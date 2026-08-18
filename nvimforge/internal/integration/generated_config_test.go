//go:build integration

// Package integration holds end-to-end checks that run a real Neovim against
// the configuration nvimforge generates. They sit behind the `integration`
// build tag because they need `nvim` on PATH, which the rest of the suite
// deliberately does not:
//
//	go test -tags integration ./...
//
// The golden-file tests in internal/genconfig assert the generated config is
// the expected *bytes*. These assert it is something Neovim can actually
// load — a template can render byte-perfect output that is invalid Lua.
package integration

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/genconfig"
)

// checkerLua parses each path passed as an argument without executing it.
// Parsing rather than sourcing is what keeps this tier fast and offline: the
// generated init.lua bootstraps lazy.nvim, so actually running it would clone
// plugins over the network. Verifying that lazy resolves the plugin specs is
// a separate, slower tier.
const checkerLua = `
local failed = false
for _, path in ipairs(_G.arg) do
  local chunk, err = loadfile(path)
  if not chunk then
    io.stderr:write(string.format("%s: %s\n", path, err))
    failed = true
  end
end
if failed then os.exit(1) end
`

func nvimBin(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("nvim")
	if err != nil {
		t.Skip("nvim not on PATH; skipping integration test")
	}
	return bin
}

func writeChecker(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "check.lua")
	if err := os.WriteFile(path, []byte(checkerLua), 0o644); err != nil {
		t.Fatalf("writing checker: %v", err)
	}
	return path
}

// collectLua returns every generated .lua file under root, sorted so failure
// output is stable.
func collectLua(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".lua") {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	sort.Strings(paths)
	return paths
}

func TestGeneratedConfigIsValidLua(t *testing.T) {
	nvim := nvimBin(t)
	checker := writeChecker(t)

	combos := map[string][]config.Language{
		"defaults":      config.DefaultLanguages,
		"all-languages": config.AllLanguages,
		// Exercises a language whose DAP block is unshared, and the only one
		// whose tooling depends on a host toolchain.
		"csharp-only": {config.LangCSharp},
		// Minimal selection: proves the templates render valid Lua even when
		// most conditional blocks are empty.
		"lua-only": {config.LangLua},
	}

	names := make([]string, 0, len(combos))
	for name := range combos {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		langs := combos[name]
		t.Run(name, func(t *testing.T) {
			deploy := t.TempDir()

			files, err := genconfig.Render(genconfig.BuildTemplateData(config.Config{Languages: langs}))
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if _, err := genconfig.Write(files, deploy, false); err != nil {
				t.Fatalf("Write: %v", err)
			}

			luaFiles := collectLua(t, deploy)
			if len(luaFiles) == 0 {
				t.Fatal("no .lua files were generated")
			}

			out, err := exec.Command(nvim, append([]string{"-l", checker}, luaFiles...)...).CombinedOutput()
			if err != nil {
				t.Fatalf("generated config is not valid Lua (%v):\n%s", err, out)
			}
		})
	}
}

// TestCheckerRejectsInvalidLua guards the guard: without it, a checker that
// silently passed everything would make TestGeneratedConfigIsValidLua a
// no-op that still reports success.
func TestCheckerRejectsInvalidLua(t *testing.T) {
	nvim := nvimBin(t)
	checker := writeChecker(t)

	bad := filepath.Join(t.TempDir(), "bad.lua")
	if err := os.WriteFile(bad, []byte("local x = = 1\n"), 0o644); err != nil {
		t.Fatalf("writing bad.lua: %v", err)
	}

	out, err := exec.Command(nvim, "-l", checker, bad).CombinedOutput()
	if err == nil {
		t.Fatalf("checker accepted invalid Lua; output:\n%s", out)
	}
	if !strings.Contains(string(out), "bad.lua") {
		t.Errorf("checker output should name the offending file, got:\n%s", out)
	}
}
