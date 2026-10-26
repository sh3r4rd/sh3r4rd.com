#!/usr/bin/env bash
# ri-issues.sh: create the Recruiter Intelligence tracking scaffolding on GitHub
# from docs/recruiter-intelligence/issues.json. See docs/recruiter-intelligence/TRACKING.md.
#
#   scripts/ri-issues.sh                 # dry run: print every action, change nothing
#   scripts/ri-issues.sh --apply         # perform them
#   scripts/ri-issues.sh --apply --phase project   # one phase only
#
# Phases, in order: labels milestones issues links existing project
#
# Idempotent: labels and milestones are matched by name, issues by exact title,
# sub-issue parents and blocked-by links are checked before being added, and the
# Project is matched by title. Created issue numbers (and the Project number) are
# written back to the manifest on --apply so {{ref:KEY}} placeholders resolve.
#
# Needs gh (token scopes: repo, project) and jq.

set -euo pipefail

ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
DOCS="$ROOT/docs/recruiter-intelligence"
MANIFEST="$DOCS/issues.json"

APPLY=0
ONLY_PHASE=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply) APPLY=1 ;;
    --phase) ONLY_PHASE="$2"; shift ;;
    -h|--help) sed -n '2,16p' "$0"; exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
  shift
done

command -v gh >/dev/null || { echo "gh is required" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

REPO="$(jq -r .repo "$MANIFEST")"
OWNER="${REPO%%/*}"
NAME="${REPO##*/}"

say()  { printf '%s\n' "$*"; }
act()  { if [[ $APPLY -eq 1 ]]; then printf '  DO    %s\n' "$*"; else printf '  WOULD %s\n' "$*"; fi; }
skip() { printf '  OK    %s\n' "$*"; }
want_phase() { [[ -z "$ONLY_PHASE" || "$ONLY_PHASE" == "$1" ]]; }

# Write a value back to the manifest (only on --apply).
manifest_set() { # jq-path-expression value-json
  [[ $APPLY -eq 1 ]] || return 0
  local tmp; tmp="$(mktemp)"
  jq "$1 = $2" "$MANIFEST" >"$tmp" && mv "$tmp" "$MANIFEST"
}

# --- lookups -----------------------------------------------------------------

# macOS ships bash 3.2 (no associative arrays), so lookups live in tab-separated files.
STATE_DIR="$(mktemp -d)"
trap 'rm -rf "$STATE_DIR"' EXIT
TITLES="$STATE_DIR/titles"   # number<TAB>title, every issue in the repo
KEYS="$STATE_DIR/keys"       # manifest key<TAB>issue number
: >"$KEYS"

title_num()  { awk -F'\t' -v t="$1" '$2 == t { print $1; exit }' "$TITLES"; }
key_num()    { awk -F'\t' -v k="$1" '$1 == k { n = $2 } END { if (n != "") print n }' "$KEYS"; }
set_key()    { printf '%s\t%s\n' "$1" "$2" >>"$KEYS"; }
known_nums() { cut -f2 "$KEYS" | sort -un; }

load_titles() {
  gh api --paginate "repos/$REPO/issues?state=all&per_page=100" \
    --jq '.[] | select(.pull_request == null) | [.number, .title] | @tsv' >"$TITLES"
}

resolve_keys() {
  local key title num found
  while IFS=$'\t' read -r key title num; do
    if [[ "$num" != "null" ]]; then
      set_key "$key" "$num"
    else
      found="$(title_num "$title")"
      if [[ -n "$found" ]]; then set_key "$key" "$found"; fi
    fi
  done < <(jq -r '.issues[] | [.key, .title, (.number|tostring)] | @tsv' "$MANIFEST")
}

ref() { # key -> "#123", or "#<KEY>" when not yet created (dry run)
  local n
  if [[ "$1" == \#* ]]; then printf '%s' "$1"; return; fi
  n="$(key_num "$1")"
  if [[ -n "$n" ]]; then printf '#%s' "$n"; else printf '#<%s>' "$1"; fi
}

render_body() { # body-file -> stdout with {{ref:KEY}} substituted
  local text key
  text="$(cat "$DOCS/$1")"
  for key in $(grep -o '{{ref:[A-Z0-9-]*}}' "$DOCS/$1" | sed 's/{{ref:\(.*\)}}/\1/' | sort -u); do
    text="${text//\{\{ref:$key\}\}/$(ref "$key")}"
  done
  printf '%s\n' "$text"
}

issue_id()      { gh api "repos/$REPO/issues/$1" --jq .id; }
issue_node_id() { gh api "repos/$REPO/issues/$1" --jq .node_id; }
parent_of()     { { gh api "repos/$REPO/issues/$1/parent" 2>/dev/null || true; } | jq -r '.number // empty'; }

# --- phases ------------------------------------------------------------------

phase_labels() {
  say "== labels"
  local existing name color desc
  existing="$(gh label list --repo "$REPO" --limit 300 --json name --jq '.[].name')"
  while IFS=$'\t' read -r name color desc; do
    if grep -qxF "$name" <<<"$existing"; then
      skip "label $name exists"
    else
      act "create label $name (#$color)"
      if [[ $APPLY -eq 1 ]]; then gh label create "$name" --repo "$REPO" --color "$color" --description "$desc" >/dev/null; fi
    fi
  done < <(jq -r '.labels[] | [.name, .color, .description] | @tsv' "$MANIFEST")
}

phase_milestones() {
  say "== milestones"
  local existing title due desc
  existing="$(gh api --paginate "repos/$REPO/milestones?state=all&per_page=100" --jq '.[].title')"
  while IFS=$'\t' read -r title due desc; do
    if grep -qxF "$title" <<<"$existing"; then
      skip "milestone $title exists"
    else
      act "create milestone $title due $due"
      if [[ $APPLY -eq 1 ]]; then
        gh api "repos/$REPO/milestones" -f title="$title" -f due_on="$due" \
        -f description="$desc" >/dev/null
      fi
    fi
  done < <(jq -r '.milestones[] | [.title, .due_on, .description] | @tsv' "$MANIFEST")
}

phase_issues() {
  say "== issues"
  local i key title milestone labels body_file num url tmp
  local count; count="$(jq '.issues | length' "$MANIFEST")"
  for ((i = 0; i < count; i++)); do
    key="$(jq -r ".issues[$i].key" "$MANIFEST")"
    title="$(jq -r ".issues[$i].title" "$MANIFEST")"
    milestone="$(jq -r ".issues[$i].milestone // empty" "$MANIFEST")"
    labels="$(jq -r ".issues[$i].labels | join(\",\")" "$MANIFEST")"
    body_file="$(jq -r ".issues[$i].body_file" "$MANIFEST")"
    num="$(key_num "$key")"
    if [[ -n "$num" ]]; then
      skip "$key exists as #$num  $title"
      manifest_set ".issues[$i].number" "$num"
      continue
    fi
    act "create $key [${milestone:-no milestone}] [$labels]  $title"
    if [[ $APPLY -eq 1 ]]; then
      tmp="$(mktemp)"
      render_body "$body_file" >"$tmp"
      args=(--repo "$REPO" --title "$title" --body-file "$tmp" --label "$labels")
      if [[ -n "$milestone" ]]; then args+=(--milestone "$milestone"); fi
      url="$(gh issue create "${args[@]}")"
      rm -f "$tmp"
      num="${url##*/}"
      set_key "$key" "$num"
      manifest_set ".issues[$i].number" "$num"
      say "        -> #$num"
    fi
  done
}

phase_links() {
  say "== links (sub-issues, blocked-by)"
  local key parent blockers child pnum current b bnum existing
  while IFS=$'\t' read -r key parent blockers; do
    child="$(key_num "$key")"
    if [[ -n "$parent" ]]; then
      pnum="$(key_num "$parent")"
      current=""
      if [[ -n "$child" ]]; then current="$(parent_of "$child")"; fi
      if [[ -n "$child" && -n "$pnum" && "$current" == "$pnum" ]]; then
        skip "$(ref "$key") already a sub-issue of $(ref "$parent")"
      else
        act "add $(ref "$key") as sub-issue of $(ref "$parent")"
        if [[ $APPLY -eq 1 ]]; then
          gh api "repos/$REPO/issues/$pnum/sub_issues" \
          -F sub_issue_id="$(issue_id "$child")" >/dev/null
        fi
      fi
    fi
    if [[ -z "$blockers" ]]; then continue; fi
    existing=""
    if [[ -n "$child" ]]; then existing="$(gh api "repos/$REPO/issues/$child/dependencies/blocked_by" --jq '.[].number' 2>/dev/null || true)"; fi
    for b in ${blockers//,/ }; do
      if [[ "$b" == \#* ]]; then bnum="${b#\#}"; else bnum="$(key_num "$b")"; fi
      if [[ -n "$bnum" ]] && grep -qxF "$bnum" <<<"$existing"; then
        skip "$(ref "$key") already blocked by $(ref "$b")"
      else
        act "mark $(ref "$key") blocked by $(ref "$b")"
        if [[ $APPLY -eq 1 ]]; then
          gh api "repos/$REPO/issues/$child/dependencies/blocked_by" \
          -F issue_id="$(issue_id "$bnum")" >/dev/null
        fi
      fi
    done
  done < <(jq -r '.issues[] | [.key, (.parent // ""), (.blocked_by | join(","))] | @tsv' "$MANIFEST")
}

phase_existing() {
  say "== existing issues"
  local i count num parent milestone labels first_line append_file pnum current body cur_ms cur_labels l tmp
  count="$(jq '.existing | length' "$MANIFEST")"
  for ((i = 0; i < count; i++)); do
    num="$(jq -r ".existing[$i].number" "$MANIFEST")"
    parent="$(jq -r ".existing[$i].parent" "$MANIFEST")"
    milestone="$(jq -r ".existing[$i].milestone" "$MANIFEST")"
    labels="$(jq -r ".existing[$i].add_labels | join(\",\")" "$MANIFEST")"
    first_line="$(jq -r ".existing[$i].replace_first_line // empty" "$MANIFEST")"
    append_file="$(jq -r ".existing[$i].append_file // empty" "$MANIFEST")"
    pnum="$(key_num "$parent")"

    current="$(parent_of "$num")"
    if [[ -n "$pnum" && "$current" == "$pnum" ]]; then
      skip "#$num already a sub-issue of $(ref "$parent")"
    else
      act "move #$num under $(ref "$parent") (current parent: ${current:+#}${current:-none}; replace_parent)"
      if [[ $APPLY -eq 1 ]]; then
        gh api "repos/$REPO/issues/$pnum/sub_issues" \
        -F sub_issue_id="$(issue_id "$num")" -F replace_parent=true >/dev/null
      fi
    fi

    cur_ms="$(gh api "repos/$REPO/issues/$num" --jq '.milestone.title // ""')"
    if [[ "$cur_ms" == "$milestone" ]]; then
      skip "#$num milestone $milestone"
    else
      act "set #$num milestone $milestone"
      if [[ $APPLY -eq 1 ]]; then gh issue edit "$num" --repo "$REPO" --milestone "$milestone" >/dev/null; fi
    fi

    cur_labels="$(gh api "repos/$REPO/issues/$num" --jq '.labels[].name')"
    for l in ${labels//,/ }; do
      if grep -qxF "$l" <<<"$cur_labels"; then skip "#$num has label $l"; else
        act "add label $l to #$num"
        if [[ $APPLY -eq 1 ]]; then gh issue edit "$num" --repo "$REPO" --add-label "$l" >/dev/null; fi
      fi
    done

    if [[ -n "$first_line" || -n "$append_file" ]]; then
      body="$(gh api "repos/$REPO/issues/$num" --jq .body)"
      local new="$body"
      if [[ -n "$first_line" ]]; then
        local fl="${first_line//\{\{ref:$parent\}\}/$(ref "$parent")}"
        if [[ "$(head -n1 <<<"$new")" != "$fl" ]]; then new="$fl"$'\n'"$(tail -n +2 <<<"$new")"; fi
      fi
      if [[ -n "$append_file" ]]; then
        local marker; marker="$(grep -m1 '^## ' "$DOCS/$append_file")"
        grep -qxF "$marker" <<<"$new" || new="$new"$'\n'"$(render_body "$append_file")"
      fi
      if [[ "$new" == "$body" ]]; then
        skip "#$num body up to date"
      else
        act "edit #$num body (first line: $(head -n1 <<<"$new"))"
        if [[ $APPLY -eq 1 ]]; then
          tmp="$(mktemp)"; printf '%s\n' "$new" >"$tmp"
          gh issue edit "$num" --repo "$REPO" --body-file "$tmp" >/dev/null
          rm -f "$tmp"
        fi
      fi
    fi
  done
}

gql() { gh api graphql "$@"; }

phase_project() {
  say "== project"
  local title proj_num proj_id viewer_id repo_id
  title="$(jq -r .project.title "$MANIFEST")"
  viewer_id="$(gql -f query='{viewer{id}}' --jq .data.viewer.id)"
  repo_id="$(gql -f query='query($o:String!,$n:String!){repository(owner:$o,name:$n){id}}' \
               -f o="$OWNER" -f n="$NAME" --jq .data.repository.id)"
  read -r proj_num proj_id < <(gql -f query='{viewer{projectsV2(first:100){nodes{number id title}}}}' \
    --jq ".data.viewer.projectsV2.nodes[] | select(.title == \"$title\") | \"\(.number) \(.id)\"" | head -n1) || true

  if [[ -n "${proj_id:-}" ]]; then
    skip "project \"$title\" exists (#$proj_num)"
  else
    act "create project \"$title\", linked to $REPO, public"
    if [[ $APPLY -eq 1 ]]; then
      read -r proj_num proj_id < <(gql -f query='mutation($o:ID!,$t:String!,$r:ID!){createProjectV2(input:{ownerId:$o,title:$t,repositoryId:$r}){projectV2{number id}}}' \
        -f o="$viewer_id" -f t="$title" -f r="$repo_id" --jq '.data.createProjectV2.projectV2 | "\(.number) \(.id)"')
      gql -f query='mutation($p:ID!){updateProjectV2(input:{projectId:$p,public:true,shortDescription:"12-week GenAI project on the Recruiter Dashboard, alongside AIP-C01 prep"}){projectV2{id}}}' \
        -f p="$proj_id" >/dev/null
    else
      say "  (dry run: field and item steps below assume the project exists)"
    fi
  fi
  manifest_set ".project.number" "${proj_num:-null}"

  # Fields
  local fields name type i n
  fields=""
  if [[ -n "${proj_id:-}" ]]; then
    fields="$(gql -f query='query($p:ID!){node(id:$p){...on ProjectV2{fields(first:50){nodes{...on ProjectV2FieldCommon{name}}}}}}' \
                                          -f p="$proj_id" --jq '.data.node.fields.nodes[].name')"
  fi
  n="$(jq '.project.fields | length' "$MANIFEST")"
  for ((i = 0; i < n; i++)); do
    name="$(jq -r ".project.fields[$i].name" "$MANIFEST")"
    type="$(jq -r ".project.fields[$i].type" "$MANIFEST")"
    if grep -qxF "$name" <<<"$fields"; then skip "field \"$name\" exists"; continue; fi
    act "create field \"$name\" ($type)"
    [[ $APPLY -eq 1 ]] || continue
    if [[ "$type" == "SINGLE_SELECT" ]]; then
      local opts
      opts="$(jq -c "[.project.fields[$i].options[] | {name: ., color: \"GRAY\", description: \"\"}]" "$MANIFEST")"
      gql --input <(jq -n --arg p "$proj_id" --arg n "$name" --argjson o "$opts" \
        '{query:"mutation($p:ID!,$n:String!,$o:[ProjectV2SingleSelectFieldOptionInput!]!){createProjectV2Field(input:{projectId:$p,dataType:SINGLE_SELECT,name:$n,singleSelectOptions:$o}){projectV2Field{...on ProjectV2SingleSelectField{id}}}}",variables:{p:$p,n:$n,o:$o}}') >/dev/null
    else
      gql -f query="mutation(\$p:ID!,\$n:String!){createProjectV2Field(input:{projectId:\$p,dataType:$type,name:\$n}){projectV2Field{...on ProjectV2Field{id}}}}" \
        -f p="$proj_id" -f n="$name" >/dev/null
    fi
  done

  # Items: every manifest issue plus the existing ones
  local nums key num
  nums="$(known_nums; jq -r '.existing[].number' "$MANIFEST")"
  local on_board=""
  if [[ -n "${proj_id:-}" ]]; then
    on_board="$(gql -f query='query($p:ID!){node(id:$p){...on ProjectV2{items(first:100){nodes{content{...on Issue{number}}}}}}}' \
                                          -f p="$proj_id" --jq '.data.node.items.nodes[].content.number')"
  fi
  for num in $(sort -n <<<"$nums"); do
    if grep -qxF "$num" <<<"$on_board"; then skip "#$num on the project"; else
      act "add #$num to the project"
      if [[ $APPLY -eq 1 ]]; then
        gql -f query='mutation($p:ID!,$c:ID!){addProjectV2ItemById(input:{projectId:$p,contentId:$c}){item{id}}}' \
        -f p="$proj_id" -f c="$(issue_node_id "$num")" >/dev/null
      fi
    fi
  done
  local pending=$(( $(jq '.issues | length' "$MANIFEST") - $(cut -f1 "$KEYS" | sort -u | wc -l) ))
  if (( pending > 0 )); then say "  ($pending manifest issues not created yet; they are added on the apply that creates them)"; fi

  if [[ $APPLY -eq 1 && -n "${proj_id:-}" ]]; then set_item_fields "$proj_id"; fi

  say ""
  say "  Manual steps (no API for views or built-in workflows), in the project's web UI:"
  say "    - Status options: add \"In review\" and \"Blocked\""
  say "    - Views: Board (by Status, current milestone), Plan (table grouped by milestone, sum Estimate),"
  say "      Approvals (label:needs-approval is:open), Exam (group by label aip:*), Roadmap (milestone dates)"
  say "    - Workflows: auto-add from $REPO with 'is:issue label:recruiter-intelligence'; item added -> Todo;"
  say "      item closed / PR merged -> Done"
}

set_item_fields() { # project-id: fill Estimate, Exam skills, Kind and Status (Todo only when empty)
  local proj_id="$1" fjson items
  fjson="$(gql -f query='query($p:ID!){node(id:$p){...on ProjectV2{fields(first:50){nodes{
            ...on ProjectV2Field{id name dataType}
            ...on ProjectV2SingleSelectField{id name dataType options{id name}}}}}}}' -f p="$proj_id" --jq '.data.node.fields.nodes')"
  items="$(gql -f query='query($p:ID!){node(id:$p){...on ProjectV2{items(first:100){nodes{id
            content{...on Issue{number}}
            fieldValueByName(name:"Status"){...on ProjectV2ItemFieldSingleSelectValue{name}}}}}}}' -f p="$proj_id" \
            --jq '.data.node.items.nodes[] | [(.content.number|tostring), .id, (.fieldValueByName.name // "")] | @tsv')"
  local fid_est fid_sk fid_kind fid_status todo
  fid_est="$(jq -r '.[] | select(.name=="Estimate (h)") | .id' <<<"$fjson")"
  fid_sk="$(jq -r '.[] | select(.name=="Exam skills") | .id' <<<"$fjson")"
  fid_kind="$(jq -r '.[] | select(.name=="Kind") | .id' <<<"$fjson")"
  fid_status="$(jq -r '.[] | select(.name=="Status") | .id' <<<"$fjson")"
  todo="$(jq -r '.[] | select(.name=="Status") | .options[] | select(.name=="Todo") | .id' <<<"$fjson")"

  local num item status meta est sk kind kid
  while IFS=$'\t' read -r num item status; do
    meta="$(jq -c --argjson n "$num" '(.issues[] | select(.number == $n)), (.existing[] | select(.number == $n))' "$MANIFEST" | head -n1)"
    if [[ -z "$meta" ]]; then continue; fi
    est="$(jq -r '.estimate // empty' <<<"$meta")"
    sk="$(jq -r '.skills // empty' <<<"$meta")"
    kind="$(jq -r '.kind // empty' <<<"$meta")"
    if [[ -n "$est" ]]; then
      # gh -F would send 1.5 as a string; build the request so the estimate is a JSON number.
      gql --input <(jq -n --arg p "$proj_id" --arg i "$item" --arg f "$fid_est" --argjson v "$est" \
        '{query: "mutation($p:ID!,$i:ID!,$f:ID!,$v:Float!){updateProjectV2ItemFieldValue(input:{projectId:$p,itemId:$i,fieldId:$f,value:{number:$v}}){projectV2Item{id}}}", variables: {p: $p, i: $i, f: $f, v: $v}}') >/dev/null
    fi
    if [[ -n "$sk" ]]; then
      gql -f query='mutation($p:ID!,$i:ID!,$f:ID!,$v:String!){updateProjectV2ItemFieldValue(input:{projectId:$p,itemId:$i,fieldId:$f,value:{text:$v}}){projectV2Item{id}}}' \
      -f p="$proj_id" -f i="$item" -f f="$fid_sk" -f v="$sk" >/dev/null
    fi
    if [[ -n "$kind" ]]; then
      kid="$(jq -r --arg k "$kind" '.[] | select(.name=="Kind") | .options[] | select(.name==$k) | .id' <<<"$fjson")"
      gql -f query='mutation($p:ID!,$i:ID!,$f:ID!,$v:String!){updateProjectV2ItemFieldValue(input:{projectId:$p,itemId:$i,fieldId:$f,value:{singleSelectOptionId:$v}}){projectV2Item{id}}}' \
        -f p="$proj_id" -f i="$item" -f f="$fid_kind" -f v="$kid" >/dev/null
    fi
    if [[ -z "$status" && -n "$todo" ]]; then
      gql -f query='mutation($p:ID!,$i:ID!,$f:ID!,$v:String!){updateProjectV2ItemFieldValue(input:{projectId:$p,itemId:$i,fieldId:$f,value:{singleSelectOptionId:$v}}){projectV2Item{id}}}' \
        -f p="$proj_id" -f i="$item" -f f="$fid_status" -f v="$todo" >/dev/null
    fi
    say "  DO    fields on #$num"
  done <<<"$items"
}

# --- main --------------------------------------------------------------------

if [[ $APPLY -eq 1 ]]; then say "APPLY: changes will be made to $REPO"; else say "DRY RUN for $REPO (pass --apply to perform)"; fi
load_titles
resolve_keys
for phase in labels milestones issues links existing project; do
  if want_phase "$phase"; then "phase_$phase"; fi
done
say "done"
