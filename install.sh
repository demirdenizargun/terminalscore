#!/bin/sh
# terminalscore kurulum betiği (Linux / macOS).
#
#   curl -fsSL https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.sh | sh
#
# Ayarlanabilenler (ortam değişkeni olarak):
#   TERMINALSCORE_VERSION      kurulacak sürüm, ör. 0.1.0 (varsayılan: en son sürüm)
#   TERMINALSCORE_INSTALL_DIR  kurulum klasörü (varsayılan: ~/.local/bin)
#
# Her şey main içinde: bağlantı yarıda koparsa yarım betik çalışmaz.

set -eu

REPO="demirdenizargun/terminalscore"

fail() {
	printf 'hata: %s\n' "$1" >&2
	exit 1
}

detect_os() {
	case "$(uname -s)" in
	Linux) echo linux ;;
	Darwin) echo darwin ;;
	*) fail "desteklenmeyen işletim sistemi: $(uname -s) (Windows için install.ps1 kullanın)" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
	x86_64 | amd64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*) fail "desteklenmeyen işlemci: $(uname -m)" ;;
	esac
}

latest_version() {
	# github.com/.../releases/latest, en son sürümün sayfasına yönlendirir: .../tag/v0.1.0
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest") ||
		fail "en son sürüm bulunamadı (henüz sürüm yayınlanmamış olabilir)"
	tag=${url##*/}
	case "$tag" in
	v*) echo "${tag#v}" ;;
	*) fail "en son sürüm bulunamadı (henüz sürüm yayınlanmamış olabilir)" ;;
	esac
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d' ' -f1
	else
		fail "sha256sum ya da shasum bulunamadı"
	fi
}

main() {
	command -v curl >/dev/null 2>&1 || fail "curl gerekli"
	command -v tar >/dev/null 2>&1 || fail "tar gerekli"

	os=$(detect_os)
	arch=$(detect_arch)
	version=${TERMINALSCORE_VERSION:-$(latest_version)}
	version=${version#v}
	install_dir=${TERMINALSCORE_INSTALL_DIR:-$HOME/.local/bin}
	base=${TERMINALSCORE_BASE_URL:-https://github.com/$REPO/releases/download/v$version}

	archive="terminalscore_${version}_${os}_${arch}.tar.gz"
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	printf 'terminalscore %s indiriliyor (%s/%s)...\n' "$version" "$os" "$arch"
	curl -fsSL "$base/$archive" -o "$tmp/$archive" || fail "$archive indirilemedi"
	curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || fail "checksums.txt indirilemedi"

	expected=$(grep " $archive\$" "$tmp/checksums.txt" | cut -d' ' -f1)
	[ -n "$expected" ] || fail "$archive için sağlama bilgisi yok"
	[ "$(sha256 "$tmp/$archive")" = "$expected" ] || fail "indirilen dosya bozuk (sağlama tutmadı)"

	tar -xzf "$tmp/$archive" -C "$tmp" terminalscore
	mkdir -p "$install_dir"
	install -m 755 "$tmp/terminalscore" "$install_dir/terminalscore" 2>/dev/null ||
		{ cp "$tmp/terminalscore" "$install_dir/terminalscore" && chmod 755 "$install_dir/terminalscore"; }

	printf 'Kuruldu: %s/terminalscore\n' "$install_dir"

	case ":$PATH:" in
	*":$install_dir:"*)
		printf 'Başlatmak için: terminalscore\n'
		;;
	*)
		printf '\n%s PATH içinde değil. Kabuk ayar dosyana (ör. ~/.bashrc) ekle:\n' "$install_dir"
		printf '  export PATH="$PATH:%s"\n' "$install_dir"
		;;
	esac
}

main "$@"
