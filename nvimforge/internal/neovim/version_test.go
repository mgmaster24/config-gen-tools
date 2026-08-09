package neovim

import "testing"

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in                              string
		wantMajor, wantMinor, wantPatch int
		wantErr                         bool
	}{
		{"v0.10.2", 0, 10, 2, false},
		{"0.10.2", 0, 10, 2, false},
		{"NVIM v0.10.2\nBuild type: Release\n", 0, 10, 2, false},
		{"v1.0.0-dev", 1, 0, 0, false},
		{"not a version", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}

	for _, tc := range cases {
		got, err := ParseVersion(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseVersion(%q) expected error, got %v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseVersion(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got.Major != tc.wantMajor || got.Minor != tc.wantMinor || got.Patch != tc.wantPatch {
			t.Errorf("ParseVersion(%q) = %+v, want {%d %d %d}", tc.in, got, tc.wantMajor, tc.wantMinor, tc.wantPatch)
		}
	}
}

func TestVersion_Compare(t *testing.T) {
	v := func(maj, min, pat int) Version { return Version{Major: maj, Minor: min, Patch: pat} }

	cases := []struct {
		a, b Version
		want int
	}{
		{v(0, 10, 2), v(0, 10, 2), 0},
		{v(0, 10, 1), v(0, 10, 2), -1},
		{v(0, 10, 2), v(0, 10, 1), 1},
		{v(0, 9, 9), v(0, 10, 0), -1},
		{v(1, 0, 0), v(0, 99, 99), 1},
	}

	for _, tc := range cases {
		if got := tc.a.Compare(tc.b); got != tc.want {
			t.Errorf("%v.Compare(%v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestVersion_String(t *testing.T) {
	v := Version{Major: 0, Minor: 10, Patch: 2}
	if got := v.String(); got != "v0.10.2" {
		t.Errorf("String() = %q, want %q", got, "v0.10.2")
	}
}
