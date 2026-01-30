## Why

The GitHub Actions deploy workflow is being skipped because it checks for branch `main` while the repository uses `master` as default branch. This prevents automatic deployments to production.

## What Changes

- Fix deploy workflow condition to check for `master` branch instead of `main`
- Single line change in `.github/workflows/deploy.yml:16`

## Capabilities

### New Capabilities

None - this is a bug fix to existing CI/CD configuration.

### Modified Capabilities

None - no spec-level behavior changes, only a configuration fix.

## Impact

- **File**: `.github/workflows/deploy.yml`
- **Effect**: Deploy workflow will trigger correctly after successful builds on `master` branch
- **Risk**: Low - minimal change, easily verifiable
