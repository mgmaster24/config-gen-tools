package runnertest

import "testing"

func TestFake_LookPath(t *testing.T) {
	f := New().WithPath("git", "/usr/bin/git")

	path, found := f.LookPath("git")
	if !found || path != "/usr/bin/git" {
		t.Errorf("LookPath(git) = (%q, %v), want (/usr/bin/git, true)", path, found)
	}

	if _, found := f.LookPath("rg"); found {
		t.Error("LookPath(rg) should be false: never registered")
	}
}

func TestFake_Output(t *testing.T) {
	f := New().WithOutput(Result{Output: "v1.2.3"}, "git", "--version")

	out, err := f.Output("git", "--version")
	if err != nil {
		t.Fatalf("Output: unexpected error %v", err)
	}
	if out != "v1.2.3" {
		t.Errorf("Output = %q, want %q", out, "v1.2.3")
	}

	if _, err := f.Output("git", "status"); err == nil {
		t.Error("Output should error for an unstubbed command+args combination")
	}
}
