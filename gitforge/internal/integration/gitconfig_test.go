//go:build integration

// Package integration checks the generated gitconfig against real git. The
// golden tests assert the expected bytes; these assert git parses those bytes
// and, crucially, that the includeIf conditions actually resolve.
//
//	go test -tags integration ./...
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgmaster24/config-gen-tools/gitforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/gitforge/internal/gengit"
)

func gitBin(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not on PATH; skipping")
	}
	return bin
}

// generate writes cfg's output into a temp deploy dir and returns its path.
func generate(t *testing.T, cfg config.Config, deployPath string) {
	t.Helper()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	files, err := gengit.Render(cfg, gengit.BuildTemplateData(cfg, deployPath))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if _, err := gengit.Write(files, deployPath, false); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// gitConfigIn runs `git config <key>` inside dir, with the generated base
// file as the global config. GIT_CONFIG_GLOBAL isolates the test completely
// from the developer's real ~/.gitconfig.
func gitConfigIn(t *testing.T, bin, dir, base, key string) string {
	t.Helper()
	cmd := exec.Command(bin, "config", "--get", key)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+base,
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git config --get %s in %s: %v\n%s", key, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func initRepo(t *testing.T, bin, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if out, err := exec.Command(bin, "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
}

// The headline behaviour: a repository under a scoped identity's directory
// must resolve that identity's email, and one outside it must fall back to
// the default.
func TestIncludeIfSelectsIdentityByDirectory(t *testing.T) {
	bin := gitBin(t)
	root := t.TempDir()
	deploy := filepath.Join(root, "deploy")
	workDir := filepath.Join(root, "work")
	personalDir := filepath.Join(root, "personal")

	cfg := config.Config{
		Identities: []config.Identity{
			{Name: "default", UserName: "Personal Name", Email: "me@personal.example"},
			// Deliberately no trailing slash: NormalizedDir must add it, or
			// git matches only this exact path and never the repos beneath.
			{Name: "work", UserName: "Work Name", Email: "me@work.example", Dir: workDir},
		},
		Features:   config.DefaultFeatures,
		DeployPath: deploy,
	}
	generate(t, cfg, deploy)

	base := filepath.Join(deploy, gengit.BaseFileName)
	workRepo := filepath.Join(workDir, "some-repo")
	personalRepo := filepath.Join(personalDir, "some-repo")
	initRepo(t, bin, workRepo)
	initRepo(t, bin, personalRepo)

	if got := gitConfigIn(t, bin, workRepo, base, "user.email"); got != "me@work.example" {
		t.Errorf("inside %s: user.email = %q, want the work identity", workRepo, got)
	}
	if got := gitConfigIn(t, bin, personalRepo, base, "user.email"); got != "me@personal.example" {
		t.Errorf("outside the work tree: user.email = %q, want the default identity", got)
	}
}

// Guards the guard: if the trailing slash were dropped, the test above would
// be the thing that catches it — so prove a slash-less gitdir really does
// fail to match, i.e. that the normalization is load-bearing.
func TestGitdirWithoutTrailingSlashDoesNotMatchSubdirectories(t *testing.T) {
	bin := gitBin(t)
	root := t.TempDir()
	deploy := filepath.Join(root, "deploy")
	workDir := filepath.Join(root, "work")

	if err := os.MkdirAll(deploy, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	identity := filepath.Join(deploy, "identity.work.gitconfig")
	if err := os.WriteFile(identity, []byte("[user]\n\temail = me@work.example\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Hand-written base with the mistake gitforge exists to prevent.
	base := filepath.Join(deploy, "gitconfig")
	content := "[user]\n\temail = me@personal.example\n\n" +
		"[includeIf \"gitdir:" + workDir + "\"]\n\tpath = " + identity + "\n"
	if err := os.WriteFile(base, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	repo := filepath.Join(workDir, "some-repo")
	initRepo(t, bin, repo)

	if got := gitConfigIn(t, bin, repo, base, "user.email"); got != "me@personal.example" {
		t.Errorf("a gitdir without a trailing slash matched %s (got %q) — "+
			"if git changed this, NormalizedDir's rationale needs revisiting", repo, got)
	}
}

func TestGeneratedConfigParses(t *testing.T) {
	bin := gitBin(t)
	root := t.TempDir()
	deploy := filepath.Join(root, "deploy")

	cfg := config.Config{
		Identities: []config.Identity{
			{Name: "default", UserName: "A", Email: "a@example.com"},
			{Name: "work", UserName: "B", Email: "b@example.com", Dir: filepath.Join(root, "work")},
			{Name: "oss", UserName: "C", Email: "c@example.com", Dir: filepath.Join(root, "oss"),
				SigningKey: "~/.ssh/id_ed25519.pub", SSHSign: true},
		},
		Features:   config.AllFeatures,
		DeployPath: deploy,
	}
	generate(t, cfg, deploy)

	entries, err := os.ReadDir(deploy)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	checked := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), "gitconfig") {
			continue
		}
		path := filepath.Join(deploy, e.Name())
		out, err := exec.Command(bin, "config", "--file", path, "--list").CombinedOutput()
		if err != nil {
			t.Errorf("git cannot parse %s: %v\n%s", e.Name(), err, out)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no gitconfig files were generated")
	}
}
