package neovim

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mgmaster24/nvimforge/internal/github"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

// downloadAndExtract fetches asset's archive, verifies its checksum when a
// sibling "<asset.Name>.sha256sum" asset is published in the same release
// (a convention, not a guarantee — its absence is not treated as an
// error, since Neovim's release format has varied across versions), and
// extracts the archive into destDir.
func (i *Installer) downloadAndExtract(ctx context.Context, release github.Release, asset github.Asset, destDir string) error {
	tmpFile, err := os.CreateTemp("", "nvimforge-nvim-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	sum := sha256.New()
	if err := downloadTo(ctx, asset.BrowserDownloadURL, io.MultiWriter(tmpFile, sum)); err != nil {
		tmpFile.Close()
		return fmt.Errorf("downloading %s: %w", asset.Name, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing downloaded file: %w", err)
	}

	if wantSum, ok, err := expectedChecksum(ctx, release, asset); err != nil {
		return err
	} else if ok {
		gotSum := hex.EncodeToString(sum.Sum(nil))
		if !strings.EqualFold(gotSum, wantSum) {
			return fmt.Errorf("checksum mismatch for %s: got %s, want %s", asset.Name, gotSum, wantSum)
		}
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", destDir, err)
	}

	switch {
	case strings.HasSuffix(asset.Name, ".zip"):
		return extractZip(tmpPath, destDir)
	case strings.HasSuffix(asset.Name, ".tar.gz"), strings.HasSuffix(asset.Name, ".tgz"):
		return extractTarGz(tmpPath, destDir)
	default:
		return fmt.Errorf("don't know how to extract asset %q (unrecognized extension)", asset.Name)
	}
}

func downloadTo(ctx context.Context, url string, dst io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	_, err = io.Copy(dst, resp.Body)
	return err
}

// expectedChecksum looks for "<asset.Name>.sha256sum" among release's other
// assets and, if present, downloads and parses it. The file may contain
// either a bare hex digest or the standard `sha256sum` output format
// ("<digest>  <filename>").
func expectedChecksum(ctx context.Context, release github.Release, asset github.Asset) (sum string, found bool, err error) {
	wantName := asset.Name + ".sha256sum"
	var checksumAsset github.Asset
	for _, a := range release.Assets {
		if a.Name == wantName {
			checksumAsset = a
			found = true
			break
		}
	}
	if !found {
		return "", false, nil
	}

	var buf strings.Builder
	if err := downloadTo(ctx, checksumAsset.BrowserDownloadURL, &buf); err != nil {
		return "", false, fmt.Errorf("downloading checksum for %s: %w", asset.Name, err)
	}
	fields := strings.Fields(buf.String())
	if len(fields) == 0 {
		return "", false, fmt.Errorf("checksum file for %s is empty", asset.Name)
	}
	return fields[0], true, nil
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip stream: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar entry: %w", err)
		}

		target, err := safeJoin(destDir, stripTopLevel(hdr.Name))
		if err != nil {
			return err
		}
		if target == "" {
			continue // top-level directory entry itself
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			// Neovim's own release archives don't ship symlinks in
			// practice; skip rather than risk writing outside destDir.
		}
	}
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		target, err := safeJoin(destDir, stripTopLevel(f.Name))
		if err != nil {
			return err
		}
		if target == "" {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

// stripTopLevel drops the first path segment from a tar/zip entry name.
// Neovim's release archives wrap everything in a single top-level
// directory (e.g. "nvim-macos-arm64/bin/nvim"); nvimforge wants that
// content directly under destDir instead.
func stripTopLevel(name string) string {
	name = filepath.ToSlash(name)
	if _, rest, ok := strings.Cut(name, "/"); ok {
		return rest
	}
	return ""
}

// safeJoin joins destDir and rel, rejecting any result that would escape
// destDir (a zip/tar-slip guard).
func safeJoin(destDir, rel string) (string, error) {
	if rel == "" {
		return "", nil
	}
	target := filepath.Join(destDir, rel)
	if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes destination directory", rel)
	}
	return target, nil
}

// linkBin exposes the just-installed Neovim binary at BinDir/nvim(.exe).
// It tries a symlink first; if that fails (most commonly on Windows
// without Developer Mode/admin privileges), it falls back to copying the
// binary so the install still succeeds.
func (i *Installer) linkBin(v Version) error {
	if err := os.MkdirAll(i.BinDir, 0o755); err != nil {
		return err
	}

	src := filepath.Join(i.versionDir(v), "bin", i.binaryName())
	dst := filepath.Join(i.BinDir, i.binaryName())

	os.Remove(dst)
	if err := os.Symlink(src, dst); err == nil {
		return nil
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
