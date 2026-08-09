package config

import "testing"

func TestLanguage_Valid(t *testing.T) {
	for _, l := range AllLanguages {
		if !l.Valid() {
			t.Errorf("%q should be valid (it's in AllLanguages)", l)
		}
	}
	if Language("cobol").Valid() {
		t.Error(`"cobol" should not be valid`)
	}
	if Language("").Valid() {
		t.Error(`"" should not be valid`)
	}
}

func TestLanguage_DisplayName(t *testing.T) {
	if got := LangRust.DisplayName(); got != "Rust" {
		t.Errorf("DisplayName() = %q, want %q", got, "Rust")
	}
	// Every known language must have a display name entry, or DisplayName
	// silently falls back to the raw value — catch that drift here.
	for _, l := range AllLanguages {
		if got := l.DisplayName(); got == string(l) {
			t.Errorf("%q has no entry in languageDisplayNames (falls back to raw value)", l)
		}
	}
	if got := Language("unknown-lang").DisplayName(); got != "unknown-lang" {
		t.Errorf("DisplayName() fallback = %q, want raw value %q", got, "unknown-lang")
	}
}

func TestParseLanguage(t *testing.T) {
	cases := []struct {
		in      string
		want    Language
		wantErr bool
	}{
		{"go", LangGo, false},
		{"Go", LangGo, false},
		{"  go  ", LangGo, false},
		{"c-cpp", LangCCpp, false},
		{"cobol", "", true},
		{"", "", true},
	}

	for _, tc := range cases {
		got, err := ParseLanguage(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseLanguage(%q) expected error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseLanguage(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseLanguage(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
