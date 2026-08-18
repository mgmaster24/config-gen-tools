module github.com/mgmaster24/config-gen-tools/gitforge

go 1.26.5

require (
	github.com/mgmaster24/config-gen-tools/forge v0.0.0
	github.com/pelletier/go-toml/v2 v2.4.3
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

// See nvimforge/go.mod: replace until forge/v0.1.0 is tagged.
replace github.com/mgmaster24/config-gen-tools/forge => ../forge
