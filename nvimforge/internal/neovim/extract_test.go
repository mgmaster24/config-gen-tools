package neovim

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func makeTarGz(t *testing.T, entries map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for name, content := range entries {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("WriteHeader: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar Close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip Close: %v", err)
	}

	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func makeZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip Close: %v", err)
	}

	path := filepath.Join(t.TempDir(), "archive.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestExtractTarGz_StripsTopLevelDir(t *testing.T) {
	archive := makeTarGz(t, map[string]string{
		"nvim-linux-x86_64/bin/nvim":          "binary-content",
		"nvim-linux-x86_64/share/nvim/README": "readme-content",
	})
	destDir := t.TempDir()

	if err := extractTarGz(archive, destDir); err != nil {
		t.Fatalf("extractTarGz: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "bin", "nvim"))
	if err != nil {
		t.Fatalf("ReadFile bin/nvim: %v", err)
	}
	if string(got) != "binary-content" {
		t.Errorf("bin/nvim content = %q", got)
	}

	got, err = os.ReadFile(filepath.Join(destDir, "share", "nvim", "README"))
	if err != nil {
		t.Fatalf("ReadFile share/nvim/README: %v", err)
	}
	if string(got) != "readme-content" {
		t.Errorf("README content = %q", got)
	}
}

func TestExtractZip_StripsTopLevelDir(t *testing.T) {
	archive := makeZip(t, map[string]string{
		"nvim-win64/bin/nvim.exe": "exe-content",
	})
	destDir := t.TempDir()

	if err := extractZip(archive, destDir); err != nil {
		t.Fatalf("extractZip: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "bin", "nvim.exe"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "exe-content" {
		t.Errorf("content = %q", got)
	}
}

func TestSafeJoin_RejectsPathTraversal(t *testing.T) {
	dest := t.TempDir()
	_, err := safeJoin(dest, "../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for a path-traversal entry")
	}
}

func TestSafeJoin_AllowsNormalRelativePath(t *testing.T) {
	dest := t.TempDir()
	got, err := safeJoin(dest, "bin/nvim")
	if err != nil {
		t.Fatalf("safeJoin: %v", err)
	}
	want := filepath.Join(dest, "bin", "nvim")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripTopLevel(t *testing.T) {
	cases := map[string]string{
		"nvim-macos-arm64/bin/nvim": "bin/nvim",
		"nvim-macos-arm64/":         "",
		"nvim-macos-arm64":          "",
		"top/mid/leaf.txt":          "mid/leaf.txt",
	}
	for in, want := range cases {
		if got := stripTopLevel(in); got != want {
			t.Errorf("stripTopLevel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")

	if err := os.WriteFile(src, []byte("payload"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "payload" {
		t.Errorf("content = %q, want %q", got, "payload")
	}
}
