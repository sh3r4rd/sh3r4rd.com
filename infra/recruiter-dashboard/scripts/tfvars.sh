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

  # Canonicalize the local file to the same one-trailing-newline form that
  # emit_remote produces, so the comparison below and the uploaded value match
  # the remote round-trip exactly. Without this, a local file saved without a
  # trailing newline would create a permanent phantom diff that "Already up to
  # date" could never satisfy. Command substitution strips trailing newlines;
  # emit_remote re-adds exactly one.
  local local_canon
  local_canon="$(cat "$TFVARS")"
  TMPFILE="$(mktemp)"
  emit_remote "$local_canon" >"$TMPFILE"

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

  if emit_remote "$remote" | diff -u - "$TFVARS" >/dev/null 2>&1; then
    echo "No drift — local terraform.tfvars matches SSM."
    return 0
  fi
  echo "Drift detected (remote <-> local):"
  emit_remote "$remote" | diff -u --label "ssm:$PARAM_NAME" --label "local:terraform.tfvars" - "$TFVARS" || true
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
