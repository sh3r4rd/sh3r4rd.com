#!/usr/bin/env bash
#
# tfvars.sh — sync infra/recruiter-dashboard/terraform.tfvars to/from AWS SSM.
#
# SSM Parameter Store is the durable, versioned source of truth for the
# git-ignored terraform.tfvars file. This script mirrors the file verbatim into
# a single SecureString parameter. Terraform itself is unaware of any of this.
#
#   pull   SSM -> local file (backs up + confirms before overwriting local edits)
#   push   local file -> SSM (refuses empty/missing local; confirms before overwrite)
#   diff   show local-vs-remote differences; never writes
#
# Exit codes:
#   0  success — for `diff`, specifically means "no drift"
#   1  drift detected (`diff`), user aborted, or a refused/guarded operation
#   2  usage error
#   *  other non-zero: AWS CLI or plumbing failure
#
# Region is hardcoded so the script never depends on AWS_DEFAULT_REGION, but it
# still respects AWS_PROFILE for credential selection.
set -euo pipefail

PARAM_NAME="/recruiter-dashboard/tfvars"
REGION="us-east-1"

# Resolve terraform.tfvars to an absolute path relative to this script's own
# location (scripts/ lives one level under the terraform root), so the script
# works regardless of the caller's working directory.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TFVARS="$(cd "$SCRIPT_DIR/.." && pwd)/terraform.tfvars"
BACKUP="$TFVARS.bak"

# Temp file cleaned up on exit (set by pull). Initialized empty for `set -u`.
TMPFILE=""
cleanup() { [[ -n "$TMPFILE" ]] && rm -f "$TMPFILE"; return 0; }
trap cleanup EXIT

usage() {
  echo "Usage: tfvars.sh {pull|push|diff} [--force]" >&2
  exit 2
}

# Fetch the raw remote parameter value on stdout (no trailing-newline guarantee:
# command substitution in callers strips trailing newlines anyway). Callers must
# re-add exactly one trailing newline via `emit_remote` when writing/comparing,
# so round-trips match the canonical terraform.tfvars form (one trailing \n).
#
# Returns 10 if the parameter does not exist yet; any other non-zero on error.
read_remote() {
  local out
  if ! out="$(aws ssm get-parameter \
      --name "$PARAM_NAME" \
      --with-decryption \
      --region "$REGION" \
      --query Parameter.Value \
      --output text 2>&1)"; then
    if printf '%s' "$out" | grep -q "ParameterNotFound"; then
      return 10
    fi
    echo "$out" >&2
    return 1
  fi
  printf '%s' "$out"
}

# Print a captured remote value normalized to exactly one trailing newline.
emit_remote() { printf '%s\n' "$1"; }

# Print the local file in canonical form: LF-only line endings, exactly one
# trailing newline. This is the form `push` uploads and the form every
# comparison must use.
#
# Two normalizations, both required for a stable round-trip:
#   1. CR stripping — the AWS CLI's file:// reader converts CRLF to LF on
#      upload, so a CR can never survive to the remote. Comparing the raw local
#      file against the remote would therefore report drift forever, and push
#      could never reach "Already up to date".
#   2. Trailing newlines — command substitution strips *all* of them and
#      emit_remote re-adds exactly one, so trailing blank lines are collapsed.
#      A file saved without a trailing newline would otherwise be a permanent
#      phantom diff too.
canon_local() { emit_remote "$(tr -d '\r' <"$TFVARS")"; }

# Warn when canon_local will silently alter the file's line endings, so the
# normalization is visible rather than surprising.
warn_if_crlf() {
  if LC_ALL=C grep -q $'\r' "$TFVARS" 2>/dev/null; then
    echo "Note: terraform.tfvars contains CR characters; they are stripped (SSM stores LF only)." >&2
  fi
}

confirm() {
  # $1 = prompt. Honors a global FORCE flag for non-interactive use.
  if [[ "${FORCE:-0}" == "1" ]]; then
    return 0
  fi
  local reply
  read -r -p "$1 [y/N] " reply
  [[ "$reply" == "y" || "$reply" == "Y" ]]
}

cmd_pull() {
  local remote rc
  set +e
  remote="$(read_remote)"
  rc=$?
  set -e
  if [[ $rc -eq 10 ]]; then
    echo "Parameter $PARAM_NAME not seeded yet — run 'make tf-vars-push' first." >&2
    exit 1
  elif [[ $rc -ne 0 ]]; then
    exit $rc
  fi

  TMPFILE="$(mktemp)"
  emit_remote "$remote" >"$TMPFILE"

  if [[ -f "$TFVARS" ]]; then
    if diff -u "$TFVARS" "$TMPFILE" >/dev/null 2>&1; then
      echo "Local terraform.tfvars already matches SSM. Nothing to do."
      return 0
    fi
    echo "Local terraform.tfvars differs from SSM:"
    diff -u "$TFVARS" "$TMPFILE" || true
    echo
    if ! confirm "Overwrite local file? (a backup will be written to terraform.tfvars.bak)"; then
      echo "Aborted." >&2
      exit 1
    fi
    cp "$TFVARS" "$BACKUP"
    echo "Backed up current local file to terraform.tfvars.bak"
  fi

  cp "$TMPFILE" "$TFVARS"
  echo "Pulled $PARAM_NAME -> terraform.tfvars"
}

cmd_push() {
  if [[ ! -s "$TFVARS" ]]; then
    echo "Refusing to push: terraform.tfvars is missing or empty (would wipe remote)." >&2
    exit 1
  fi

  # Upload the canonical form (see canon_local), so what is compared below is
  # byte-for-byte what SSM will store and what a later pull will write back.
  warn_if_crlf
  TMPFILE="$(mktemp)"
  canon_local >"$TMPFILE"

  local remote rc
  set +e
  remote="$(read_remote)"
  rc=$?
  set -e

  if [[ $rc -eq 0 ]]; then
    if emit_remote "$remote" | diff -u - "$TMPFILE" >/dev/null 2>&1; then
      echo "Already up to date — local matches SSM."
      return 0
    fi
    echo "Local terraform.tfvars differs from SSM (remote <-> local):"
    emit_remote "$remote" | diff -u --label "ssm:$PARAM_NAME" --label "local:terraform.tfvars" - "$TMPFILE" || true
    echo
  elif [[ $rc -eq 10 ]]; then
    echo "Parameter $PARAM_NAME does not exist yet — this push will create it (version 1)."
  else
    exit $rc
  fi

  if ! confirm "Push local terraform.tfvars to $PARAM_NAME?"; then
    echo "Aborted." >&2
    exit 1
  fi

  aws ssm put-parameter \
    --name "$PARAM_NAME" \
    --type SecureString \
    --value file://"$TMPFILE" \
    --tier Intelligent-Tiering \
    --overwrite \
    --region "$REGION" >/dev/null
  echo "Pushed terraform.tfvars -> $PARAM_NAME"
}

cmd_diff() {
  local remote rc
  set +e
  remote="$(read_remote)"
  rc=$?
  set -e

  if [[ $rc -eq 10 ]]; then
    echo "Parameter $PARAM_NAME not seeded yet — run 'make tf-vars-push' first." >&2
    exit 1
  elif [[ $rc -ne 0 ]]; then
    exit $rc
  fi

  if [[ ! -f "$TFVARS" ]]; then
    echo "No local terraform.tfvars — run 'make tf-vars-pull' to fetch it." >&2
    exit 1
  fi

  # Compare the canonical local form, not the raw file: push uploads the
  # canonical form, so raw-file differences that push would normalize away are
  # not drift and must not be reported as such.
  warn_if_crlf
  TMPFILE="$(mktemp)"
  canon_local >"$TMPFILE"

  if emit_remote "$remote" | diff -u - "$TMPFILE" >/dev/null 2>&1; then
    echo "No drift — local terraform.tfvars matches SSM."
    return 0
  fi
  echo "Drift detected (remote <-> local):"
  emit_remote "$remote" | diff -u --label "ssm:$PARAM_NAME" --label "local:terraform.tfvars" - "$TMPFILE" || true
  # Exit non-zero so `make tf-vars-diff` is usable as a check in scripts, CI, or
  # a pre-apply guard. A `make ... Error 1` here is the intended drift signal.
  exit 1
}

main() {
  local sub="${1:-}"
  [[ -n "$sub" ]] || usage
  shift || true

  FORCE=0
  for arg in "$@"; do
    case "$arg" in
      --force) FORCE=1 ;;
      *) echo "Unknown option: $arg" >&2; usage ;;
    esac
  done

  case "$sub" in
    pull) cmd_pull ;;
    push) cmd_push ;;
    diff) cmd_diff ;;
    *) usage ;;
  esac
}

main "$@"
