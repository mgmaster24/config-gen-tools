// Package runner abstracts the small slice of os/exec that forge tools need
// (probing PATH, running a command for its output) behind an interface, so
// prereq detection and the Neovim installer can be unit tested against a
// fake instead of the real system.
package runner

import "os/exec"

// Runner shells out to and probes the host system.
type Runner interface {
	// LookPath reports the resolved path to binary if it exists on PATH.
	LookPath(binary string) (path string, found bool)

	// Output runs name with args and returns combined stdout+stderr.
	Output(name string, args ...string) (output string, err error)
}

// OSRunner is the real Runner, backed by os/exec.
type OSRunner struct{}

func (OSRunner) LookPath(binary string) (string, bool) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", false
	}
	return path, true
}

func (OSRunner) Output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}
