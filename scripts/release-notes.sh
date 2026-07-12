#!/usr/bin/env bash
#
# gen-release-notes.sh
#
# Generates release notes for the range between the most recent GitHub release
# and the current branch head, grouping commits by their `Change-Category` git
# trailer. The `Change-Impact` trailer is used only to derive the recommended
# semantic-version bump (breaking -> major, additive -> minor, otherwise patch).
#
# Usage:
#   scripts/gen-release-notes.sh [--to <ref>] [--from <tag>] [--by-component]
#
# Options:
#   --from <tag>     Override the "since" tag (defaults to the latest GitHub
#                    release discovered via `gh`).
#   --to <ref>       Override the "until" ref (defaults to HEAD).
#   --by-component   Group the report by `Component` first, then by category.
#                    Commits touching multiple components appear under each.

set -euo pipefail

# Record and unit separators used to delimit the machine-readable git log so
# that subjects and trailer values can safely contain arbitrary punctuation.
readonly RS=$'\x1e'
readonly US=$'\x1f'

# Category rendering order and human-facing section titles. Any category not
# present here is bucketed under "uncategorized" and rendered last.
readonly CATEGORY_ORDER=(feature bugfix security deprecation internal uncategorized)

declare -A CATEGORY_TITLE=(
  [feature]="Features"
  [bugfix]="Bug Fixes"
  [security]="Security"
  [deprecation]="Deprecations"
  [internal]="Internal"
  [uncategorized]="Uncategorized"
)

# Aliases for historical or mistaken `Change-Category` values, mapped onto their
# canonical bucket.
declare -A CATEGORY_ALIAS=(
  [removal]="deprecation"
  [improvement]="feature"
)

# Label used for commits that carry no `Component` trailer.
readonly NO_COMPONENT="(unscoped)"

# Parsed commits, stored as parallel arrays indexed by commit.
declare -a COMMIT_CATEGORY=()
declare -a COMMIT_LINE=()
declare -a COMMIT_COMPONENTS=()

# Highest impact observed across the range, used to derive the version bump.
declare -i HAS_BREAKING=0
declare -i HAS_ADDITIVE=0

function die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

function require_tools() {
  local tool
  for tool in git gh; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      die "required tool '$tool' not found on PATH"
    fi
  done
}

# Prints the tag name of the most recent GitHub release, or an empty string if
# the repository has no releases yet. Fails only when the query itself cannot be
# performed (e.g. `gh` is unauthenticated or the repo cannot be resolved).
function latest_release_tag() {
  local tag
  if ! tag=$(gh release list --limit 1 --json tagName --jq '.[0].tagName // ""' 2>/dev/null); then
    die "unable to query GitHub releases (is 'gh' authenticated for this repo?)"
  fi
  printf '%s\n' "$tag"
}

# Normalizes a raw category trailer value into a known bucket. Unknown or empty
# values collapse to "uncategorized".
function normalize_category() {
  local category="${1,,}"
  if [[ -n "$category" && -v "CATEGORY_ALIAS[$category]" ]]; then
    category="${CATEGORY_ALIAS[$category]}"
  fi
  if [[ -n "$category" && -v "CATEGORY_TITLE[$category]" ]]; then
    printf '%s\n' "$category"
  else
    printf 'uncategorized\n'
  fi
}

# Reads the commit range into the COMMIT_* arrays and records the highest impact
# seen. Arguments: <from-tag> <to-ref>. An empty <from-tag> selects the entire
# history reachable from <to-ref> (i.e. the first-release case).
function collect_commits() {
  local from="$1" to="$2"
  local format record hash subject category impact component
  local -a range

  if [[ -n "$from" ]]; then
    range=("${from}..${to}")
  else
    range=("$to")
  fi

  # Each record is prefixed with RS; fields are separated by US.
  format="${RS}%H${US}%s${US}"
  format+="%(trailers:key=Change-Category,valueonly,separator=%x2c)${US}"
  format+="%(trailers:key=Change-Impact,valueonly,separator=%x2c)${US}"
  format+="%(trailers:key=Component,valueonly,separator=%x2c)"

  while IFS= read -r -d "$RS" record; do
    # The read before the first RS yields an empty leading record; skip it.
    [[ -z "$record" ]] && continue

    IFS="$US" read -r hash subject category impact component <<<"$record"

    case "${impact,,}" in
      breaking) HAS_BREAKING=1 ;;
      additive) HAS_ADDITIVE=1 ;;
    esac

    COMMIT_CATEGORY+=("$(normalize_category "$category")")
    COMMIT_LINE+=("- ${subject} (${hash:0:8})")
    COMMIT_COMPONENTS+=("$component")
  done < <(git log --no-merges --format="$format" "${range[@]}")
}

# Prints the recommended semantic-version bump derived from the observed impacts.
function version_bump() {
  if ((HAS_BREAKING)); then
    printf 'major\n'
  elif ((HAS_ADDITIVE)); then
    printf 'minor\n'
  else
    printf 'patch\n'
  fi
}

# Splits a raw comma-separated `Component` value into one trimmed component per
# line, emitting NO_COMPONENT when the value is empty.
function split_components() {
  local raw="$1" token
  local -a parts
  local -i any=0

  IFS=',' read -ra parts <<<"$raw"
  for token in "${parts[@]}"; do
    token="${token//[[:space:]]/}"
    [[ -z "$token" ]] && continue
    printf '%s\n' "$token"
    any=1
  done
  ((any)) || printf '%s\n' "$NO_COMPONENT"
}

# Succeeds if commit <index> belongs to component <name>.
function commit_in_component() {
  local index="$1" name="$2" token
  while IFS= read -r token; do
    [[ "$token" == "$name" ]] && return 0
  done < <(split_components "${COMMIT_COMPONENTS[index]}")
  return 1
}

# Renders the report grouped by category (the default layout).
function render_by_category() {
  local category first i
  for category in "${CATEGORY_ORDER[@]}"; do
    first=1
    for ((i = 0; i < ${#COMMIT_LINE[@]}; i++)); do
      if [[ "${COMMIT_CATEGORY[i]}" == "$category" ]]; then
        if ((first)); then
          printf '\n## %s\n\n' "${CATEGORY_TITLE[$category]}"
          first=0
        fi
        printf '%s\n' "${COMMIT_LINE[i]}"
      fi
    done
  done
}

# Renders a single component heading followed by its category sub-sections,
# limited to commits belonging to that component. Argument: <component>.
function render_component_block() {
  local component="$1" category first i

  printf '\n## %s\n' "$component"
  for category in "${CATEGORY_ORDER[@]}"; do
    first=1
    for ((i = 0; i < ${#COMMIT_LINE[@]}; i++)); do
      if [[ "${COMMIT_CATEGORY[i]}" == "$category" ]] && commit_in_component "$i" "$component"; then
        if ((first)); then
          printf '\n### %s\n\n' "${CATEGORY_TITLE[$category]}"
          first=0
        fi
        printf '%s\n' "${COMMIT_LINE[i]}"
      fi
    done
  done
}

# Renders the report grouped by component first, then category. Commits touching
# multiple components are listed under each of them.
function render_by_component() {
  local -A seen=()
  local -a components=()
  local i token component

  for ((i = 0; i < ${#COMMIT_LINE[@]}; i++)); do
    while IFS= read -r token; do
      if [[ ! -v "seen[$token]" ]]; then
        seen[$token]=1
        components+=("$token")
      fi
    done < <(split_components "${COMMIT_COMPONENTS[i]}")
  done

  ((${#components[@]} > 0)) || return 0

  mapfile -t components < <(printf '%s\n' "${components[@]}" | sort)

  # Named components first (alphabetical), with the unscoped bucket kept last.
  for component in "${components[@]}"; do
    [[ "$component" == "$NO_COMPONENT" ]] && continue
    render_component_block "$component"
  done
  if [[ -v "seen[$NO_COMPONENT]" ]]; then
    render_component_block "$NO_COMPONENT"
  fi
}

# Renders the full report to stdout. Arguments: <from> <to> <group-by>.
function render() {
  local from="$1" to="$2" group_by="$3"

  printf '# Release Notes\n\n'
  if [[ -n "$from" ]]; then
    printf '_Changes from **%s** to **%s**._\n' "$from" "$to"
  else
    printf '_Initial release — all changes up to **%s**._\n' "$to"
  fi
  printf '_Recommended version bump: **%s**._\n' "$(version_bump)"

  if [[ "$group_by" == "component" ]]; then
    render_by_component
  else
    render_by_category
  fi
}

function main() {
  local from="" to="HEAD" group_by="category"

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --from)
        [[ $# -ge 2 ]] || die "--from requires an argument"
        from="$2"
        shift 2
        ;;
      --to)
        [[ $# -ge 2 ]] || die "--to requires an argument"
        to="$2"
        shift 2
        ;;
      --by-component)
        group_by="component"
        shift
        ;;
      -h | --help)
        grep '^#' "$0" | sed 's/^# \{0,1\}//'
        exit 0
        ;;
      *)
        die "unknown argument: $1"
        ;;
    esac
  done

  require_tools

  if [[ -z "$from" ]]; then
    from=$(latest_release_tag)
  fi

  if [[ -n "$from" ]] && ! git rev-parse --verify --quiet "$from" >/dev/null; then
    die "cannot resolve 'from' ref: $from"
  fi
  if ! git rev-parse --verify --quiet "$to" >/dev/null; then
    die "cannot resolve 'to' ref: $to"
  fi

  collect_commits "$from" "$to"
  render "$from" "$to" "$group_by"
}

main "$@"
