#!/usr/bin/env bash
#
# End-to-end smoke test run by .github/workflows/ci.yml. Writes nothing:
# --dry-run renders to stdout only.
set -euo pipefail

go run ./cmd/gitforge --help >/dev/null
go run ./cmd/gitforge version

# generate needs identities, so drive it from a throwaway config.
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cat > "$work/gitforge.toml" <<'TOML'
deploy_path = "/tmp/gitforge-ci"
features = ["rerere", "default-branch-main"]

[[identities]]
name = "default"
user_name = "CI"
email = "ci@example.com"

[[identities]]
name = "work"
user_name = "CI Work"
email = "ci@work.example"
dir = "~/work"
TOML

go run ./cmd/gitforge generate --config "$work/gitforge.toml" --dry-run
