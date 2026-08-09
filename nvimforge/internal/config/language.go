package config

import (
	"fmt"
	"slices"
	"strings"
)

// Language identifies one of the toolchains nvimforge can wire into the
// generated Neovim config. It is a closed, finite set — any value not in
// AllLanguages is invalid.
type Language string

const (
	LangRust       Language = "rust"
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangTypeScript Language = "typescript"
	LangLua        Language = "lua"
	LangCCpp       Language = "c-cpp"
	LangBash       Language = "bash"
	LangDockerYAML Language = "docker-yaml"
)

// AllLanguages lists every language nvimforge supports in v1, in the order
// they should be presented to a user (e.g. in the interactive prompt).
var AllLanguages = []Language{
	LangRust,
	LangGo,
	LangPython,
	LangTypeScript,
	LangLua,
	LangCCpp,
	LangBash,
	LangDockerYAML,
}

var languageDisplayNames = map[Language]string{
	LangRust:       "Rust",
	LangGo:         "Go",
	LangPython:     "Python",
	LangTypeScript: "TypeScript",
	LangLua:        "Lua",
	LangCCpp:       "C/C++",
	LangBash:       "Bash",
	LangDockerYAML: "Docker/YAML",
}

// Valid reports whether l is one of AllLanguages.
func (l Language) Valid() bool {
	return slices.Contains(AllLanguages, l)
}

// DisplayName returns a human-readable label for l, falling back to the raw
// value for any (invalid) language not in languageDisplayNames.
func (l Language) DisplayName() string {
	if name, ok := languageDisplayNames[l]; ok {
		return name
	}
	return string(l)
}

// ParseLanguage normalizes and validates a user-supplied language string
// (e.g. from a --lang flag), returning an error listing valid values if it
// doesn't match any known Language.
func ParseLanguage(s string) (Language, error) {
	l := Language(strings.ToLower(strings.TrimSpace(s)))
	if !l.Valid() {
		names := make([]string, len(AllLanguages))
		for i, known := range AllLanguages {
			names[i] = string(known)
		}
		return "", fmt.Errorf("unknown language %q (valid: %s)", s, strings.Join(names, ", "))
	}
	return l, nil
}
