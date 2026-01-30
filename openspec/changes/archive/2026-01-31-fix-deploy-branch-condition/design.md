## Context

The deploy workflow uses `workflow_run` trigger to deploy after successful builds. The condition on line 16 checks `github.event.workflow_run.head_branch == 'main'`, but the repository's default branch is `master`. This causes the job to be skipped.

## Goals / Non-Goals

**Goals:**
- Fix the branch condition so deploys trigger correctly on `master`

**Non-Goals:**
- Renaming the default branch to `main`
- Changing the workflow architecture
- Supporting multiple branch names

## Decisions

**Decision: Change `'main'` to `'master'` in the workflow condition**

Rationale: This is the minimal change that fixes the issue. The repository uses `master` as its default branch, so the condition should match. Alternatives (renaming branch, supporting both) add unnecessary complexity for a straightforward configuration fix.

## Risks / Trade-offs

**Risk: Future confusion if team expects `main`**
→ Mitigation: This is an existing repository with established conventions. The fix aligns with current state.

**Risk: None significant** - Single line change, easily verifiable, low impact.
