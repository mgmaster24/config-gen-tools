package runner

import "testing"

func TestOSRunner_LookPath(t *testing.T) {
	r := OSRunner{}

	// The go toolchain is guaranteed to be on PATH while `go test` runs.
	if _, found := r.LookPath("go"); !found {
		t.Error(`LookPath("go") should find the go binary running this test`)
	}

	if _, found := r.LookPath("nvimforge-definitely-not-a-real-binary"); found {
		t.Error("LookPath should report false for a binary that doesn't exist")
	}
}

func TestOSRunner_Output(t *testing.T) {
	r := OSRunner{}

	out, err := r.Output("go", "env", "GOOS")
	if err != nil {
		t.Fatalf("Output(go env GOOS) error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty GOOS output")
	}

	if _, err := r.Output("nvimforge-definitely-not-a-real-binary"); err == nil {
		t.Error("Output should error for a nonexistent binary")
	}
}
