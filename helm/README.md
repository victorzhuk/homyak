# Homyak Helm Chart

Helm chart for deploying homyak application to Kubernetes.

## Prerequisites

- Kubernetes 1.23+
- Helm 3.x
- NGINX Ingress Controller
- cert-manager with ClusterIssuer `letsencrypt-prod`

## Installation

Add the Helm repository (optional, if hosting chart):

```bash
helm repo add homyak https://victorzh.uk/charts
helm repo update
```

Install the chart:

```bash
helm install homyak ./helm
```

Or with custom values:

```bash
helm install homyak ./helm --set image.tag=latest
```

## Configuration

| Parameter                                | Description                | Default                     |
|------------------------------------------|----------------------------|-----------------------------|
| `replicaCount`                           | Number of replicas         | `2`                         |
| `image.repository`                       | Container image repository | `ghcr.io/victorzhuk/homyak` |
| `image.tag`                              | Container image tag        | `Chart.AppVersion`          |
| `image.pullPolicy`                       | Image pull policy          | `IfNotPresent`              |
| `serviceAccount.create`                  | Create service account     | `true`                      |
| `podSecurityContext.runAsNonRoot`        | Run as non-root            | `true`                      |
| `podSecurityContext.runAsUser`           | User ID to run as          | `1000`                      |
| `podSecurityContext.fsGroup`             | Filesystem group ID        | `1000`                      |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem  | `true`                      |
| `service.type`                           | Kubernetes service type    | `ClusterIP`                 |
| `service.port`                           | Service port               | `8080`                      |
| `ingress.enabled`                        | Enable ingress             | `true`                      |
| `ingress.className`                      | Ingress class name         | `nginx`                     |
| `ingress.hosts[0].host`                  | Host name                  | `victorzh.uk`               |
| `ingress.tls[0].secretName`              | TLS secret name            | `homyak-tls`                |
| `resources.requests.cpu`                 | CPU request                | `100m`                      |
| `resources.requests.memory`              | Memory request             | `128Mi`                     |
| `resources.limits.cpu`                   | CPU limit                  | `200m`                      |
| `resources.limits.memory`                | Memory limit               | `256Mi`                     |
| `configMap.logLevel`                     | Log level                  | `info`                      |
| `configMap.port`                         | Server port                | `8080`                      |

### Secrets

The chart creates a Secret with placeholder values. For production, either:

1. **Create secret manually before install**:
   ```bash
   kubectl create secret generic homyak-secret \
     --from-literal=database-url="postgresql://..." \
     --from-literal=api-key="your-api-key"
   ```
   Then set `secret.enabled: false` in values.yaml.

2. **Use Helm --set**:
   ```bash
   helm install homyak ./helm \
     --set-string secret.databaseUrl="postgresql://..." \
     --set-string secret.apiKey="your-api-key"
   ```

## Upgrading

Upgrade the release:

```bash
helm upgrade homyak ./helm
```

Or with new image tag:

```bash
helm upgrade homyak ./helm --set image.tag=v1.2.3
```

## Rollback

List releases:

```bash
helm history homyak
```

Rollback to previous version:

```bash
helm rollback homyak
```

Or to specific revision:

```bash
helm rollback homyak 2
```

## Uninstalling

Remove the release:

```bash
helm uninstall homyak
```

## Troubleshooting

### Check pod status:

```bash
kubectl get pods -l app.kubernetes.io/name=homyak
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### Check events:

```bash
kubectl get events --sort-by='.lastTimestamp'
```

### Debug deployment:

```bash
helm get manifest homyak
helm status homyak
```

### View logs:

```bash
kubectl logs -l app.kubernetes.io/name=homyak --tail=100 -f
```

### TLS Certificate Issues:

Check certificate status:

```bash
kubectl get certificate -A
kubectl describe certificate homyak-tls
```

Check cert-manager logs:

```bash
kubectl logs -n cert-manager -l app.kubernetes.io/name=cert-manager
```

## Resource Requirements

Default resource requests/limits:
- CPU: 100m / 200m
- Memory: 128Mi / 256Mi

Adjust based on workload:

```bash
helm upgrade homyak ./helm \
  --set resources.requests.cpu=200m \
  --set resources.requests.memory=256Mi \
  --set resources.limits.cpu=500m \
  --set resources.limits.memory=512Mi
```

## Security

- Runs as non-root user (UID 1000)
- Read-only root filesystem
- All capabilities dropped
- No privilege escalation
- Service account token not mounted

## Maintenance

Update dependencies:

```bash
helm dependency update
```

Lint the chart:

```bash
helm lint ./helm
```

Test templates:

```bash
helm template --dry-run test-release ./helm
```
