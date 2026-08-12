#!/usr/bin/env bash
#
# End-to-end smoke test run by .github/workflows/ci.yml.
#
# Hits the real GitHub API for neovim/neovim's latest release metadata (to
# report what --dry-run would do) but performs zero downloads or writes — the
# one intentional exception to "no network" in the rest of the test suite.
set -euo pipefail

deploy="${RUNNER_TEMP:-/tmp}/nvimforge-ci-deploy"

go run ./cmd/nvimforge install \
  --dry-run \
  --yes \
  --lang go \
  --lang lua \
  --lang csharp \
  --deploy-path "$deploy"
