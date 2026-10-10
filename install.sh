#!/bin/sh
#   curl -fsSL https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.sh | sh
#
# TERMINALSCORE_VERSION=0.1.0       specific version (default: latest)
# TERMINALSCORE_INSTALL_DIR=...     default: a writable directory already in PATH
#                                   ($PREFIX/bin on Termux, /usr/local/bin, ~/.local/bin),
#                                   otherwise /usr/local/bin with sudo

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

in_path() {
	case ":$PATH:" in *":$1:"*) return 0 ;; esac
	return 1
}

is_termux() {
	[ -n "${TERMUX_VERSION:-}" ] && return 0
	case "${PREFIX:-}" in *com.termux*) return 0 ;; esac
	return 1
}

# Kurulumdan hemen sonra çalışsın diye PATH'te olan, yazılabilir bir klasör
# seçer; yoksa sudo ile /usr/local/bin. install_dir ve sudo'yu ayarlar.
pick_dir() {
	sudo=""
	if [ -n "${TERMINALSCORE_INSTALL_DIR:-}" ]; then
		install_dir=$TERMINALSCORE_INSTALL_DIR
		return
	fi
	if is_termux; then
		install_dir="$PREFIX/bin"
		return
	fi
	for d in /usr/local/bin "$HOME/.local/bin" "$HOME/bin"; do
		in_path "$d" || continue
		if [ -d "$d" ] && [ -w "$d" ]; then
			install_dir=$d
			return
		fi
		# PATH'te ama henüz yok (Ubuntu ~/.local/bin'i böyle ekliyor)
		if [ ! -e "$d" ] && [ "${d#"$HOME"/}" != "$d" ]; then
			install_dir=$d
			return
		fi
	done
	if [ "$(id -u)" != 0 ] && command -v sudo >/dev/null 2>&1 && in_path /usr/local/bin; then
		install_dir=/usr/local/bin
		sudo=sudo
		return
	fi
	install_dir="$HOME/.local/bin"
}

# Son çare: klasörü kabuk ayar dosyasına ekle.
add_to_path() {
	shell=$(basename "${SHELL:-sh}")
	case "$shell" in
	bash) rc="$HOME/.bashrc" ;;
	zsh) rc="${ZDOTDIR:-$HOME}/.zshrc" ;;
	fish) rc="$HOME/.config/fish/config.fish" ;;
	*) rc="$HOME/.profile" ;;
	esac
	if [ "$shell" = fish ]; then
		line="fish_add_path $1"
	else
		line="export PATH=\"\$PATH:$1\""
	fi
	mkdir -p "$(dirname "$rc")"
	grep -qsF "$line" "$rc" || printf '\n# terminalscore\n%s\n' "$line" >>"$rc"
	printf 'Added %s to PATH in %s.\nOpen a new terminal, or run this once in the current one:\n  %s\n' "$1" "$rc" "$line"
}

main() {
	command -v curl >/dev/null 2>&1 || fail "curl is required"
	command -v tar >/dev/null 2>&1 || fail "tar is required"

	os=$(detect_os)
	arch=$(detect_arch)
	version=${TERMINALSCORE_VERSION:-$(latest_version)}
	version=${version#v}
	pick_dir
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
	[ -n "$sudo" ] && printf 'Installing to %s (needs sudo)...\n' "$install_dir"
	$sudo mkdir -p "$install_dir"
	$sudo install -m 755 "$tmp/terminalscore" "$install_dir/terminalscore" 2>/dev/null ||
		{ $sudo cp "$tmp/terminalscore" "$install_dir/terminalscore" && $sudo chmod 755 "$install_dir/terminalscore"; } ||
		fail "couldn't write to $install_dir"

	printf 'Installed: %s/terminalscore\n' "$install_dir"

	if ! in_path "$install_dir"; then
		add_to_path "$install_dir"
		return
	fi

	found=$(command -v terminalscore 2>/dev/null || true)
	if [ -n "$found" ] && [ "$found" != "$install_dir/terminalscore" ]; then
		printf 'Note: an older copy at %s comes first in PATH, remove it to use this one.\n' "$found"
	fi
	printf 'Run it with: terminalscore\n'
}

main "$@"
