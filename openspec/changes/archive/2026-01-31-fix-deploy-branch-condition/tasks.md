## 1. Fix Workflow Condition

- [x] 1.1 Update `.github/workflows/deploy.yml:16` to check for `master` instead of `main` ✓ (2026-01-31)

## 2. Verification

- [x] 2.1 Push change to `master` branch and verify deploy workflow runs (not skipped) ✓ (2026-01-31)
- [x] 2.2 Confirm deployment completes successfully via `gh run list --workflow=deploy` ✓ (2026-01-31)
