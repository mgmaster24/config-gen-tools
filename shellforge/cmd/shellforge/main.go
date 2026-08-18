// Command shellforge generates an ordered shell init script from a set of
// selected tool integrations.
package main

import (
	"os"

	"github.com/mgmaster24/config-gen-tools/shellforge/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
