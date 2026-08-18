// Package config defines shellforge's own TOML configuration: which shell to
// generate for, which integrations to enable, and where to write the result.
// It is distinct from the shell config shellforge generates (internal/genshell).
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

// DefaultDeployPath is a directory, not an rc file. shellforge deliberately
// never writes to ~/.zshrc: it generates a self-contained init script and
// prints the one line to source it, so a user's own rc file stays theirs.
const DefaultDeployPath = "~/.config/shellforge"

// FileName is the config file searched for in the current directory.
const FileName = "shellforge.toml"

// Config is shellforge's own configuration.
type Config struct {
	Shell        Shell         `toml:"shell"`
	Integrations []Integration `toml:"integrations"`
	DeployPath   string        `toml:"deploy_path"`
	Backup       bool          `toml:"backup"`
}

// Default returns the built-in defaults. Integrations is copied so callers
// can overwrite it without mutating the package-level default.
func Default() Config {
	return Config{
		Shell:        ShellZsh,
		Integrations: append([]Integration(nil), DefaultIntegrations...),
		DeployPath:   DefaultDeployPath,
		Backup:       true,
	}
}

// Validate checks c for internal consistency without touching the filesystem.
func (c Config) Validate() error {
	if !c.Shell.Valid() {
		return fmt.Errorf("invalid shell %q", c.Shell)
	}
	if len(c.Integrations) == 0 {
		return errors.New("at least one integration must be selected")
	}
	seen := make(map[Integration]bool, len(c.Integrations))
	for _, i := range c.Integrations {
		if !i.Valid() {
			return fmt.Errorf("invalid integration %q", i)
		}
		if seen[i] {
			return fmt.Errorf("duplicate integration %q", i)
		}
		seen[i] = true
	}
	if c.DeployPath == "" {
		return errors.New("deploy_path must not be empty")
	}
	return nil
}

// ExpandedDeployPath resolves a leading "~" to the user's home directory.
func (c Config) ExpandedDeployPath() (string, error) {
	return fsutil.ExpandHome(c.DeployPath)
}

// Load reads and strictly decodes a TOML config, rejecting unknown fields so
// a typo surfaces immediately instead of silently falling back to a default.
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

// DefaultUserConfigPath is the fallback when no shellforge.toml is in cwd.
func DefaultUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "shellforge", "config.toml"), nil
}

// Resolve determines which config file path to use, preferring an explicit
// flag, then ./shellforge.toml, then the user config path.
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
