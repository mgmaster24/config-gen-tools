#!/usr/bin/env bash
# Installs the latest (or GITFORGE_VERSION-pinned) gitforge release
# binary for this machine's OS/arch. This script only downloads, verifies,
# and places the binary — all of gitforge's actual logic lives in the Go
# binary itself. For Windows, use install.ps1 instead.
set -euo pipefail

REPO="mgmaster24/config-gen-tools"
# config-gen-tools holds several tools, each released under its own
# tool-scoped tag (e.g. gitforge/v1.2.3), so releases must be filtered by
# this prefix rather than taking the repo's newest release.
TOOL="gitforge"
INSTALL_DIR="${GITFORGE_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${GITFORGE_VERSION:-}"

info() { printf '\033[1;34m==>\033[0m %s\n' "$1"; }
err() {
  printf '\033[1;31merror:\033[0m %s\n' "$1" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *) err "unsupported OS: $(uname -s). On Windows, use install.ps1 instead." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ -z "$VERSION" ]; then
  info "Resolving latest ${TOOL} release..."
  # /releases/latest would return whichever tool released most recently, so
  # list releases (newest first) and take the first one tagged for this tool.
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=100" |
    grep '"tag_name":' |
    sed -E 's/.*"tag_name": *"([^"]+)".*/\1/' |
    grep "^${TOOL}/" |
    head -n1 |
    sed "s|^${TOOL}/||")"
  [ -n "$VERSION" ] || err "could not resolve the latest ${TOOL} release"
fi

# Accept either a bare version (v1.2.3) or a fully-qualified tag
# (gitforge/v1.2.3) in GITFORGE_VERSION.
VERSION="${VERSION#"${TOOL}/"}"
VERSION_NUM="${VERSION#v}"
ARCHIVE="gitforge_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${TOOL}/${VERSION}"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

info "Downloading ${ARCHIVE} (${VERSION})..."
curl -fsSL -o "${WORKDIR}/${ARCHIVE}" "${BASE_URL}/${ARCHIVE}"
curl -fsSL -o "${WORKDIR}/checksums.txt" "${BASE_URL}/checksums.txt"

info "Verifying checksum..."
(
  cd "$WORKDIR"
  expected="$(grep " ${ARCHIVE}\$" checksums.txt | awk '{print $1}')"
  [ -n "$expected" ] || err "no checksum entry found for ${ARCHIVE}"
  if command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${ARCHIVE}" | awk '{print $1}')"
  elif command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${ARCHIVE}" | awk '{print $1}')"
  else
    err "neither shasum nor sha256sum is available to verify the download"
  fi
  [ "$expected" = "$actual" ] || err "checksum mismatch for ${ARCHIVE} (expected ${expected}, got ${actual})"
)

info "Extracting..."
tar -xzf "${WORKDIR}/${ARCHIVE}" -C "$WORKDIR"

mkdir -p "$INSTALL_DIR"
mv "${WORKDIR}/gitforge" "${INSTALL_DIR}/gitforge"
chmod +x "${INSTALL_DIR}/gitforge"

info "Installed gitforge ${VERSION} to ${INSTALL_DIR}/gitforge"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    printf '\nNote: %s is not on your PATH. Add this to your shell profile:\n\n  export PATH="%s:$PATH"\n\n' \
      "$INSTALL_DIR" "$INSTALL_DIR"
    ;;
esac

info "Run 'gitforge generate' to get started."
