#!/usr/bin/env bash
#
# End-to-end smoke test run by .github/workflows/ci.yml. Writes nothing:
# --dry-run renders to stdout only.
set -euo pipefail

go run ./cmd/shellforge --help >/dev/null
go run ./cmd/shellforge version
