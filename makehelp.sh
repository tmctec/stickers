#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# makehelp.sh — Complex logic extracted from Makefile
# ============================================================================
# The Makefile defines WHAT; this script implements HOW for anything too
# complex for inline make recipes (>10 lines, OS branching, multi-step ops).
# ============================================================================

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------

check_command() {
    command -v "$1" &>/dev/null
}

# ---------------------------------------------------------------------------
# Command functions
# ---------------------------------------------------------------------------

cmd_prereqs() {
    echo "Checking prerequisites..."

    if ! check_command go; then
        echo "Error: Go 1.26 or later is required but not installed." >&2
        case "$(uname -s)" in
            Darwin) echo "Install it with: brew install go" >&2 ;;
            *)      echo "Install it from: https://go.dev/dl/" >&2 ;;
        esac
        exit 1
    fi
    echo "  go: $(go env GOVERSION)"

    if ! check_command golangci-lint; then
        echo "  golangci-lint: missing, installing it with go install..."
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
        if ! check_command golangci-lint; then
            echo "Error: golangci-lint is installed in $(go env GOPATH)/bin, which is not on PATH." >&2
            echo "Add it with: export PATH=\"\$PATH:$(go env GOPATH)/bin\"" >&2
            exit 1
        fi
    fi
    echo "  golangci-lint: $(golangci-lint version --short 2>/dev/null || golangci-lint version | head -1)"

    echo "All prerequisites satisfied"
}

SECURITY_TOOLS_DIR="./bin/tools"
GITLEAKS_VERSION="v8.30.1"
GOVULNCHECK_VERSION="v1.9.0"
GOSEC_VERSION="v2.30.0"

install_tool() {
    # Installs a Go tool at a pinned version into ./bin/tools, once.
    local name="$1" pkg="$2"
    if [[ ! -x "${SECURITY_TOOLS_DIR}/${name}" ]]; then
        echo "security: installing ${pkg}"
        GOBIN="$(pwd)/${SECURITY_TOOLS_DIR#./}" go install "$pkg"
    fi
}

cmd_security() {
    # Scans for secrets (tree and every commit), known vulnerabilities and Go
    # security defects in both modules, and names that must not appear. The
    # names are a regular expression in BANNED_NAMES or in the file
    # BANNED_NAMES_FILE, kept out of the repository so the check does not
    # publish the names it guards.
    mkdir -p "$SECURITY_TOOLS_DIR"
    install_tool gitleaks "github.com/zricethezav/gitleaks/v8@${GITLEAKS_VERSION}"
    install_tool govulncheck "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
    install_tool gosec "github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}"
    local tools failed=0
    tools="$(pwd)/${SECURITY_TOOLS_DIR#./}"

    echo "security: gitleaks (working tree)"
    "${tools}/gitleaks" dir . --no-banner --redact --exit-code 1 || failed=1
    echo "security: gitleaks (every commit)"
    "${tools}/gitleaks" git . --no-banner --redact --exit-code 1 --log-opts=--all || failed=1
    local dir
    for dir in . example; do
        echo "security: govulncheck (${dir})"
        (cd "$dir" && "${tools}/govulncheck" ./...) || failed=1
        echo "security: gosec (${dir})"
        (cd "$dir" && "${tools}/gosec" -quiet ./...) || failed=1
    done

    local names="${BANNED_NAMES:-}"
    local names_file="${BANNED_NAMES_FILE:-${HOME}/.config/stickers/banned-names}"
    if [[ -z "$names" && -r "$names_file" ]]; then
        names="$(grep -v '^[[:space:]]*\(#\|$\)' "$names_file" | paste -sd'|' -)"
    fi
    if [[ -z "$names" ]]; then
        echo "security: name check skipped: set BANNED_NAMES or create ${names_file}"
    else
        echo "security: banned names"
        if git grep -nIiE "$names" -- . >/dev/null; then
            echo "security: FAIL: a banned name appears in a tracked file:" >&2
            git grep -nIiE "$names" -- . | cut -d: -f1,2 >&2
            failed=1
        fi
        if git log --all --format='%an %ae %cn %ce%n%B' | grep -qiE "$names"; then
            echo "security: FAIL: a banned name appears in commit metadata" >&2
            failed=1
        fi
    fi

    if [[ $failed -ne 0 ]]; then
        echo "security: FAIL" >&2
        exit 1
    fi
    echo "security: ok"
}

# ---------------------------------------------------------------------------
# Dispatcher — one entry per makehelp.sh-delegated target
# ---------------------------------------------------------------------------

case "${1:-}" in
    prereqs) cmd_prereqs ;;
    security) cmd_security ;;
    *)
        echo "Usage: $0 {prereqs|security}" >&2
        exit 1
        ;;
esac
