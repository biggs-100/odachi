#!/usr/bin/env bash
# odachi installer · un solo binario, sin daemon, sin telemetría.
#
# Baja el binario de una release, verifica su sha256 y recién ahí lo instala.
# Si el checksum no coincide, no se instala nada: un instalador que deja un
# binario a medio verificar es peor que no tener instalador.
#
# Variables de entorno (las tres son opcionales):
#   ODACHI_VERSION   versión a instalar (default: la última release publicada)
#   ODACHI_BASE_URL  base de descarga (default: la URL de releases de GitHub).
#                    También acepta file://, que es lo que permite probar este
#                    script sin red.
#   ODACHI_BIN_DIR   carpeta de destino (default: $HOME/.local/bin)
set -euo pipefail

REPO="biggz/odachi"

die() {
	printf 'odachi: %s\n' "$1" >&2
	exit 1
}

# 0 · herramientas: sin curl no hay descarga, sin tar no hay binario, y sin un
# verificador de sha256 no se instala nada (nunca "seguir sin verificar").
for tool in curl tar; do
	command -v "$tool" >/dev/null 2>&1 || die "me falta $tool para bajar y descomprimir la release; instalalo y volvé a correr esto"
done

sha256_of() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d' ' -f1
	else
		return 1
	fi
}

if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
	die "no tengo sha256sum ni shasum para verificar la descarga, y no voy a instalar algo sin verificar"
fi

# 1 · plataforma. Lo que no está acá se dice, no se adivina.
case "$(uname -s 2>/dev/null || echo desconocido)" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*)
		die "no sé instalar en $(uname -s 2>/dev/null || echo esta plataforma); este instalador cubre linux y darwin (en Windows el equivalente es scripts/install.ps1)"
		;;
esac

case "$(uname -m 2>/dev/null || echo desconocido)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*)
		die "no sé instalar en la arquitectura $(uname -m 2>/dev/null || echo desconocida); las releases cubren amd64 y arm64"
		;;
esac

# 2 · versión: la que pidió el humano, o la última release.
version=${ODACHI_VERSION:-}
if [ -z "$version" ]; then
	# La respuesta se guarda primero y se parsea después: así el pipeline no le
	# corta la descarga a curl cuando el parseo ya encontró lo que buscaba.
	latest=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest") ||
		die "no pude consultar la última release en GitHub; pasá ODACHI_VERSION=<versión> y volvé a correr esto"
	version=$(printf '%s\n' "$latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
	[ -n "$version" ] || die "no pude averiguar la última versión en GitHub; pasá ODACHI_VERSION=<versión> y volvé a correr esto"
fi
# El tag lleva "v" (v1.2.3); el archivo se llama odachi_1.2.3_...
version=${version#v}
[ -n "$version" ] || die "la versión llegó vacía; pasá ODACHI_VERSION=<versión>"

asset="odachi_${version}_${os}_${arch}.tar.gz"

# 3 · base de descarga configurable: es la costura que permite probar esto sin red.
base=${ODACHI_BASE_URL:-https://github.com/$REPO/releases/download/v$version}
base=${base%/}

tmp=$(mktemp -d "${TMPDIR:-/tmp}/odachi-install.XXXXXX")
trap 'rm -rf -- "$tmp"' EXIT

# 4 · descarga y verificación, en un directorio temporal: el destino no se toca
# hasta que el archivo demuestra ser el que la release publicó.
curl -fsSL -o "$tmp/$asset" "$base/$asset" || die "no pude descargar $base/$asset; revisá la versión y tu conexión"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" || die "no pude descargar $base/checksums.txt; sin checksums no instalo nada"

expected=$(awk -v name="$asset" '($2 == name || $2 == "*" name) { print $1; exit }' "$tmp/checksums.txt")
[ -n "$expected" ] || die "checksums.txt no trae el hash de $asset; no instalo algo que la release no certifica"

actual=$(sha256_of "$tmp/$asset") || die "no pude calcular el sha256 de $asset"
if [ "$expected" != "$actual" ]; then
	die "el checksum de $asset no coincide (esperaba $expected, obtuve $actual): el archivo está corrupto o no es el que dice ser, y no instalé nada"
fi

mkdir -p "$tmp/out"
tar -xzf "$tmp/$asset" -C "$tmp/out" || die "no pude descomprimir $asset"
[ -f "$tmp/out/odachi" ] || die "$asset no trae el binario odachi"

# 5 · instalación: $ODACHI_BIN_DIR, o $HOME/.local/bin, o al lado de donde se
# corrió el script si no hay permiso de escritura.
if [ -z "${ODACHI_BIN_DIR:-}" ] && [ -z "${HOME:-}" ]; then
	die "no sé dónde instalar: no hay HOME en este entorno; pasá ODACHI_BIN_DIR=<carpeta>"
fi
bin_dir=${ODACHI_BIN_DIR:-$HOME/.local/bin}

# Copy next to the destination and move it into place: an interrupted copy must
# never leave a truncated binary where a working one used to be. Same rule the
# tool applies to every file it writes.
install_into() {
	target_dir=$1
	mkdir -p "$target_dir" 2>/dev/null || return 1
	staged="$target_dir/.odachi.$$"
	cp "$tmp/out/odachi" "$staged" 2>/dev/null || { rm -f "$staged"; return 1; }
	chmod +x "$staged" 2>/dev/null || { rm -f "$staged"; return 1; }
	mv -f "$staged" "$target_dir/odachi" 2>/dev/null || { rm -f "$staged"; return 1; }
	dest="$target_dir/odachi"
	return 0
}

if ! install_into "$bin_dir"; then
	dest="./odachi"
	cp "$tmp/out/odachi" "$dest" || die "no pude escribir en $bin_dir ni en la carpeta actual"
	printf 'odachi: no pude escribir en %s, así que lo dejé en %s\n' "$bin_dir" "$dest" >&2
fi
chmod +x "$dest"

# 6 · qué quedó instalado, y cómo usarlo desde cualquier carpeta.
printf 'odachi %s instalado en %s\n' "$version" "$dest"

if [ "$dest" = "$bin_dir/odachi" ]; then
	case ":${PATH:-}:" in
		*":$bin_dir:"*) ;;
		*)
			printf 'odachi: %s no está en tu PATH; pegá esto en tu ~/.bashrc (o ~/.zshrc) y abrí una terminal nueva:\n  export PATH="%s:$PATH"\n' "$bin_dir" "$bin_dir" >&2
			;;
	esac
fi
