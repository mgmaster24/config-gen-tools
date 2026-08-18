package genshell

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/mgmaster24/config-gen-tools/forge/fsutil"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
)

// markerName identifies shellforge's own generated output, so regenerating
// doesn't keep backing up a directory shellforge itself wrote.
var markerName = fsutil.MarkerName("shellforge")

// File is one rendered output file, relative to the deploy directory.
type File struct {
	RelPath string
	Content []byte
}

// Render executes the init template against data. Rendering is a pure
// function of data — no timestamps, no filesystem access — which the
// golden-file tests depend on.
func Render(shell config.Shell, data TemplateData) ([]File, error) {
	raw, err := templatesFS.ReadFile("templates/init.sh.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading embedded template: %w", err)
	}
	tmpl, err := template.New("init").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering template: %w", err)
	}
	return []File{{RelPath: shell.InitFileName(), Content: buf.Bytes()}}, nil
}

// Write writes files into deployPath, backing up pre-existing content that
// shellforge did not itself generate. It writes a marker afterwards so a
// re-run recognizes its own prior output.
func Write(files []File, deployPath string, backup bool) (backedUpTo string, err error) {
	if backup {
		needsBackup, err := fsutil.NeedsBackup(deployPath, markerName)
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
		full := filepath.Join(deployPath, filepath.FromSlash(f.RelPath))
		if err := fsutil.AtomicWriteFile(full, f.Content, 0o644); err != nil {
			return backedUpTo, fmt.Errorf("writing %s: %w", f.RelPath, err)
		}
	}

	marker := filepath.Join(deployPath, markerName)
	if err := fsutil.AtomicWriteFile(marker, []byte(`{"generated_by":"shellforge"}`), 0o644); err != nil {
		return backedUpTo, fmt.Errorf("writing generated marker: %w", err)
	}
	return backedUpTo, nil
}

// SourceLine is the single line a user adds to their rc file to activate the
// generated config. Returned so the CLI can print it verbatim.
func SourceLine(shell config.Shell, deployPath string) string {
	target := filepath.ToSlash(filepath.Join(deployPath, shell.InitFileName()))
	return fmt.Sprintf(`[ -f "%s" ] && . "%s"`, target, target)
}
