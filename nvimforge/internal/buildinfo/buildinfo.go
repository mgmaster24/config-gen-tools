// Package buildinfo holds version metadata injected at build time via
// -ldflags in .goreleaser.yaml. The defaults below apply to local
// `go build`/`go run`, where no ldflags are set.
package buildinfo

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
