#!/usr/bin/env bash
set -euo pipefail

REPO="reloadlife/cursor-account-switcher"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY="cursor-switch"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$arch" in
x86_64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*)
	echo "unsupported arch: $arch" >&2
	exit 1
	;;
esac

case "$os" in
darwin | linux) ;;
*)
	echo "unsupported os: $os" >&2
	exit 1
	;;
esac

asset="${BINARY}-${os}-${arch}"
url=$(
	curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
		grep "browser_download_url.*/${asset}\"" |
		cut -d '"' -f 4 |
		head -n 1
)

if [ -z "$url" ]; then
	echo "release asset not found: ${asset}" >&2
	exit 1
fi

mkdir -p "$INSTALL_DIR"
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"
mv "$tmp" "$INSTALL_DIR/$BINARY"
echo "installed ${INSTALL_DIR}/${BINARY}"
