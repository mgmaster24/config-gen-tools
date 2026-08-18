// Package config defines gitforge's own TOML configuration: the git
// identities to generate, which opinionated features to enable, and where to
// write the result. It is distinct from the gitconfig gitforge generates
// (internal/gengit).
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/mgmaster24/config-gen-tools/forge/fsutil"
)

// DefaultDeployPath is a directory. gitforge never writes ~/.gitconfig
// itself: it generates includable files and prints the one-line [include] to
// add, so a user's own gitconfig stays theirs.
const DefaultDeployPath = "~/.config/gitforge"

// FileName is the config file searched for in the current directory.
const FileName = "gitforge.toml"

// Config is gitforge's own configuration.
type Config struct {
	Identities []Identity `toml:"identities"`
	Features   []Feature  `toml:"features"`
	DeployPath string     `toml:"deploy_path"`
	Backup     bool       `toml:"backup"`
}

// Default returns the built-in defaults. Identities is intentionally empty:
// unlike a language selection, there is no sensible default name and email,
// so a first run must supply them.
func Default() Config {
	return Config{
		Identities: nil,
		Features:   append([]Feature(nil), DefaultFeatures...),
		DeployPath: DefaultDeployPath,
		Backup:     true,
	}
}

// Validate checks c for internal consistency without touching the filesystem.
func (c Config) Validate() error {
	if len(c.Identities) == 0 {
		return errors.New("at least one identity must be configured")
	}

	seenName := make(map[string]bool, len(c.Identities))
	seenDir := make(map[string]bool, len(c.Identities))
	defaults := 0

	for _, id := range c.Identities {
		if err := id.Validate(); err != nil {
			return err
		}
		if seenName[id.Name] {
			return fmt.Errorf("duplicate identity name %q", id.Name)
		}
		seenName[id.Name] = true

		if id.IsDefault() {
			defaults++
			continue
		}
		// Two identities matching the same directory would make which one
		// wins depend on file order — a confusing, silent failure.
		dir := id.NormalizedDir()
		if seenDir[dir] {
			return fmt.Errorf("identities %q and another both match gitdir %q", id.Name, dir)
		}
		seenDir[dir] = true
	}

	if defaults != 1 {
		return fmt.Errorf("exactly one identity must be the default (empty dir), found %d", defaults)
	}

	seenFeature := make(map[Feature]bool, len(c.Features))
	for _, f := range c.Features {
		if !f.Valid() {
			return fmt.Errorf("invalid feature %q", f)
		}
		if seenFeature[f] {
			return fmt.Errorf("duplicate feature %q", f)
		}
		seenFeature[f] = true
	}

	if c.DeployPath == "" {
		return errors.New("deploy_path must not be empty")
	}
	return nil
}

// DefaultIdentity returns the identity with no gitdir condition. Validate
// guarantees exactly one exists.
func (c Config) DefaultIdentity() (Identity, bool) {
	for _, id := range c.Identities {
		if id.IsDefault() {
			return id, true
		}
	}
	return Identity{}, false
}

// ScopedIdentities returns the non-default identities, i.e. those attached to
// an includeIf gitdir condition.
func (c Config) ScopedIdentities() []Identity {
	var out []Identity
	for _, id := range c.Identities {
		if !id.IsDefault() {
			out = append(out, id)
		}
	}
	return out
}

// HasFeature reports whether f is enabled.
func (c Config) HasFeature(f Feature) bool {
	for _, got := range c.Features {
		if got == f {
			return true
		}
	}
	return false
}

// ExpandedDeployPath resolves a leading "~" to the user's home directory.
func (c Config) ExpandedDeployPath() (string, error) {
	return fsutil.ExpandHome(c.DeployPath)
}

// Load reads and strictly decodes a TOML config, rejecting unknown fields.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config %s: %w", path, err)
	}
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	c := Default()
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return c, nil
}

// Save writes c to path as TOML, creating parent directories as needed.
func Save(path string, c Config) error {
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return fsutil.AtomicWriteFile(path, data, 0o644)
}

// DefaultUserConfigPath is the fallback when no gitforge.toml is in cwd.
func DefaultUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "gitforge", "config.toml"), nil
}

// Resolve determines which config file path to use.
func Resolve(explicitPath string) (path string, existed bool, err error) {
	if explicitPath != "" {
		_, statErr := os.Stat(explicitPath)
		return explicitPath, statErr == nil, nil
	}
	if _, statErr := os.Stat(FileName); statErr == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false, fmt.Errorf("resolving working directory: %w", err)
		}
		return filepath.Join(cwd, FileName), true, nil
	}
	userPath, err := DefaultUserConfigPath()
	if err != nil {
		return "", false, err
	}
	_, statErr := os.Stat(userPath)
	return userPath, statErr == nil, nil
}
