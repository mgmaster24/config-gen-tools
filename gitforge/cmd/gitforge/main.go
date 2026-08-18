// Command gitforge generates an includable gitconfig with directory-scoped
// identities.
package main

import (
	"os"

	"github.com/mgmaster24/config-gen-tools/gitforge/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
