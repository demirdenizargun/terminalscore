#!/bin/sh
#   curl -fsSL https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.sh | sh
#
# TERMINALSCORE_VERSION=0.1.0       specific version (default: latest)
# TERMINALSCORE_INSTALL_DIR=...     default: ~/.local/bin, /usr/local/bin for root

set -eu

REPO="demirdenizargun/terminalscore"

fail() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

detect_os() {
	case "$(uname -s)" in
	Linux) echo linux ;;
	Darwin) echo darwin ;;
	*) fail "unsupported OS: $(uname -s) (use install.ps1 on Windows)" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
	x86_64 | amd64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*) fail "unsupported CPU: $(uname -m)" ;;
	esac
}

latest_version() {
	# releases/latest -> .../releases/tag/v0.1.0
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest") ||
		fail "couldn't find the latest release"
	tag=${url##*/}
	case "$tag" in
	v*) echo "${tag#v}" ;;
	*) fail "couldn't find the latest release" ;;
	esac
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d' ' -f1
	else
		fail "sha256sum or shasum is required"
	fi
}

default_dir() {
	if [ "$(id -u)" = 0 ]; then
		echo /usr/local/bin
	else
		echo "$HOME/.local/bin"
	fi
}

main() {
	command -v curl >/dev/null 2>&1 || fail "curl is required"
	command -v tar >/dev/null 2>&1 || fail "tar is required"

	os=$(detect_os)
	arch=$(detect_arch)
	version=${TERMINALSCORE_VERSION:-$(latest_version)}
	version=${version#v}
	install_dir=${TERMINALSCORE_INSTALL_DIR:-$(default_dir)}
	base=${TERMINALSCORE_BASE_URL:-https://github.com/$REPO/releases/download/v$version}

	archive="terminalscore_${version}_${os}_${arch}.tar.gz"
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	printf 'Downloading terminalscore %s (%s/%s)...\n' "$version" "$os" "$arch"
	curl -fsSL "$base/$archive" -o "$tmp/$archive" || fail "couldn't download $archive"
	curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || fail "couldn't download checksums.txt"

	expected=$(grep " $archive\$" "$tmp/checksums.txt" | cut -d' ' -f1)
	[ -n "$expected" ] || fail "no checksum for $archive"
	[ "$(sha256 "$tmp/$archive")" = "$expected" ] || fail "checksum mismatch, the download is corrupted"

	tar -xzf "$tmp/$archive" -C "$tmp" terminalscore
	mkdir -p "$install_dir"
	install -m 755 "$tmp/terminalscore" "$install_dir/terminalscore" 2>/dev/null ||
		{ cp "$tmp/terminalscore" "$install_dir/terminalscore" && chmod 755 "$install_dir/terminalscore"; }

	printf 'Installed: %s/terminalscore\n' "$install_dir"

	case ":$PATH:" in
	*":$install_dir:"*)
		printf 'Run it with: terminalscore\n'
		;;
	*)
		printf '\n%s is not in your PATH. Add this to your shell config (e.g. ~/.bashrc):\n' "$install_dir"
		printf '  export PATH="$PATH:%s"\n' "$install_dir"
		;;
	esac
}

main "$@"
