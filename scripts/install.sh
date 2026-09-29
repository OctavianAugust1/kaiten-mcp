#!/bin/sh

# Install a verified pre-built kaiten-mcp binary from GitHub Releases.
set -eu

repository="OctavianAugust1/kaiten-mcp"
version="${KAITEN_MCP_VERSION:-}"
install_dir="${KAITEN_MCP_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
	printf '%s\n' "install: $*" >&2
	exit 1
}

if [ -z "$version" ]; then
	version="$(curl -fsSL "https://api.github.com/repos/${repository}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
	[ -n "$version" ] || fail "could not determine the latest release; set KAITEN_MCP_VERSION"
fi

case "$(uname -s)" in
	Linux) operating_system="linux" ;;
	Darwin) operating_system="darwin" ;;
	*) fail "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
	x86_64|amd64) architecture="amd64" ;;
	aarch64|arm64) architecture="arm64" ;;
	*) fail "unsupported architecture: $(uname -m)" ;;
esac

release_version="${version#v}"
archive="kaiten-mcp_${release_version}_${operating_system}_${architecture}.tar.gz"
release_url="https://github.com/${repository}/releases/download/${version}"
temporary_dir="$(mktemp -d)"
trap 'rm -rf "$temporary_dir"' 0 HUP INT TERM

curl -fsSL "${release_url}/${archive}" -o "${temporary_dir}/${archive}"
curl -fsSL "${release_url}/kaiten-mcp_checksums.txt" -o "${temporary_dir}/checksums.txt"
expected_checksum="$(awk -v file="$archive" '$2 == file { print $1 }' "${temporary_dir}/checksums.txt")"
[ -n "$expected_checksum" ] || fail "checksum for ${archive} is missing"

if command -v sha256sum >/dev/null 2>&1; then
	actual_checksum="$(sha256sum "${temporary_dir}/${archive}" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
	actual_checksum="$(shasum -a 256 "${temporary_dir}/${archive}" | awk '{ print $1 }')"
else
	fail "sha256sum or shasum is required to verify the download"
fi
[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum verification failed"

tar -xzf "${temporary_dir}/${archive}" -C "$temporary_dir"
[ -f "${temporary_dir}/kaiten-mcp" ] || fail "archive does not contain kaiten-mcp"
mkdir -p "$install_dir"

previous_version=""
if [ -r "${install_dir}/.kaiten-mcp-version" ]; then
	previous_version="$(cat "${install_dir}/.kaiten-mcp-version")"
fi

install -m 0755 "${temporary_dir}/kaiten-mcp" "${install_dir}/kaiten-mcp-bin"

# Codex starts the configured command directly, so this launcher loads the
# user's exported Kaiten settings before it delegates to the downloaded binary.
printf '%s\n' \
	'#!/usr/bin/env bash' \
	'set -euo pipefail' \
	'script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"' \
	'if [[ -r "${HOME}/.bashrc" ]]; then' \
	'  source "${HOME}/.bashrc" >/dev/null 2>&1 || true' \
	'fi' \
	'exec "${script_dir}/kaiten-mcp-bin" "$@"' \
	> "${install_dir}/kaiten-mcp"
chmod 0755 "${install_dir}/kaiten-mcp"
printf '%s\n' "$version" > "${install_dir}/.kaiten-mcp-version"

if [ -n "$previous_version" ] && [ "$previous_version" != "$version" ]; then
	printf '%s\n' "Updated kaiten-mcp from ${previous_version} to ${version} in ${install_dir}/kaiten-mcp"
elif [ -n "$previous_version" ]; then
	printf '%s\n' "Reinstalled current kaiten-mcp ${version} in ${install_dir}/kaiten-mcp"
else
	printf '%s\n' "Installed kaiten-mcp ${version} to ${install_dir}/kaiten-mcp"
fi

register_codex() {
	if ! command -v codex >/dev/null 2>&1; then
		printf '%s\n' "Codex CLI was not found; skipped Codex MCP registration"
		return
	fi
	if codex mcp get kaiten >/dev/null 2>&1; then
		printf '%s\n' "Codex MCP server 'kaiten' is already configured"
		return
	fi
	if codex mcp add kaiten -- "${install_dir}/kaiten-mcp"; then
		printf '%s\n' "Added global MCP server 'kaiten' to Codex"
	else
		printf '%s\n' "install: could not register kaiten in Codex; add it manually" >&2
	fi
}

register_claude() {
	if ! command -v claude >/dev/null 2>&1; then
		printf '%s\n' "Claude Code CLI was not found; skipped Claude Code MCP registration"
		return
	fi
	if claude mcp get kaiten >/dev/null 2>&1; then
		printf '%s\n' "Claude Code MCP server 'kaiten' is already configured"
		return
	fi
	if claude mcp add kaiten --scope user -- "${install_dir}/kaiten-mcp"; then
		printf '%s\n' "Added user-scope MCP server 'kaiten' to Claude Code"
	else
		printf '%s\n' "install: could not register kaiten in Claude Code; add it manually" >&2
	fi
}

if [ "${KAITEN_MCP_REGISTER_CLIENTS:-1}" != "0" ]; then
	register_codex
	register_claude
fi
