package neovim

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version is a parsed Neovim release version, e.g. from a git tag
// ("v0.10.2") or the first line of `nvim --version` ("NVIM v0.10.2").
type Version struct {
	Major, Minor, Patch int
	Raw                 string
}

var versionPattern = regexp.MustCompile(`v?(\d+)\.(\d+)\.(\d+)`)

// ParseVersion extracts a Version from s.
func ParseVersion(s string) (Version, error) {
	m := versionPattern.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("no version found in %q", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return Version{Major: major, Minor: minor, Patch: patch, Raw: strings.TrimSpace(s)}, nil
}

// Compare returns -1, 0, or 1 as v is less than, equal to, or greater than
// other.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return cmpInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmpInt(v.Minor, other.Minor)
	}
	return cmpInt(v.Patch, other.Patch)
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}
