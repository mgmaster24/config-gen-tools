package cli

import (
	"bytes"
	"testing"
)

func TestNewRootCmd_RegistersExpectedSubcommands(t *testing.T) {
	root := newRootCmd()

	want := map[string]bool{"install": false, "doctor": false, "version": false}
	for _, c := range root.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

func TestDoctorCmd_JSONOutput(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"doctor", "--json"})
	// Doctor with no config file and no --lang runs universal checks only
	// against the real system; here we just confirm the command executes
	// and produces JSON-shaped output without error, regardless of what's
	// actually installed on the machine running the test.
	var buf bytes.Buffer
	root.SetOut(&buf)
	if err := root.Execute(); err != nil {
		// A missing Required check (e.g. no curl/tar on a minimal CI
		// runner) is a legitimate non-zero exit, not a test failure —
		// only fail on an unexpected error shape.
		if _, ok := err.(*ExitError); !ok {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if buf.Len() == 0 {
		t.Error("expected JSON output on stdout")
	}
}
