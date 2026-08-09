// Package runnertest provides a shared test double for runner.Runner, used
// by both internal/prereq and internal/neovim's tests so each doesn't
// reinvent its own fake.
package runnertest

import (
	"fmt"
	"strings"
)

// Result stubs the outcome of one Output call.
type Result struct {
	Output string
	Err    error
}

// Fake is an in-memory runner.Runner for tests. A nil or absent entry in
// Paths means the binary is reported not found; a nil or absent entry in
// Outputs means Output returns an error naming the unstubbed command, which
// surfaces missing test setup instead of silently returning a zero value.
type Fake struct {
	Paths   map[string]string
	Outputs map[string]Result
}

// New returns an empty Fake ready to have Paths/Outputs populated.
func New() *Fake {
	return &Fake{Paths: map[string]string{}, Outputs: map[string]Result{}}
}

// WithPath registers binary as found at path (any non-empty path works;
// tests typically just use the binary name itself).
func (f *Fake) WithPath(binary, path string) *Fake {
	f.Paths[binary] = path
	return f
}

// WithOutput stubs the result of Output(name, args...).
func (f *Fake) WithOutput(result Result, name string, args ...string) *Fake {
	f.Outputs[key(name, args)] = result
	return f
}

func (f *Fake) LookPath(binary string) (string, bool) {
	path, ok := f.Paths[binary]
	return path, ok
}

func (f *Fake) Output(name string, args ...string) (string, error) {
	r, ok := f.Outputs[key(name, args)]
	if !ok {
		return "", fmt.Errorf("runnertest: no stubbed output for %q", key(name, args))
	}
	return r.Output, r.Err
}

func key(name string, args []string) string {
	return strings.Join(append([]string{name}, args...), " ")
}
