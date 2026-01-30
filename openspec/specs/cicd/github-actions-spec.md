# CI/CD Deployment Specification

## Overview

Extend GitHub Actions workflow to automatically deploy the homyak application to Kubernetes using Helm chart when code is pushed to the master branch.

## Requirements

### FR-1: Workflow Trigger
**Given** a new GitHub Actions workflow for deployment
**When** code is pushed to the master branch
**Then** the deployment workflow is triggered automatically

### FR-2: Kubernetes Authentication
**Given** the deployment workflow runs
**When** authenticating to the Kubernetes cluster
**Then** it uses the KUBECONFIG_BASE64 secret to configure kubectl
**And** the kubeconfig is decoded from base64 and written to a file
**And** kubectl can successfully connect to the cluster

### FR-3: Helm Upgrade
**Given** the deployment workflow runs
**When** upgrading the Helm release
**Then** it runs `helm upgrade --install homyak ./helm --namespace homyak`
**And** the upgrade uses the latest Docker image tag from the build
**And** the upgrade waits for the rollout to complete

### FR-4: Rollout Verification
**Given** the Helm upgrade completes
**When** verifying the deployment
**Then** it runs `kubectl rollout status deployment/homyak --namespace homyak`
**And** the command exits with success (all pods ready)

### FR-5: Error Handling
**Given** the deployment workflow encounters an error
**When** a failure occurs during deployment
**Then** the workflow fails immediately
**And** GitHub Actions status shows failure
**And** notifications are sent (if configured)

### FR-6: Image Tag Management
**Given** the deployment workflow runs
**When** deploying a new version
**Then** it uses the Docker image tag from the build workflow
**And** the tag is passed to Helm via `--set image.tag=<tag>`

### FR-7: Namespace Management
**Given** the deployment workflow runs
**When** deploying to the cluster
**Then** it creates the homyak namespace if it doesn't exist
**And** it uses the `--create-namespace` flag with Helm

### FR-8: Deployment Notification
**Given** the deployment workflow completes
**When** the deployment is successful
**Then** the workflow status shows success
**And** a summary is displayed with deployment details

## Technical Details

### Workflow File: `.github/workflows/deploy.yml`

```yaml
name: deploy
on:
  push:
    branches:
      - master
  workflow_dispatch:  # Allow manual trigger

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
  HELM_CHART_PATH: ./helm
  NAMESPACE: homyak
  RELEASE_NAME: homyak

jobs:
  deploy:
    runs-on: ubuntu-latest
    timeout-minutes: 10

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Set up kubectl
        uses: azure/setup-kubectl@v4
        with:
          version: latest

      - name: Configure kubectl
        run: |
          mkdir -p $HOME/.kube
          echo "${{ secrets.KUBECONFIG_BASE64 }}" | base64 -d > $HOME/.kube/config
          chmod 600 $HOME/.kube/config

      - name: Verify cluster connection
        run: kubectl cluster-info

      - name: Set up Helm
        uses: azure/setup-helm@v4
        with:
          version: latest

      - name: Get image tag
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=sha

      - name: Create namespace
        run: |
          kubectl create namespace ${{ env.NAMESPACE }} --dry-run=client -o yaml | kubectl apply -f -

      - name: Deploy with Helm
        run: |
          helm upgrade --install ${{ env.RELEASE_NAME }} ${{ env.HELM_CHART_PATH }} \
            --namespace ${{ env.NAMESPACE }} \
            --create-namespace \
            --set image.tag=${{ steps.meta.outputs.version }} \
            --wait \
            --timeout 5m

      - name: Verify deployment
        run: |
          kubectl rollout status deployment/${{ env.RELEASE_NAME }} --namespace ${{ env.NAMESPACE }}
          kubectl get pods -n ${{ env.NAMESPACE }} -l app.kubernetes.io/name=${{ env.RELEASE_NAME }}

      - name: Get deployment status
        if: always()
        run: |
          echo "### Deployment Status 🚀" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "```" >> $GITHUB_STEP_SUMMARY
          kubectl get all -n ${{ env.NAMESPACE }} >> $GITHUB_STEP_SUMMARY
          echo "```" >> $GITHUB_STEP_SUMMARY
```

### Required GitHub Secrets

| Secret Name | Description | How to Generate |
|-------------|-------------|----------------|
| `KUBECONFIG_BASE64` | Base64-encoded kubeconfig file | `cat ~/.kube/config \| base64 -w 0` |

### Kubeconfig Requirements

The kubeconfig file must:
- Have admin access to the cluster
- Be able to create namespaces, deployments, services, ingress
- Use a service account with appropriate RBAC
- Be valid and not expired

### Workflow Dependencies

The deployment workflow depends on:
1. **Build workflow** (`.github/workflows/build.yml`) - Must complete successfully
2. **Docker registry** - Image must be available in ghcr.io
3. **Kubernetes cluster** - Must be accessible and running
4. **cert-manager** - Must be installed in cluster
5. **Ingress controller** - Must be running (assumed nginx)

### Deployment Sequence

```
Push to master branch
         │
         ▼
┌─────────────────┐
│ Build workflow  │ (already exists)
│ - Build Docker  │
│ - Push image    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Deploy workflow │ (this change)
│ - Checkout      │
│ - Setup kubectl │
│ - Setup Helm    │
│ - Deploy chart  │
│ - Verify        │
└─────────────────┘
```

### Rollback Procedure

**Manual Rollback via GitHub Actions**:
1. Go to Actions tab
2. Select failed deployment run
3. Re-run previous workflow (or rollback commit)

**Manual Rollback via Helm**:
```bash
# View release history
helm history homyak --namespace homyak

# Rollback to previous version
helm rollback homyak --namespace homyak
```

**Git-based Rollback**:
```bash
# Revert commit
git revert <commit-hash>
git push origin master
```

### Testing the Workflow

**Test in development**:
1. Create test namespace: `kubectl create namespace homyak-test`
2. Modify workflow to use test namespace
3. Run workflow manually via `workflow_dispatch`
4. Verify deployment
5. Clean up: `helm uninstall homyak -n homyak-test`

**Test dry-run**:
```bash
# Test helm template locally
helm template --dry-run homyak ./helm --namespace homyak

# Test deployment with dry-run
helm upgrade --install --dry-run --debug homyak ./helm --namespace homyak
```

## Acceptance Criteria

- [ ] Workflow triggers on push to master branch
- [ ] KUBECONFIG_BASE64 secret is configured in GitHub
- [ ] `kubectl cluster-info` succeeds in workflow
- [ ] Helm upgrade completes without errors
- [ ] `kubectl rollout status` shows all pods ready
- [ ] Application is accessible at https://victorzh.uk
- [ ] TLS certificate is valid
- [ ] Workflow fails appropriately on errors
- [ ] Deployment summary is displayed in GitHub Actions UI

## Security Considerations

- **Secrets**: KUBECONFIG_BASE64 must be stored as GitHub secret (never in code)
- **RBAC**: Service account in kubeconfig should have least privilege
- **Secret Rotation**: Rotate kubeconfig regularly (e.g., every 90 days)
- **Audit Logs**: Enable Kubernetes audit logging for deployment activities
- **Network Policies**: Consider adding network policies to restrict pod communication

## Dependencies

- Kubernetes cluster (kubespray-managed)
- GitHub repository with Actions enabled
- Docker image built and pushed to ghcr.io
- Helm 3.x installed in workflow
- kubectl configured in workflow
- KUBECONFIG_BASE64 secret in GitHub

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| `Unable to connect to the server` | Invalid KUBECONFIG | Verify secret is correct, check kubeconfig |
| `ImagePullBackOff` | Image not available | Ensure build workflow completed, check image tag |
| `Certificate failed` | cert-manager not ready | Verify cert-manager is running, check ClusterIssuer |
| `502 Bad Gateway` | Pods not ready | Check pod logs, verify health probes |
| `Helm release already exists` | Previous deployment | Use `helm upgrade --install` instead of `helm install` |

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Helm Best Practices for CI/CD](https://helm.sh/docs/howto/charts_tips_and_tricks/)
- [Kubernetes Deployment Strategies](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [GitHub Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
