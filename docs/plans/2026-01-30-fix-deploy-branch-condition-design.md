# Fix GitHub Actions Deploy Workflow Branch Condition

**Date:** 2026-01-30
**Status:** Approved
**Type:** Bug Fix

## Problem

The deploy workflow is being skipped with status "This job was skipped" because of a branch name mismatch:

- Default branch is `master` (confirmed in GitHub repository settings)
- Build workflow runs successfully on `master` branch
- Deploy workflow condition checks: `github.event.workflow_run.head_branch == 'main'`
- Condition never matches, causing deployment to skip

## Root Cause

The deploy workflow (`.github/workflows/deploy.yml:16`) has a hardcoded check for `main`:

```yaml
if: ${{ github.event.workflow_run.conclusion == 'success' && github.event.workflow_run.head_branch == 'main' }}
```

This was likely copied from a template that assumes GitHub's modern default branch name, but this repository uses `master`.

## Solution

**Approach:** Change the workflow condition to check for `master` instead of `main`.

**File:** `.github/workflows/deploy.yml`
**Line:** 16

**Change:**
```diff
- if: ${{ github.event.workflow_run.conclusion == 'success' && github.event.workflow_run.head_branch == 'main' }}
+ if: ${{ github.event.workflow_run.conclusion == 'success' && github.event.workflow_run.head_branch == 'master' }}
```

## Verification Plan

### Testing Steps

1. Commit the fix to `master` branch
2. Push to GitHub
3. Monitor workflow chain: `golangci-lint` → `build` → `deploy`
4. Verify deploy job runs (not skipped)

### Validation Commands

```bash
# Watch workflow runs
gh run watch

# Check latest deploy run status
gh run list --workflow=deploy --limit 1

# Verify K8s deployment (if kubectl configured)
kubectl get pods -n <namespace>
kubectl rollout status deployment/<helm-release> -n <namespace>
```

### Expected Behavior

```
golangci-lint (on push to master) ✓
    ↓
build (if lint success) ✓
    ↓
deploy (if build success AND branch == master) ✓ [FIXED]
    ↓
K8s production deployment ✓
```

## Additional Notes

### Other Branch References

- Only one reference to `main` exists in workflows (the line being fixed)
- Docker metadata action uses `{{is_default_branch}}` which automatically detects `master` ✓
- No other changes needed

### Required GitHub Configuration

The deploy workflow requires these to be configured:

**Secrets:**
- `KUBECONFIG_BASE64` - Base64-encoded kubeconfig for K8s cluster access

**Variables:**
- `HELM_RELEASE` - Helm release name
- `HELM_NAMESPACE` - K8s namespace for deployment
- `DOMAIN` - Application domain

These should already be configured based on the workflow structure.

## Decision Rationale

**Why this approach:**
- Minimal change (one line)
- Low risk
- Preserves current branch naming convention
- Immediate fix without broader repository changes

**Alternatives considered:**
- Rename default branch to `main` (rejected: more work, broader impact)
- Support both branches in condition (rejected: unnecessary complexity)

## Implementation

Simple one-line change in `.github/workflows/deploy.yml:16`.
