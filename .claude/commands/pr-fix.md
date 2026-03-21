Fix issues identified in a PR review. Optionally pass a PR number: $ARGUMENTS

If no PR number is provided, detect the PR for the current branch using `gh pr view --json number`.

## Phase 1: Fetch & Triage

1. **Fetch PR metadata** — Run `gh pr view <number> --json title,body,headRefName,headRefOid,files` to get the PR context and changed files.
2. **Fetch all comments** — Run all three in parallel:
   - `gh api repos/sh3r4rd/sh3r4rd.com/pulls/<number>/comments` (inline review comments)
   - `gh api repos/sh3r4rd/sh3r4rd.com/pulls/<number>/reviews` (review bodies)
   - `gh api repos/sh3r4rd/sh3r4rd.com/issues/<number>/comments` (issue-level comments)
3. **Triage each comment** — For every comment, read the referenced code and classify it as:
   - **actionable** — Real bug, logic error, missing edge case, or necessary fix
   - **nitpick** — Subjective preference with no functional impact
   - **false positive** — Incorrect suggestion or based on a misunderstanding
   - **pre-existing** — Valid concern but not introduced by this PR
   - **resolved** — Already addressed in a subsequent commit
4. **Present triage results** — Show a numbered table with columns: `#`, `Source`, `File:Line`, `Category`, `Summary`. Group actionable items first.

**APPROVAL GATE** — Use `AskUserQuestion` to ask: "Enter the numbers of the issues to fix (e.g. '1,2,4'), 'all' for all actionable items, or 'none' to skip."

If the user says "none" or there are no actionable items, report that no fixes are needed and stop.

## Phase 2: Fix Each Issue

For each approved issue, in order:

1. **Read context** — Open the file and understand the surrounding code.
2. **Make the fix** — Edit the code to address the issue. Keep changes minimal and focused.
3. **Run verification** based on which files were changed:
   - **Go files** (`infra/recruiter-dashboard/lambda-src/**`) — Run `go vet ./...` from the relevant module directory. If tests exist, run `go test ./...`.
   - **Frontend files** (`src/**`) — Run `npm run lint`
   - **Terraform files** (`infra/**/*.tf`) — Run `terraform -chdir=infra/recruiter-dashboard validate` and `terraform -chdir=infra/recruiter-dashboard fmt -recursive -check`
   - If verification fails, attempt to fix. If it still fails after one retry, stop and report the error.
4. **Show the diff** — Run `git diff` for the changed files.

**APPROVAL GATE** — Use `AskUserQuestion` to ask: "Diff for issue #N: [one-line summary]. Accept this fix? (yes / no / edit)"
- **yes** — Proceed to the next issue.
- **no** — Revert the changes with `git checkout -- <files>` and move on.
- **edit** — Ask what to change, apply the edit, re-verify, and show the updated diff.

Repeat for each approved issue.

## Phase 3: Commit & Push

1. **Group changes** — If multiple fixes touch the same logical area, combine into one commit. Otherwise, one commit per fix.
2. **Draft commit messages** — Use conventional commit format (e.g. `fix(lambda):`, `fix(frontend):`). Reference the PR in the body if relevant.
3. **Present commit plan** — Show each proposed commit with its message and included files.

**APPROVAL GATE** — Use `AskUserQuestion` to ask: "Commit plan above. Approve, edit a message, or abort? (approve / edit / abort)"
- **approve** — Stage files, commit with the approved messages (include co-author trailer), and `git push`.
- **edit** — Ask which message to change, then re-present.
- **abort** — Leave changes unstaged and stop.

4. **Report summary** — List what was fixed, what was skipped, and the commits pushed.
