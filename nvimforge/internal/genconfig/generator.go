package genconfig

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/mgmaster24/nvimforge/internal/fsutil"
)

const (
	templateRoot = "templates"
	templateExt  = ".tmpl"
)

// File is one rendered output file, relative to the deploy directory.
type File struct {
	RelPath string
	Content []byte
}

// Render executes every embedded template against data, returning the
// generated files sorted by RelPath. Rendering is a pure function of data:
// no timestamps, no filesystem access, no randomness — required for
// golden-file tests to be stable byte-for-byte.
func Render(data TemplateData) ([]File, error) {
	var tmplPaths []string
	err := fs.WalkDir(templatesFS, templateRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, templateExt) {
			tmplPaths = append(tmplPaths, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking embedded templates: %w", err)
	}
	sort.Strings(tmplPaths)

	files := make([]File, 0, len(tmplPaths))
	for _, tp := range tmplPaths {
		content, err := templatesFS.ReadFile(tp)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", tp, err)
		}
		tmpl, err := template.New(tp).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", tp, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("rendering template %s: %w", tp, err)
		}

		relPath := strings.TrimSuffix(strings.TrimPrefix(tp, templateRoot+"/"), templateExt)
		files = append(files, File{RelPath: relPath, Content: buf.Bytes()})
	}

	return files, nil
}

// Write writes files into deployPath. If backup is true and deployPath
// holds pre-existing content nvimforge didn't itself generate (see
// fsutil.NeedsBackup), that content is preserved first and its new
// location returned as backedUpTo. A hidden marker is written after a
// successful generation so future re-generations of the same deployPath
// are recognized as nvimforge's own prior output rather than backed up
// again.
func Write(files []File, deployPath string, backup bool) (backedUpTo string, err error) {
	if backup {
		needsBackup, err := fsutil.NeedsBackup(deployPath)
		if err != nil {
			return "", err
		}
		if needsBackup {
			backedUpTo, err = fsutil.Backup(deployPath)
			if err != nil {
				return "", err
			}
		}
	}

	for _, f := range files {
		fullPath := filepath.Join(deployPath, filepath.FromSlash(f.RelPath))
		if err := fsutil.AtomicWriteFile(fullPath, f.Content, 0o644); err != nil {
			return backedUpTo, fmt.Errorf("writing %s: %w", f.RelPath, err)
		}
	}

	markerPath := filepath.Join(deployPath, fsutil.GeneratedMarkerName)
	if err := fsutil.AtomicWriteFile(markerPath, []byte(`{"generated_by":"nvimforge"}`), 0o644); err != nil {
		return backedUpTo, fmt.Errorf("writing generated marker: %w", err)
	}

	return backedUpTo, nil
}
