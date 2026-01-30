# GitHub Environments Setup

This document describes how to configure GitHub Environments for the homyak deployment.

## Overview

We use GitHub Environments to manage deployment configurations and secrets separately from the workflow files. This provides:

- **Environment-specific variables** (domain, namespace, etc.)
- **Environment-specific secrets** (kubeconfig)
- **Deployment protection rules** (optional approvals, wait timers)
- **Audit logging** of all deployments to each environment

## Current Environments

| Environment | Purpose | Branch |
|-------------|---------|--------|
| `production` | Live production deployment | `main` |

## Required Configuration

### 1. Create the Production Environment

1. Go to: https://github.com/victorzhuk/homyak/settings/environments
2. Click: **New environment**
3. Name: `production`
4. Click: **Configure environment**

### 2. Configure Environment Variables

Add these **Variables** (not secrets - these are visible in logs):

| Name | Value | Description |
|------|-------|-------------|
| `HELM_RELEASE` | `homyak` | Helm release name |
| `HELM_NAMESPACE` | `homyak` | Kubernetes namespace |
| `DOMAIN` | `victorzh.uk` | Domain for ingress |

**How to add:**
1. In the environment settings, find **Environment variables**
2. Click **Add variable**
3. Enter name and value
4. Click **Add variable**

### 3. Configure Environment Secrets

Add this **Secret** (encrypted, not visible in logs):

| Name | Value | Description |
|------|-------|-------------|
| `KUBECONFIG_BASE64` | `<base64-encoded-kubeconfig>` | Kubernetes cluster access |

**How to generate:**
```bash
# On your VPS with kubectl access:
cat ~/.kube/config | base64 -w0
```

**How to add:**
1. In the environment settings, find **Environment secrets**
2. Click **Add secret**
3. Name: `KUBECONFIG_BASE64`
4. Value: Paste the base64 string from above
5. Click **Add secret**

### 4. Configure Protection Rules (Optional)

For production, you may want to add:

- **Required reviewers**: Require approval before deployment
- **Wait timer**: Add a delay before deployment starts
- **Deployment branches**: Only allow deployment from `main` branch

**Recommended for production:**
1. Check **Required reviewers**
2. Add yourself as a reviewer
3. Check **Deployment branches**
4. Add pattern: `main`

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    GitHub Repository                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Workflow: deploy.yml                        │   │
│  │  ┌─────────────┐      ┌─────────────────────────────┐  │   │
│  │  │  build job  │─────▶│  deploy job                 │  │   │
│  │  │             │      │  environment: production    │  │   │
│  │  └─────────────┘      │                             │  │   │
│  │                       │  Uses:                      │  │   │
│  │                       │  • vars.HELM_NAMESPACE      │  │   │
│  │                       │  • vars.DOMAIN              │  │   │
│  │                       │  • secrets.KUBECONFIG_BASE64│  │   │
│  │                       └─────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         Environment: production                          │   │
│  │  ┌─────────────────┐    ┌─────────────────────────┐    │   │
│  │  │  Variables      │    │  Secrets                │    │   │
│  │  │  • NAMESPACE    │    │  • KUBECONFIG_BASE64    │    │   │
│  │  │  • DOMAIN       │    │                         │    │   │
│  │  │  • HELM_RELEASE │    │                         │    │   │
│  │  └─────────────────┘    └─────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Workflow Changes

The deploy workflow now uses:

```yaml
jobs:
  deploy:
    environment: production  # <-- Pulls vars/secrets from here
    steps:
      - name: Deploy
        run: |
          helm upgrade --install ${{ vars.HELM_RELEASE }} ./helm \
            --namespace ${{ vars.HELM_NAMESPACE }} \
            --set ingress.host=${{ vars.DOMAIN }}
```

## Benefits

1. **Separation of concerns**: Config is separate from code
2. **Security**: Secrets are encrypted and environment-scoped
3. **Flexibility**: Easy to add staging environment later
4. **Audit trail**: GitHub logs all deployments to each environment
5. **Protection**: Can require approvals for production deployments

## Adding Staging Environment (Future)

When ready for staging:

1. Create new environment: `staging`
2. Add variables:
   - `HELM_NAMESPACE`: `homyak-staging`
   - `DOMAIN`: `staging.victorzh.uk`
3. Add secret: `KUBECONFIG_BASE64` (same or different cluster)
4. Update workflow to deploy to staging on PRs

## Troubleshooting

### "Environment not found"
- Verify environment name matches exactly (`production`)
- Check environment is created in repository settings

### "Variable not found"
- Verify variable is added to the **environment**, not repository secrets
- Check variable name matches exactly (case-sensitive)

### "Secret not found"
- Verify secret is added to the **environment**, not repository secrets
- Ensure `environment: production` is set in the job

## References

- [GitHub Environments Documentation](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
- [Environment Variables and Secrets](https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions)
- [Deployment Protection Rules](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#deployment-protection-rules)
