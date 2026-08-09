// Package config defines nvimforge's own TOML configuration file — the
// user's selected languages, where the generated Neovim config gets
// deployed, and small UI preferences. It is distinct from the Neovim
// config nvimforge generates (see internal/genconfig).
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/mgmaster24/nvimforge/internal/fsutil"
)

// DefaultDeployPath is where the generated Neovim config is written unless
// overridden by config file or flag.
const DefaultDeployPath = "~/.config/nvim"

// FileName is the conventional config file name searched for in the
// current directory.
const FileName = "nvimforge.toml"

// Config is nvimforge's own configuration, loaded from / saved to TOML.
type Config struct {
	Languages  []Language `toml:"languages"`
	DeployPath string     `toml:"deploy_path"`
	Backup     bool       `toml:"backup"`
	ShowBanner bool       `toml:"show_banner"`
}

// Default returns the built-in defaults used when no config file exists and
// no flags override a given field. Languages is intentionally empty — v1
// makes no assumption about which languages a new user wants.
func Default() Config {
	return Config{
		Languages:  nil,
		DeployPath: DefaultDeployPath,
		Backup:     true,
		ShowBanner: true,
	}
}

// Validate checks c for internal consistency. It does not touch the
// filesystem.
func (c Config) Validate() error {
	if len(c.Languages) == 0 {
		return errors.New("at least one language must be selected")
	}

	seen := make(map[Language]bool, len(c.Languages))
	for _, l := range c.Languages {
		if !l.Valid() {
			return fmt.Errorf("invalid language %q", l)
		}
		if seen[l] {
			return fmt.Errorf("duplicate language %q", l)
		}
		seen[l] = true
	}

	if c.DeployPath == "" {
		return errors.New("deploy_path must not be empty")
	}

	return nil
}

// ExpandedDeployPath returns DeployPath with a leading "~" resolved to the
// current user's home directory.
func (c Config) ExpandedDeployPath() (string, error) {
	return fsutil.ExpandHome(c.DeployPath)
}

// Load reads and strictly decodes a TOML config file at path, rejecting
// unknown fields so a typo in the file surfaces immediately rather than
// silently falling back to a default.
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

// DefaultUserConfigPath returns the fallback config location used when no
// nvimforge.toml exists in the current directory:
// ~/.config/nvimforge/config.toml.
func DefaultUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "nvimforge", "config.toml"), nil
}

// Resolve determines which config file path to use, given an explicit
// --config flag value (possibly empty). An explicit path is returned
// as-is. Otherwise it prefers ./nvimforge.toml if present, falling back to
// DefaultUserConfigPath even if that file doesn't exist yet (the caller
// decides how to handle a missing file).
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
