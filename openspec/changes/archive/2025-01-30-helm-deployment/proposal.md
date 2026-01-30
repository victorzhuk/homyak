# Proposal: Helm Chart Deployment for Kubernetes

## Problem Statement

The homyak application currently builds Docker images but lacks automated deployment to Kubernetes. Manual deployment is error-prone and time-consuming. We need:

1. **Infrastructure as Code**: Helm chart for repeatable deployments
2. **CI/CD Integration**: Automated deployment on push to main
3. **TLS Termination**: Secure HTTPS via cert-manager
4. **Configuration Management**: Environment-based configuration via ConfigMaps/Secrets

## Goals

1. **Helm Chart**: Create production-ready Helm chart for homyak
2. **GitHub Actions**: Extend workflow to deploy to Kubernetes cluster
3. **TLS Support**: Configure cert-manager for victorzh.uk domain
4. **Stateless Design**: Use ConfigMaps/Secrets for all configuration
5. **Namespace Isolation**: Deploy to dedicated `homyak` namespace

## Non-Goals

- Database migration automation (handled separately)
- Multi-environment deployments (only production initially)
- Helm chart repository hosting (use chart directly from repo)
- Kubernetes cluster setup (cluster already exists via kubespray)

## Success Criteria

- [ ] Helm chart installs successfully: `helm install homyak ./helm`
- [ ] GitHub Actions deploys on push to main branch
- [ ] Application accessible at https://victorzh.uk with valid TLS
- [ ] All configuration via environment variables (no hardcoded values)
- [ ] Pods start healthy and respond to health checks
- [ ] Deployment can be rolled back via `helm rollback`

## Proposed Changes

### 1. Helm Chart Structure

```
helm/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   └── serviceaccount.yaml
└── README.md
```

### 2. GitHub Actions Workflow Extension

**New workflow**: `.github/workflows/deploy.yml`

**Trigger**: Push to `main` branch

**Steps**:
1. Checkout repository
2. Configure kubectl with `KUBECONFIG_BASE64` secret
3. Upgrade Helm chart: `helm upgrade --install homyak ./helm`
4. Verify deployment: `kubectl rollout status deployment/homyak`

### 3. Kubernetes Resources

| Resource | Purpose |
|----------|---------|
| **Deployment** | 2 replicas, rolling updates |
| **Service** | ClusterIP, port 8080 |
| **Ingress** | TLS via cert-manager, victorzh.uk host |
| **ConfigMap** | Non-sensitive config (log level, etc.) |
| **Secret** | Sensitive config (DB connection, API keys) |
| **ServiceAccount** | Dedicated service account for homyak |

### 4. Configuration Strategy

**Stateless by Design**: All configuration via environment variables

**Configuration Sources**:
- `values.yaml`: Default values for dev/testing
- `--set-file`: Override in CI/CD for production
- Kubernetes Secrets: Encrypted at rest

**Example config**:
```yaml
# values.yaml
replicaCount: 2
image:
  repository: ghcr.io/victorzhuk/homyak
  pullPolicy: IfNotPresent
env:
  logLevel: info
  port: "8080"
ingress:
  enabled: true
  host: victorzh.uk
  tls: true
  certManager: true
```

### 5. TLS Configuration

**cert-manager** will:
- Watch Ingress resource for `cert-manager.io/cluster-issuer` annotation
- Automatically provision Let's Encrypt certificate
- Store certificate in TLS Secret

**Issuer**: `letsencrypt-prod` (assumed to exist in cluster)

## Effort Estimate

| Phase | Description | Estimated Effort |
|-------|-------------|------------------|
| 1 | Helm chart development | 3 hours |
| 2 | GitHub Actions workflow | 2 hours |
| 3 | Testing in cluster | 2 hours |
| 4 | Documentation | 1 hour |
| **Total** | | **~8 hours** |

## Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| cert-manager not configured | Medium | High | Verify issuer exists before deployment |
| Wrong KUBECONFIG in secret | Low | High | Validate kubectl connection in CI |
| Port conflicts | Low | Medium | Check existing Services in namespace |
| Image pull errors | Low | Medium | Use imagePullSecrets for private registry |

## Rollback Plan

1. **Helm rollback**: `helm rollback homyak <revision>`
2. **Git revert**: Revert commit and re-run workflow
3. **Manual intervention**: `kubectl delete -f manifests/` if Helm fails

## References

- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Kubernetes Deployment Strategies](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
