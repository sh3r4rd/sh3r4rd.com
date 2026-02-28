Run a pre-deployment checklist for this project. Execute each step and report the results:

1. **Lint** — Run `npm run lint` and report any errors or warnings
2. **Build** — Run `npm run build` and confirm it succeeds without errors
3. **Git status** — Run `git status` and flag any uncommitted changes or untracked files
4. **Branch check** — Confirm the current branch is `main` (deployment only triggers on push to `main`)
5. **Stale boilerplate** — Check that `src/App.css` and `src/assets/react.svg` are not imported anywhere in the codebase

Report a summary with pass/fail for each step. If any step fails, explain what needs to be fixed before deploying.
