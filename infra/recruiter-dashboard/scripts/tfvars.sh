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
#   3  cannot compare — SSM parameter not seeded, or no local terraform.tfvars
#   4  operational failure — AWS CLI, or `diff` itself failed
#
# 1 means *only* "the two sides differ" (or the user said no). A gate such as
# `make tf-vars-diff || remediate` must be able to tell real drift apart from a
# fresh checkout that has not pulled yet (3) and from expired credentials (4),
# since those three need opposite remedies.
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
# Returns 10 if the parameter does not exist yet (an internal sentinel callers
# translate into their own message); 4 for any other AWS CLI failure, so a
# credential or network error never reaches a caller as the drift code 1.
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
    return 4
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

# Warn when canon_local's normalization is in play, so it is visible rather than
# surprising. Worded to hold for all three subcommands: push strips the CRs on
# upload, diff and pull only ignore them when comparing (pull leaves the bytes
# on disk untouched — see cmd_pull).
warn_if_crlf() {
  if LC_ALL=C grep -q $'\r' "$TFVARS" 2>/dev/null; then
    echo "Note: terraform.tfvars contains CR characters; SSM stores LF only, so they are ignored when comparing and stripped on push." >&2
  fi
}

# Echo a unique, timestamped backup path for the local file.
#
# Backups are per-pull rather than a single fixed terraform.tfvars.bak slot: two
# pulls in a row would otherwise overwrite the first backup and silently destroy
# local edits that were never pushed. SSM version history is the primary
# recovery path; these cover local-only state it never saw. They are git-ignored
# and never pruned automatically — delete them by hand when you're done.
backup_path() {
  local stamp base candidate n
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  base="$TFVARS.$stamp"
  candidate="$base.bak"
  # Two pulls within the same second would collide; disambiguate rather than
  # clobber, since not clobbering is the entire point of this function.
  n=1
  while [[ -e "$candidate" ]]; do
    candidate="$base-$n.bak"
    n=$((n + 1))
  done
  printf '%s' "$candidate"
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
    exit 3
  elif [[ $rc -ne 0 ]]; then
    exit $rc
  fi

  TMPFILE="$(mktemp)"
  emit_remote "$remote" >"$TMPFILE"

  if [[ -f "$TFVARS" ]]; then
    # Compare the canonical local form (see canon_local), not the raw file. A
    # CRLF or trailing-blank-line difference is normalized away by push, so it
    # is not a real difference — and comparing raw would make a Windows-edited
    # file prompt and write a fresh timestamped backup on *every* pull, forever,
    # accumulating silently under TFVARS_ARGS=--force. The cost is that such a
    # file keeps its CR bytes on disk instead of being rewritten; Terraform
    # reads it identically either way, and push still uploads LF only.
    warn_if_crlf
    if canon_local | diff -u - "$TMPFILE" >/dev/null 2>&1; then
      echo "Local terraform.tfvars already matches SSM. Nothing to do."
      return 0
    fi
    echo "Local terraform.tfvars differs from SSM:"
    canon_local | diff -u --label "local:terraform.tfvars" --label "ssm:$PARAM_NAME" - "$TMPFILE" || true
    echo
    if ! confirm "Overwrite local file? (a timestamped backup will be written alongside it)"; then
      echo "Aborted." >&2
      exit 1
    fi
    local backup
    backup="$(backup_path)"
    cp "$TFVARS" "$backup"
    echo "Backed up current local file to $(basename "$backup")"
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
  # Note this normalizes the file: CRs are stripped and any trailing blank lines
  # collapse to a single trailing newline. Both are required for a stable
  # round-trip, so the uploaded value may differ from the raw file on disk.
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

  # --value file:// (not an inline value) keeps the secret out of argv.
  # Intelligent-Tiering auto-promotes to Advanced past 4096 bytes, which costs
  # $0.05/parameter/month; under that it stays on the free Standard tier. The
  # file is ~1.3 KB today, so this is only a concern if tfvars grows a lot.
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

  # Both "not seeded" and "no local file" mean there is nothing to compare, not
  # that the two sides differ — exit 3, so a gate does not mistake a fresh
  # checkout for drift and try to remediate it as one.
  if [[ $rc -eq 10 ]]; then
    echo "Parameter $PARAM_NAME not seeded yet — run 'make tf-vars-push' first." >&2
    exit 3
  elif [[ $rc -ne 0 ]]; then
    exit $rc
  fi

  if [[ ! -f "$TFVARS" ]]; then
    echo "No local terraform.tfvars — run 'make tf-vars-pull' to fetch it." >&2
    exit 3
  fi

  # Compare the canonical local form, not the raw file: push uploads the
  # canonical form, so raw-file differences that push would normalize away are
  # not drift and must not be reported as such.
  warn_if_crlf
  TMPFILE="$(mktemp)"
  canon_local >"$TMPFILE"

  # Run diff once and branch on its exact status. `if diff ...` would collapse
  # status 2 (diff itself failed — unreadable input, etc.) into the drift path,
  # reporting a plumbing failure as drift to whatever is gating on this.
  local out status
  set +e
  out="$(emit_remote "$remote" | diff -u --label "ssm:$PARAM_NAME" --label "local:terraform.tfvars" - "$TMPFILE" 2>&1)"
  status=$?
  set -e

  if [[ $status -eq 0 ]]; then
    echo "No drift — local terraform.tfvars matches SSM."
    return 0
  elif [[ $status -ne 1 ]]; then
    echo "Comparison failed (diff exited $status):" >&2
    printf '%s\n' "$out" >&2
    exit 4
  fi

  echo "Drift detected (remote <-> local):"
  printf '%s\n' "$out"
  # Exit 1 so `make tf-vars-diff` is usable as a check in scripts, CI, or a
  # pre-apply guard. A `make ... Error 1` here is the intended drift signal;
  # 3 and 4 mean the comparison never happened (see the header).
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
