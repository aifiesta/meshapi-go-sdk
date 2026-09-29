---
name: pre-push-review
description: >
  Review a branch locally with the pr-reviewer agent before pushing it, and again after
  every round of changes: FULL on the branch's first review, VERIFY on each later round.
  Use before `git push` or `gh pr create`, and instead of `@greptile-apps review` when
  checking fixes to Greptile findings, review comments or CI failures.
tools: Read, Grep, Bash, Agent
---

# Pre-push review

Greptile bills every review, re-triggers included. This runs the same kind of review
locally, so Greptile is needed once per PR.

## 1. Commit, then collect the inputs

The reviewer reads commits, not the working tree, so commit first.

```bash
git fetch -q origin main
REPO=$(git rev-parse --show-toplevel)
BASE=$(git merge-base origin/main HEAD)
HEAD_SHA=$(git rev-parse HEAD)
STATE="$(git rev-parse --git-common-dir)/pr-review/$(git branch --show-current | tr / _).md"
mkdir -p "$(dirname "$STATE")"
```

Also gather the Jira ticket (from the branch name, commits or PR body), the decisions
already settled with the user in this conversation or the PR body, and `siblings` for a
change that spans repos.

## 2. Pick the mode

- **No `$STATE` file: FULL.** It runs once per branch. Never run FULL again on a branch
  that already has one: a second FULL re-reviews everything and brings back the
  flip-flopping this exists to stop.
- **`$STATE` exists: VERIFY.** Pass `reviewed` and `reviewed_base` from its last entry,
  plus every blocking finding from that entry with a resolution: `fixed`,
  `accepted by user`, `left (P3)` or `follow-up <ticket>`. Once the PR is open, add that
  round's feedback as items too: Greptile threads, review comments, CI failures.

## 3. Run it

Dispatch the `pr-reviewer` agent with `repo`, `base`, `head` and the rest. Pass its report
on unedited: the first line is the verdict, and the rest is parsed.

## 4. Record the run

Append the mode, `base`, `head`, the time and the full report to `$STATE`. It lives inside
`.git`, so it is never committed.

## 5. Act on the verdict

- `VERDICT: CLEAN`: push. On the first push, put the verdict line and the head SHA in the
  PR body.
- `VERDICT: <n> BLOCKING`: fix, commit, and go back to step 1.
- Anything under `Questions for the user:` goes to the user before you push.
