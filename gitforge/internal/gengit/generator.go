// Package gengit renders a set of includable gitconfig files from a
// gitforge configuration: one base file carrying shared settings and the
// includeIf conditions, plus one file per scoped identity.
package gengit

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/mgmaster24/config-gen-tools/forge/fsutil"
	"github.com/mgmaster24/config-gen-tools/gitforge/internal/config"
)

//go:embed templates
var templatesFS embed.FS

// markerName identifies gitforge's own generated output.
var markerName = fsutil.MarkerName("gitforge")

// BaseFileName is the file a user includes from their own ~/.gitconfig.
const BaseFileName = "gitconfig"

// File is one rendered output file, relative to the deploy directory.
type File struct {
	RelPath string
	Content []byte
}

// identityFileName is the per-identity file included by a gitdir condition.
func identityFileName(name string) string {
	return "identity." + name + ".gitconfig"
}

// IdentityData is one scoped identity as the base template sees it.
type IdentityData struct {
	Name string
	Dir  string
	Path string
}

// TemplateData is everything the templates render from. It holds no
// non-deterministic content, which the golden tests depend on.
type TemplateData struct {
	UserName   string
	Email      string
	SigningKey string
	SSHSign    bool
	Identities []IdentityData

	Delta         bool
	Rerere        bool
	AutoStash     bool
	Prune         bool
	RebaseOnPull  bool
	DefaultBranch bool
	Zdiff3        bool
}

// BuildTemplateData resolves cfg into deterministic template input.
// deployPath is needed because includeIf paths must be absolute (or ~-based)
// — git resolves a relative include path against the including file, which
// is the user's ~/.gitconfig, not ours.
func BuildTemplateData(cfg config.Config, deployPath string) TemplateData {
	def, _ := cfg.DefaultIdentity()

	data := TemplateData{
		UserName:      def.UserName,
		Email:         def.Email,
		SigningKey:    def.SigningKey,
		SSHSign:       def.SSHSign,
		Delta:         cfg.HasFeature(config.FeatureDelta),
		Rerere:        cfg.HasFeature(config.FeatureRerere),
		AutoStash:     cfg.HasFeature(config.FeatureAutoStash),
		Prune:         cfg.HasFeature(config.FeaturePruneOnFetch),
		RebaseOnPull:  cfg.HasFeature(config.FeatureRebaseOnPull),
		DefaultBranch: cfg.HasFeature(config.FeatureDefaultBranch),
		Zdiff3:        cfg.HasFeature(config.FeatureZdiff3),
	}

	scoped := cfg.ScopedIdentities()
	sort.Slice(scoped, func(a, b int) bool { return scoped[a].Name < scoped[b].Name })
	for _, id := range scoped {
		data.Identities = append(data.Identities, IdentityData{
			Name: id.Name,
			Dir:  id.NormalizedDir(),
			Path: path.Join(filepath.ToSlash(deployPath), identityFileName(id.Name)),
		})
	}
	return data
}

func render(name string, data any) ([]byte, error) {
	raw, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, fmt.Errorf("reading embedded template %s: %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// Render produces the base gitconfig plus one file per scoped identity,
// sorted by RelPath for deterministic output.
func Render(cfg config.Config, data TemplateData) ([]File, error) {
	base, err := render("gitconfig.tmpl", data)
	if err != nil {
		return nil, err
	}
	files := []File{{RelPath: BaseFileName, Content: base}}

	for _, id := range cfg.ScopedIdentities() {
		content, err := render("identity.gitconfig.tmpl", TemplateData{
			UserName:   id.UserName,
			Email:      id.Email,
			SigningKey: id.SigningKey,
			SSHSign:    id.SSHSign,
		})
		if err != nil {
			return nil, err
		}
		files = append(files, File{RelPath: identityFileName(id.Name), Content: content})
	}

	sort.Slice(files, func(a, b int) bool { return files[a].RelPath < files[b].RelPath })
	return files, nil
}

// Write writes files into deployPath, backing up pre-existing content
// gitforge did not itself generate.
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
	if err := fsutil.AtomicWriteFile(marker, []byte(`{"generated_by":"gitforge"}`), 0o644); err != nil {
		return backedUpTo, fmt.Errorf("writing generated marker: %w", err)
	}
	return backedUpTo, nil
}

// IncludeSnippet is what the user adds to their own ~/.gitconfig.
func IncludeSnippet(deployPath string) string {
	target := path.Join(filepath.ToSlash(deployPath), BaseFileName)
	return fmt.Sprintf("[include]\n\tpath = %s", target)
}
