# Design Document: Helm Deployment Architecture

## Architecture Decisions

### ADR-001: Helm Chart Structure

**Status**: Accepted

**Context**: Need to package homyak for Kubernetes deployment with best practices.

**Options Considered**:

| Option | Structure | Pros | Cons |
|--------|-----------|------|------|
| A | Single deployment.yaml | Simple | Not reusable, no upgrade path |
| B | **Full Helm chart** | **Reusable, templated, upgrades** | More files |
| C | Kustomize overlays | GitOps-friendly | No dependency management |

**Decision**: Use full Helm chart with templates

**Rationale**:
- Helm is industry standard for Kubernetes packaging
- Built-in upgrade/rollback mechanisms
- Template system for environment variations
- Easy to install/upgrade with single command
- Existing ecosystem of Helm plugins/tools

**Consequences**:
- Need to maintain Helm chart versioning
- Requires Helm 3.x in CI/CD
- Learning curve for team unfamiliar with Helm

---

### ADR-002: Deployment Strategy

**Status**: Accepted

**Context**: Choose rolling update strategy for zero-downtime deployments.

**Options Considered**:

| Option | Strategy | Pros | Cons |
|--------|----------|------|------|
| A | Recreate | Simple | Full downtime during deploy |
| B | **Rolling Update** | **Zero downtime, gradual rollout** | Slower deploy |
| C | Blue/Green | Instant rollback | Double resources needed |

**Decision**: Rolling Update with 2 replicas

**Rationale**:
- Stateless application (no session state)
- Zero downtime requirement
- Resource efficiency (no double allocation)
- Kubernetes native
- Gradual rollout allows early failure detection

**Configuration**:
```yaml
replicas: 2
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0
```

**Consequences**:
- Must ensure backward compatibility during updates
- Health checks must be fast and reliable
- Database schema changes need separate migration strategy

---

### ADR-003: Ingress and TLS Configuration

**Status**: Accepted

**Context**: Secure external access via HTTPS with automatic certificate management.

**Options Considered**:

| Option | TLS Approach | Pros | Cons |
|--------|--------------|------|------|
| A | Manual TLS Secret | Full control | Manual renewal, operational burden |
| B | **cert-manager + Let's Encrypt** | **Automatic, secure** | Requires issuer setup |
| C | External load balancer | Offloaded | Additional cost/complexity |

**Decision**: cert-manager with Let's Encrypt via Ingress

**Rationale**:
- Automatic certificate provisioning and renewal
- Industry standard solution
- Free certificates (Let's Encrypt)
- No operational burden
- Kubernetes native integration

**Ingress Configuration**:
```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: victorzh.uk
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: homyak-tls
      hosts:
        - victorzh.uk
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
``**Prerequisites**:
- cert-manager installed in cluster
- `letsencrypt-prod` ClusterIssuer configured
- DNS A record pointing to cluster ingress controller

**Consequences**:
- ClusterIssuer must be set up before first deployment
- Rate limits on Let's Encrypt (50 certs/week)
- Certificate propagation may take up to 60 seconds

---

### ADR-004: Configuration Management

**Status**: Accepted

**Context**: Stateless application requires configuration via environment variables.

**Options Considered**:

| Option | Config Approach | Pros | Cons |
|--------|-----------------|------|------|
| A | Mounted config files | Complex configs | Needs app changes |
| B | **Environment variables** | **Simple, universal** | Secrets visible in pod spec |
| C | Etcd/Consul | Dynamic updates | Operational complexity |

**Decision**: Environment variables via ConfigMap + Secret

**Rationale**:
- Go application already uses env vars (cleanenv)
- No code changes required
- Simple and well-understood
- Kubernetes native
- Separation of concerns (ConfigMap vs Secret)

**Implementation**:
```yaml
# ConfigMap for non-sensitive
apiVersion: v1
kind: ConfigMap
data:
  LOG_LEVEL: info
  SERVER_PORT: "8080"

# Secret for sensitive
apiVersion: v1
kind: Secret
type: Opaque
stringData:
  DATABASE_URL: postgres://...
```

**Consequences**:
- Secrets base64-encoded (not encrypted at rest by default)
- Need proper RBAC on Secrets
- Pod restart required for config changes
- CI/CD must handle sensitive values securely

**Security Considerations**:
- Use Sealed Secrets or External Secrets Operator for production
- Never commit secrets to git
- Rotate secrets regularly
- Use Kubernetes secret encryption at rest (if available)

---

### ADR-005: GitHub Actions Deployment Strategy

**Status**: Accepted

**Context**: Automate deployment to Kubernetes on git push.

**Options Considered**:

| Option | Trigger | Pros | Cons |
|--------|---------|------|------|
| A | Manual approval | Control | Slower, friction |
| B | **Automated on main** | **Fast, GitOps** | Less control |
| C | Pull request | Preview deployments | Resource usage |

**Decision**: Automated deployment on push to main branch

**Rationale**:
- True GitOps workflow
- Fast feedback loop
- Main branch protected (require PR review)
- Rollback via git revert
- Team is small (can trust main branch)

**Workflow Steps**:
1. Checkout code
2. Setup kubectl with KUBECONFIG_BASE64 secret
3. Build and push Docker image (existing workflow)
4. Wait for image to be available
5. Helm upgrade with new image tag
6. Verify rollout status
7. Notify on failure

**Secrets Required**:
```yaml
- name: KUBECONFIG_BASE64
  description: Base64-encoded kubeconfig file
  required: true
```

**Consequences**:
- Must protect main branch (require review, status checks)
- Failed deployments block main until fixed
- Need proper error notification (Slack/email)
- KUBECONFIG must be kept secret and rotated

**Mitigation**:
- Require PR approval before merging to main
- Add manual approval step for production (optional)
- Monitor deployment status and alert on failures
- Keep recent releases in Helm history for rollback

---

## Technical Approach

### Deployment Architecture

```
┌─────────────────┐
│   GitHub Actions │
│   (CI/CD)        │
└────────┬────────┘
         │
         │ helm upgrade
         ▼
┌─────────────────┐
│   Kubernetes    │
│   Cluster       │
│                 │
│  ┌───────────┐  │
│  │ Ingress   │◄─┼─── cert-manager ──► Let's Encrypt
│  │ (nginx)   │  │
│  └─────┬─────┘  │
│        │        │
│        ▼        │
│  ┌───────────┐  │
│  │ Service   │  │
│  │ (8080)    │  │
│  └─────┬─────┘  │
│        │        │
│        ▼        │
│  ┌───────────┐  │
│  │Deployment │  │
│  │ (2 pods)  │  │
│  └───────────┘  │
└─────────────────┘
```

### Helm Chart Template Structure

```yaml
helm/
├── Chart.yaml              # Chart metadata (version, description)
├── values.yaml              # Default configuration values
├── .helmignore              # Files to exclude from chart
└── templates/
    ├── _helpers.tpl         # Template helpers (optional)
    ├── deployment.yaml      # Deployment resource
    ├── service.yaml         # Service resource
    ├── ingress.yaml         # Ingress resource
    ├── configmap.yaml       # ConfigMap for non-secret config
    ├── secret.yaml          # Secret for sensitive config
    ├── serviceaccount.yaml  # ServiceAccount for pod identity
    └── NOTES.txt            # Post-install instructions
```

### Rollback Strategy

**Helm Native Rollback**:
```bash
# View release history
helm history homyak

# Rollback to previous revision
helm rollback homyak

# Rollback to specific revision
helm rollback homyak 5
```

**Git-Based Rollback**:
```bash
# Revert commit
git revert <commit-hash>

# Push to trigger redeployment
git push origin main
```

**Emergency Rollback**:
```bash
# If Helm fails, use kubectl directly
kubectl rollout undo deployment/homyak
```

### Testing Strategy

| Test Type | Tool | Scope |
|-----------|------|-------|
| Lint | `helm lint` | Chart syntax and best practices |
| Template validation | `helm template --dry-run` | Render manifests without installing |
| Integration | `helm install --dry-run --debug` | Validate against cluster |
| Smoke test | `kubectl rollout status` | Pod health and readiness |
| End-to-end | `curl https://victorzh.uk` | Full application flow |

### Monitoring and Observability

**Health Checks**:
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

**Logging**:
- Structured JSON logs (already configured)
- Log level configurable via ConfigMap
- Centralized logging via cluster log aggregation (if available)

**Metrics** (future enhancement):
- Prometheus metrics endpoint
- Pod resource metrics via Metrics Server
- Alert on deployment failures

## Files Created

| File | Purpose |
|------|---------|
| `helm/Chart.yaml` | Helm chart metadata |
| `helm/values.yaml` | Default configuration values |
| `helm/templates/deployment.yaml` | Deployment resource |
| `helm/templates/service.yaml` | Service resource |
| `helm/templates/ingress.yaml` | Ingress with TLS |
| `helm/templates/configmap.yaml` | Non-sensitive config |
| `helm/templates/secret.yaml` | Sensitive config |
| `helm/templates/serviceaccount.yaml` | ServiceAccount |
| `.github/workflows/deploy.yml` | CI/CD deployment workflow |

## Success Metrics

- **Deployment time**: < 5 minutes from push to production
- **Rollback time**: < 2 minutes to previous version
- **Uptime**: 99.9% during deployments (rolling updates)
- **Certificate validity**: Always valid (auto-renewal)
- **Pod health**: All pods passing health checks

## Future Enhancements

1. **Multi-environment**: Add staging/dev environments
2. **Horizontal Pod Autoscaler**: Scale based on CPU/memory
3. **PodDisruptionBudget**: Ensure availability during node maintenance
4. **Network Policies**: Restrict pod-to-pod communication
5. **Sealed Secrets**: Encrypt secrets in git
6. **Prometheus Operator**: Advanced metrics and alerting
7. **Canary Deployments**: Gradual traffic shifting with Flagger
