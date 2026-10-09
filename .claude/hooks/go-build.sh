#!/bin/sh
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

# A Claude Code PostToolUse hook: after an edit, compile the Go module the edited
# file belongs to, so a broken build is reported on the edit that caused it
# rather than at push or in CI. Registered in .claude/settings.json.
#
# Edit and Write name the file they touched. A Bash command can touch any file,
# so for it the Go files modified since the previous run are found by
# timestamp. Each affected module is built with `go build ./...`; the SDK with
# GOWORK=off, as CI builds it. go build compiles neither test files nor the
# smoke-tagged live runs, so the package of such a file is vetted as well, which
# type-checks it.
#
# Silent when everything compiles. On failure the compiler output goes to
# stderr and the exit status is 2, which Claude Code hands back to the agent.

set -u

input=$(cat)

field() {
	if command -v jq >/dev/null 2>&1; then
		printf '%s' "$input" | jq -r "$1 // empty"
	else
		printf '%s' "$input" | python3 -c '
import json, sys
value = json.load(sys.stdin)
for key in sys.argv[1].lstrip(".").split("."):
    value = value.get(key) if isinstance(value, dict) else None
print(value if isinstance(value, str) else "")
' "$1"
	fi
}

project=${CLAUDE_PROJECT_DIR:-$(field .cwd)}
root=$(git -C "$project" rev-parse --show-toplevel 2>/dev/null) || exit 0
command -v go >/dev/null 2>&1 || exit 0
stamp="$(git -C "$root" rev-parse --absolute-git-dir)/claude-go-build.stamp"

case $(field .tool_name) in
Bash)
	if [ ! -f "$stamp" ]; then
		touch "$stamp"
		exit 0
	fi
	changed=$(find "$root" \( -name .git -o -name bin -o -name dist -o -name node_modules \
		-o -path "$root/.claude/worktrees" \) -prune \
		-o -name '*.go' -newer "$stamp" -print)
	;;
*)
	changed=$(field .tool_input.file_path)
	;;
esac
touch "$stamp"

# One line per module directory, and one tab-separated module and package
# directory per package to vet. Paths are read a line at a time so that spaces
# survive.
tab=$(printf '\t')
modules=
tests=
while IFS= read -r file; do
	case $file in
	"$root"/*.go) ;;
	*) continue ;;
	esac
	dir=$(dirname "$file")
	module=$dir
	while [ ! -f "$module/go.mod" ] && [ "$module" != "$root" ]; do
		module=$(dirname "$module")
	done
	[ -f "$module/go.mod" ] || continue
	modules=$(printf '%s\n%s' "$modules" "$module")
	case $file in
	*_test.go | "$root"/test/smoke/*) tests=$(printf '%s\n%s\t%s' "$tests" "$module" "$dir") ;;
	esac
done <<EOF
$changed
EOF

[ -n "$modules" ] || exit 0

failed=0

# go_in <module> <go arguments>... runs go inside module, outside the
# workspace for the SDK, which CI builds on its own.
go_in() {
	module=$1
	shift
	if [ "$module" = "$root/sdk" ]; then
		(cd "$module" && GOWORK=off go "$@")
	else
		(cd "$module" && go "$@")
	fi
}

# label <dir> names a directory relative to the repository root.
label() {
	if [ "$1" = "$root" ]; then
		echo "the repository root"
	else
		echo "${1#"$root"/}"
	fi
}

while IFS= read -r module; do
	[ -n "$module" ] || continue
	if ! out=$(go_in "$module" build -o /dev/null ./... 2>&1 </dev/null); then
		echo "go build failed in $(label "$module"):" >&2
		echo "$out" >&2
		failed=1
	fi
done <<EOF
$(printf '%s\n' "$modules" | sort -u)
EOF

while IFS="$tab" read -r module dir; do
	[ -n "$module" ] || continue
	tags=
	case $dir in
	"$root"/test/smoke | "$root"/test/smoke/*) tags=-tags=smoke ;;
	esac
	if ! out=$(go_in "$module" vet $tags ".${dir#"$module"}" 2>&1 </dev/null); then
		echo "go vet failed in $(label "$dir"):" >&2
		echo "$out" >&2
		failed=1
	fi
done <<EOF
$(printf '%s\n' "$tests" | sort -u)
EOF

[ "$failed" -eq 0 ] || exit 2
