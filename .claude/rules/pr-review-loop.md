---
paths:
  - "**/*"
---

## PR review loop — local review first, Greptile once

A PR is done at green CI, a local `pr-reviewer` `VERDICT: CLEAN` on the pushed head, and
every Greptile thread resolved. Not at "opened", and not at a re-triggered Greptile 5/5.
Do not report a PR as finished before that, and do not merge it yourself.

Why: Greptile bills every completed review, and a manual re-trigger is a review ($1 per
credit past the 50 included per seat each month). In September 2026 the Mesh repos had
about 1,020 `@greptile-apps review` re-triggers (mesh-gateway 745, mesh-client 166,
mesh-admin-client 93). The local reviewer does the verify rounds now, and Greptile stays as
one independent pass per PR.

0. **Before the first push**, run the `pre-push-review` skill (the `pr-reviewer` agent in
   FULL mode). Fix every blocking finding and rerun it (VERIFY) until `VERDICT: CLEAN`.
   Put the verdict line and the reviewed SHA in the PR body.
1. **Greptile, once.** It reviews automatically when the PR opens, but not reliably on
   small PRs: if nothing has arrived after 15 minutes, comment `@greptile-apps review` as a
   top-level PR comment (a reply inside a thread does nothing). That is the first review,
   not a re-trigger. Read every inline finding, fix or refute it in code, and reply on the
   thread naming the commit. A description that overclaims is a finding too: fix the
   claim, don't argue.
   - Check the fixes with `pre-push-review` (VERIFY), not with `@greptile-apps review`.
   - Re-trigger Greptile at most once more, and only when a fix round materially changed
     credentials, transport or a public type. Never to move the score to 5/5.
   - Over 15 changed files Greptile skips the PR. Force it with `@greptile-apps review`
     only when the PR touches one of those surfaces; otherwise the local FULL review stands.
   - Docs- or test-only PR: add the `skip-greptile` label before opening it.
2. **CI.** `gh pr checks <n> --watch` in the background and act on the result.
   `gh pr checks` after a push can still show the previous commit; confirm
   `gh pr view --json headRefOid` matches what you pushed.
3. **Report** with one line per PR: number, what, CI state, the local verdict and the SHA
   it covers, open Greptile threads — and what is deliberately left open.
