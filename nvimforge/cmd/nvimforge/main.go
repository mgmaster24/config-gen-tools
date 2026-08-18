package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if exitErr, ok := errors.AsType[*cli.ExitError](err); ok {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
