package config

import "testing"

func ident(name, dir string) Identity {
	return Identity{Name: name, UserName: "A Person", Email: name + "@example.com", Dir: dir}
}

func TestNormalizedDir_AddsTrailingSlash(t *testing.T) {
	cases := map[string]string{
		"~/work":       "~/work/",
		"~/work/":      "~/work/",
		"/src/work//":  "/src/work/",
		`C:\src\work`:  "C:/src/work/",
		"  ~/work  ":   "~/work/",
		"~/work/./sub": "~/work/sub/",
	}
	for in, want := range cases {
		got := Identity{Dir: in}.NormalizedDir()
		if got != want {
			t.Errorf("NormalizedDir(%q) = %q, want %q", in, got, want)
		}
	}
	if got := (Identity{Dir: ""}).NormalizedDir(); got != "" {
		t.Errorf("empty Dir should stay empty, got %q", got)
	}
}

func TestValidate_RequiresExactlyOneDefaultIdentity(t *testing.T) {
	base := func(ids ...Identity) Config {
		return Config{Identities: ids, Features: DefaultFeatures, DeployPath: "~/x"}
	}

	if err := base(ident("work", "~/work")).Validate(); err == nil {
		t.Error("expected an error when no identity is the default")
	}
	if err := base(ident("a", ""), ident("b", "")).Validate(); err == nil {
		t.Error("expected an error when two identities are the default")
	}
	if err := base(ident("a", ""), ident("work", "~/work")).Validate(); err != nil {
		t.Errorf("one default plus one scoped identity should be valid, got %v", err)
	}
}

// Two identities matching the same directory would make the winner depend on
// file order — a silent, confusing failure.
func TestValidate_RejectsOverlappingGitdirs(t *testing.T) {
	cfg := Config{
		Identities: []Identity{
			ident("a", ""),
			ident("work", "~/work"),
			// Same directory once normalized.
			ident("work2", "~/work/"),
		},
		Features:   DefaultFeatures,
		DeployPath: "~/x",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected an error for two identities matching the same gitdir")
	}
}

func TestValidate_RejectsBadIdentities(t *testing.T) {
	cases := map[string]Identity{
		"empty name":     {Name: "", UserName: "A", Email: "a@example.com"},
		"uppercase name": {Name: "Work", UserName: "A", Email: "a@example.com"},
		"empty username": {Name: "work", UserName: "", Email: "a@example.com"},
		"bad email":      {Name: "work", UserName: "A", Email: "not-an-email"},
		"ssh sign without key": {
			Name: "work", UserName: "A", Email: "a@example.com", SSHSign: true,
		},
	}
	for name, id := range cases {
		cfg := Config{Identities: []Identity{id}, Features: DefaultFeatures, DeployPath: "~/x"}
		if err := cfg.Validate(); err == nil {
			t.Errorf("%s: expected a validation error", name)
		}
	}
}

func TestDefault_HasNoIdentitiesButIsOtherwisePopulated(t *testing.T) {
	c := Default()
	// There is no sensible default name/email, so a first run must supply
	// them — unlike nvimforge, where default languages are meaningful.
	if len(c.Identities) != 0 {
		t.Errorf("Default().Identities should be empty, got %v", c.Identities)
	}
	if err := c.Validate(); err == nil {
		t.Error("Default() alone should not validate: it has no identities")
	}
	if len(c.Features) != len(DefaultFeatures) {
		t.Errorf("Features = %v, want %v", c.Features, DefaultFeatures)
	}
	c.Features[0] = "mutated"
	if DefaultFeatures[0] == "mutated" {
		t.Error("Default() aliased DefaultFeatures instead of copying it")
	}
}

func TestParseFeature(t *testing.T) {
	if _, err := ParseFeature("rerere"); err != nil {
		t.Errorf("rerere should parse: %v", err)
	}
	if _, err := ParseFeature("not-a-feature"); err == nil {
		t.Error("expected an error for an unknown feature")
	}
}
