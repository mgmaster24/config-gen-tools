package config

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// Identity is one git author identity. The default identity has an empty
// Dir; every other identity is activated by an includeIf gitdir condition.
type Identity struct {
	// Name is the identity's key (e.g. "work"), used for its filename and in
	// prompts — not the git user.name.
	Name string `toml:"name"`
	// UserName and Email are what land in [user].
	UserName string `toml:"user_name"`
	Email    string `toml:"email"`
	// Dir is the gitdir prefix that activates this identity, e.g. "~/work/".
	// Empty means this is the default identity.
	Dir string `toml:"dir"`
	// SigningKey is an SSH public key path or a GPG key id. Empty disables
	// signing for this identity.
	SigningKey string `toml:"signing_key"`
	// SSHSign selects gpg.format=ssh rather than openpgp.
	SSHSign bool `toml:"ssh_sign"`
}

var identityNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// IsDefault reports whether this is the fallback identity (no gitdir
// condition).
func (i Identity) IsDefault() bool { return strings.TrimSpace(i.Dir) == "" }

// NormalizedDir returns Dir in the form git actually matches on.
//
// This is the whole reason to generate gitconfig rather than hand-write it:
// a gitdir condition without a trailing slash matches only that exact path,
// not paths beneath it, so `gitdir:~/work` silently fails to apply to
// ~/work/some-repo. Git's own docs call this out and people get it wrong
// constantly. Normalizing here means it cannot be got wrong.
func (i Identity) NormalizedDir() string {
	d := strings.TrimSpace(i.Dir)
	if d == "" {
		return ""
	}
	d = path.Clean(strings.ReplaceAll(d, `\`, "/"))
	if !strings.HasSuffix(d, "/") {
		d += "/"
	}
	return d
}

// Validate checks a single identity in isolation.
func (i Identity) Validate() error {
	if !identityNamePattern.MatchString(i.Name) {
		return fmt.Errorf("identity name %q must be lowercase alphanumeric with dashes", i.Name)
	}
	if strings.TrimSpace(i.UserName) == "" {
		return fmt.Errorf("identity %q: user_name must not be empty", i.Name)
	}
	if !strings.Contains(i.Email, "@") {
		return fmt.Errorf("identity %q: email %q does not look like an address", i.Name, i.Email)
	}
	if i.SSHSign && i.SigningKey == "" {
		return fmt.Errorf("identity %q: ssh_sign is set but signing_key is empty", i.Name)
	}
	return nil
}

// Feature is an opinionated git behaviour gitforge can enable.
type Feature string

const (
	FeatureDelta         Feature = "delta"
	FeatureRerere        Feature = "rerere"
	FeatureAutoStash     Feature = "autostash"
	FeaturePruneOnFetch  Feature = "prune"
	FeatureRebaseOnPull  Feature = "rebase-on-pull"
	FeatureDefaultBranch Feature = "default-branch-main"
	FeatureZdiff3        Feature = "zdiff3"
)

var AllFeatures = []Feature{
	FeatureDelta,
	FeatureRerere,
	FeatureAutoStash,
	FeaturePruneOnFetch,
	FeatureRebaseOnPull,
	FeatureDefaultBranch,
	FeatureZdiff3,
}

// DefaultFeatures excludes delta, which needs a binary that may not be
// installed; the rest are pure git config with no external dependency.
var DefaultFeatures = []Feature{
	FeatureRerere,
	FeatureAutoStash,
	FeaturePruneOnFetch,
	FeatureRebaseOnPull,
	FeatureDefaultBranch,
	FeatureZdiff3,
}

var featureDescriptions = map[Feature]string{
	FeatureDelta:         "use delta as the diff pager",
	FeatureRerere:        "remember and reuse conflict resolutions",
	FeatureAutoStash:     "auto-stash local changes when rebasing",
	FeaturePruneOnFetch:  "prune deleted remote branches on fetch",
	FeatureRebaseOnPull:  "rebase instead of merge on pull",
	FeatureDefaultBranch: "name new repositories' first branch main",
	FeatureZdiff3:        "use the zdiff3 conflict style",
}

func (f Feature) Valid() bool {
	_, ok := featureDescriptions[f]
	return ok
}

func (f Feature) Description() string { return featureDescriptions[f] }

// ParseFeature normalizes and validates a user-supplied feature name.
func ParseFeature(v string) (Feature, error) {
	f := Feature(strings.ToLower(strings.TrimSpace(v)))
	if !f.Valid() {
		names := make([]string, len(AllFeatures))
		for i, known := range AllFeatures {
			names[i] = string(known)
		}
		return "", fmt.Errorf("unknown feature %q (valid: %s)", v, strings.Join(names, ", "))
	}
	return f, nil
}
