#!/usr/bin/env bash
# Installs the latest (or NVIMFORGE_VERSION-pinned) nvimforge release
# binary for this machine's OS/arch. This script only downloads, verifies,
# and places the binary — all of nvimforge's actual logic lives in the Go
# binary itself. For Windows, use install.ps1 instead.
set -euo pipefail

REPO="mgmaster24/nvimforge"
INSTALL_DIR="${NVIMFORGE_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${NVIMFORGE_VERSION:-}"

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
  info "Resolving latest nvimforge release..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
    grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  [ -n "$VERSION" ] || err "could not resolve the latest release version"
fi

VERSION_NUM="${VERSION#v}"
ARCHIVE="nvimforge_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

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
mv "${WORKDIR}/nvimforge" "${INSTALL_DIR}/nvimforge"
chmod +x "${INSTALL_DIR}/nvimforge"

info "Installed nvimforge ${VERSION} to ${INSTALL_DIR}/nvimforge"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    printf '\nNote: %s is not on your PATH. Add this to your shell profile:\n\n  export PATH="%s:$PATH"\n\n' \
      "$INSTALL_DIR" "$INSTALL_DIR"
    ;;
esac

info "Run 'nvimforge install' to get started."
