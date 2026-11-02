#!/usr/bin/env bash
# Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Installs the Asgardeo CLI (asg) on macOS and Linux.
#
#	curl -fsSL <install url> | bash                    # newest stable release
#	curl -fsSL <install url> | bash -s asg-v0.1.0      # a pinned release
#
# Releases live in wso2/wso2-cli under asg-v<semver> tags, beside the WSO2 CLI's
# own. The repository's Latest marker belongs to the WSO2 CLI, so the newest asg
# release is found by listing releases and comparing versions, never by asking
# GitHub for Latest and never by date: a fix published to an older line would
# otherwise win.
#
# This script is meant to be read before it is run, so it stays flat: one
# function per step, nothing that needs elevated privileges. It downloads the
# binary for this platform, verifies it against the checksums.txt published
# beside it, and refuses to install anything that fails that check.
#
# Running it again is safe. If the installed asg already reports the version it
# would install, it says so and stops; otherwise it replaces it, which is how an
# existing install moves to a new release.
#
# Variables it reads:
#
#	ASG_VERSION            A release tag to install, as the argument does.
#	ASG_PRERELEASE=true    Resolve the newest release including candidates,
#	                       not just the newest stable one.
#	ASG_INSTALL_DIR        Where to put the binary. Default ~/.local/bin. When
#	                       set, no shell profile is edited.
#	ASG_NO_PROFILE=1       Install without editing any shell profile.
#	ASG_RELEASE_BASE_URL   Where releases are downloaded from, and where they
#	ASG_RELEASE_API_URL    are listed. Overridden by tests; users have no
#	                       reason to set either.

set -euo pipefail

REPO="wso2/wso2-cli"
TAG_PREFIX="asg-v"
RELEASE_BASE_URL="${ASG_RELEASE_BASE_URL:-https://github.com/${REPO}/releases/download}"
RELEASE_API_URL="${ASG_RELEASE_API_URL:-https://api.github.com/repos/${REPO}/releases?per_page=100}"

fail() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "$1 is required but was not found."
}

# detect_asset prints the release asset name for this machine.
detect_asset() {
	local os arch
	case "$(uname -s)" in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	*) fail "no asg release for $(uname -s). Download a binary from https://github.com/${REPO}/releases" ;;
	esac

	case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	armv6l | armv7l | armv8l | arm) arch=arm ;;
	i386 | i686) arch=386 ;;
	*) fail "no asg release for ${os}/$(uname -m)." ;;
	esac

	# A shell running under Rosetta reports x86_64 on Apple silicon. The native
	# binary is the one to install there.
	if [ "$os" = darwin ] && [ "$arch" = amd64 ] &&
		[ "$(sysctl -n hw.optional.arm64 2>/dev/null || echo 0)" = 1 ]; then
		arch=arm64
	fi

	if [ "$os" = darwin ] && [ "$arch" != amd64 ] && [ "$arch" != arm64 ]; then
		fail "no asg release for darwin/${arch}."
	fi
	printf 'asg-%s-%s\n' "$os" "$arch"
}

# resolve_tag prints the release tag to install.
#
# An explicit tag wins. Otherwise releases are listed and filtered to asg-v*
# tags that are published, not drafts, and, unless ASG_PRERELEASE is set, not
# marked pre-release. What decides a candidate is the Pre-release flag set when
# publishing, not a suffix in the tag name. The survivors are ordered by
# semantic version, so 0.10.0 sorts above 0.9.0 and 0.1.0 above 0.1.0-rc.1.
resolve_tag() {
	local requested="$1" listing versions
	if [ -n "$requested" ]; then
		case "$requested" in
		"${TAG_PREFIX}"[0-9]*) printf '%s\n' "$requested" ;;
		*) fail "a release tag looks like ${TAG_PREFIX}0.1.0 (got ${requested})." ;;
		esac
		return
	fi

	listing="$(curl -fsSL -H 'Accept: application/vnd.github+json' "$RELEASE_API_URL")" ||
		fail "could not list releases from GitHub (network, or the unauthenticated API rate limit). Pass a tag to skip this step."

	# Each release carries tag_name, then draft, then prerelease. Nothing nested
	# inside a release uses those three names, so reading them in order yields
	# one record per release without a JSON parser.
	versions="$(printf '%s' "$listing" |
		grep -oE '"(tag_name|draft|prerelease)": *("[^"]*"|true|false)' |
		awk -v want_pre="${ASG_PRERELEASE:-false}" -F': *' '
			/"tag_name"/   { t = $2; gsub(/"/, "", t); d = ""; next }
			/"draft"/      { d = $2; next }
			/"prerelease"/ {
				if (t != "" && d == "false" && ($2 == "false" || want_pre == "true")) print t
				t = ""
			}' |
		grep -E "^${TAG_PREFIX}[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$" |
		sed "s/^${TAG_PREFIX}//")" || true

	[ -n "$versions" ] || fail "no published ${TAG_PREFIX}* release found in ${REPO}. Pass a tag, or set ASG_PRERELEASE=true for a release candidate."

	printf '%s%s\n' "$TAG_PREFIX" "$(printf '%s\n' "$versions" | newest_version)"
}

# newest_version reads semantic versions on stdin and prints the highest, by
# the precedence rules of semantic versioning rather than by text: 0.10.0 above
# 0.9.0, a release above its own candidates, and within a candidate suffix each
# dot-separated part compared on its own, numbers as numbers (rc.10 above rc.2)
# and below any word, and a longer suffix above its own prefix (rc.1 above rc).
#
# Each version is turned into a key that byte order sorts correctly: numbers
# zero-padded, then 1 for a release or 0 for a candidate, then each suffix part
# as 0 and a padded number or 1 and the word, joined by a comma, which sorts
# below every character a part may contain.
newest_version() {
	awk '{
		core = $0; pre = ""
		i = index($0, "-")
		if (i > 0) { core = substr($0, 1, i - 1); pre = substr($0, i + 1) }
		split(core, n, ".")
		key = sprintf("%010d.%010d.%010d.", n[1], n[2], n[3])
		if (pre == "") {
			key = key "1"
		} else {
			key = key "0"
			count = split(pre, part, ".")
			for (j = 1; j <= count; j++) {
				if (j > 1) key = key ","
				if (part[j] ~ /^[0-9]+$/) key = key "0" sprintf("%010d", part[j])
				else key = key "1" part[j]
			}
		}
		print key " " $0
	}' | LC_ALL=C sort | tail -n 1 | awk '{print $2}'
}

# installed_version prints the version an existing binary reports, or nothing.
# A binary too old to have --version reports nothing and is replaced.
installed_version() {
	[ -x "$1" ] || return 0
	"$1" --version 2>/dev/null | awk '/^asg version /{ print $3; exit }' || true
}

sha256_of() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{ print $1 }'
	else
		shasum -a 256 "$1" | awk '{ print $1 }'
	fi
}

# download_verified fetches the asset and checksums.txt into a directory and
# refuses unless the asset is listed and its digest matches.
download_verified() {
	local tag="$1" asset="$2" into="$3" expected actual
	curl -fsSL -o "${into}/${asset}" "${RELEASE_BASE_URL}/${tag}/${asset}" ||
		fail "could not download ${asset} from release ${tag}."
	curl -fsSL -o "${into}/checksums.txt" "${RELEASE_BASE_URL}/${tag}/checksums.txt" ||
		fail "could not download checksums.txt from release ${tag}."

	expected="$(awk -v f="$asset" '$2 == f || $2 == "*" f { print $1; exit }' "${into}/checksums.txt")"
	[ -n "$expected" ] || fail "checksums.txt for ${tag} does not list ${asset}, so the download cannot be verified."
	actual="$(sha256_of "${into}/${asset}")"
	if [ "$expected" != "$actual" ]; then
		printf 'error: checksum mismatch for %s\n  expected %s\n  actual   %s\n' "$asset" "$expected" "$actual" >&2
		fail 'refusing to install a binary that failed verification.'
	fi
}

# place_binary moves the verified binary into place by rename. Writing over a
# binary that has already run invalidates its ad-hoc signature on macOS and the
# next run exits with "Killed: 9"; a rename gives the path a new file instead,
# and a failed copy never leaves a half-written binary behind.
place_binary() {
	local from="$1" dir="$2" staged
	mkdir -p "$dir"
	staged="${dir}/.asg.$$.tmp"
	cp "$from" "$staged"
	chmod 0755 "$staged"
	mv -f "$staged" "${dir}/asg"
}

# wire_path makes sure the install directory is on PATH, editing one shell
# profile at most, and prints what it did.
wire_path() {
	local dir="$1" profile line
	case ":${PATH}:" in
	*":${dir}:"*)
		printf 'already on PATH'
		return
		;;
	esac

	if [ -n "${ASG_INSTALL_DIR:-}" ] || [ "${ASG_NO_PROFILE:-}" = 1 ]; then
		printf 'not on PATH; add %s to it' "$dir"
		return
	fi

	case "$(basename "${SHELL:-}")" in
	zsh)
		profile="${HOME}/.zshrc"
		line="export PATH=\"${dir}:\$PATH\""
		;;
	fish)
		profile="${HOME}/.config/fish/config.fish"
		line="set -gx PATH \"${dir}\" \$PATH"
		;;
	*)
		profile="${HOME}/.bashrc"
		line="export PATH=\"${dir}:\$PATH\""
		;;
	esac

	if ! grep -qF "$dir" "$profile" 2>/dev/null; then
		mkdir -p "$(dirname "$profile")"
		printf '\n# Added by the asg installer.\n%s\n' "$line" >>"$profile"
	fi
	printf 'added %s to PATH in %s; open a new terminal' "$dir" "$profile"
}

main() {
	need curl
	need uname
	command -v sha256sum >/dev/null 2>&1 || need shasum

	# work and dir are global on purpose: the EXIT trap that cleans them up runs
	# after main has returned, when its locals no longer exist.
	local asset tag target have path_note found
	asset="$(detect_asset)"
	tag="$(resolve_tag "${1:-${ASG_VERSION:-}}")"
	dir="${ASG_INSTALL_DIR:-${HOME}/.local/bin}"
	target="${dir}/asg"

	have="$(installed_version "$target")"
	if [ -n "$have" ] && [ "$have" = "${tag#"$TAG_PREFIX"}" ]; then
		printf '%s is already installed at %s\n' "$tag" "$target"
		return
	fi

	work="$(mktemp -d)"
	trap 'rm -rf "$work"; rm -f "${dir}/.asg.$$.tmp"' EXIT

	printf 'Installing %s for %s.\n' "$tag" "${asset#asg-}"
	download_verified "$tag" "$asset" "$work"
	place_binary "${work}/${asset}" "$dir"
	path_note="$(wire_path "$dir")"

	printf '%s installed at %s (checksum verified; %s)\n' "$tag" "$target" "$path_note"

	found="$(command -v asg 2>/dev/null || true)"
	if [ -n "$found" ] && [ "$found" != "$target" ]; then
		printf 'note: another asg at %s comes first on PATH and will run instead.\n' "$found"
	fi
}

# Everything runs from here, so a download cut off half way through a
# `curl | bash` defines functions and runs nothing.
main "$@"
