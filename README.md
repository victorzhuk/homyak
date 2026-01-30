# Homyak

Homyak is a web application built with Go and deployed to Kubernetes via Helm.

## Quick Start

### Local Development

```bash
# Build the application
make build

# Run the application
./bin/homyak

# Run tests
make test
```

### Deployment to Kubernetes

Prerequisites:

- Kubernetes cluster (kubespray-managed)
- kubectl configured
- Helm 3.x installed
- GitHub Container Registry access

```bash
# Deploy to production
helm install homyak ./helm --namespace homyak --create-namespace

# Deploy with custom values
helm install homyak ./helm --namespace homyak \
  --set image.tag=latest \
  --set replicaCount=3
```

## Documentation

### Deployment Documentation

- **[Helm Chart README](helm/README.md)** - Complete guide for deploying with Helm
- **[Operations Guide](docs/operations.md)** - Daily operations, troubleshooting, maintenance
- **[GitHub Secrets](docs/secrets.md)** - Required secrets and setup instructions
- **[cert-manager Setup](docs/cert-manager.md)** - TLS certificate configuration

### Architecture

The application follows Clean Architecture principles:

```
cmd/homyak/          - Application entry point
internal/
  domain/            - Business logic and entities
  repository/         - Data access layer
  usecase/            - Application use cases
  transport/           - HTTP handlers and middleware
```

## CI/CD

### GitHub Actions

The project uses GitHub Actions for:

1. **Build** - Docker image builds on every push
2. **Lint** - Code quality checks with golangci-lint
3. **Deploy** - Automatic deployment to Kubernetes on push to `main` branch

### Build Workflow

- Triggers: Push to any branch
- Builds Docker image with commit SHA tag
- Pushes to GitHub Container Registry: `ghcr.io/victorzhuk/homyak`

### Deploy Workflow

- Triggers: Push to `main` branch, manual workflow_dispatch
- Deploys using Helm chart
- Configures Kubernetes via KUBECONFIG_BASE64 secret
- Deploys to `homyak` namespace

## Configuration

Application is configured via environment variables with `APP_` prefix:

| Variable                      | Description                                  | Default                 |
|-------------------------------|----------------------------------------------|-------------------------|
| `APP_DEBUG`                   | Enable debug logging                         | `false`                 |
| `APP_ENV`                     | Environment (local, development, production) | `local`                 |
| `APP_HTTP_ADDR`               | HTTP server address                          | `:8080`                 |
| `APP_HTTP_MAX_HEADER_SIZE_MB` | Max header size in MB                        | `1`                     |
| `APP_HTTP_READ_TIMEOUT`       | HTTP read timeout                            | `3s`                    |
| `APP_HTTP_WRITE_TIMEOUT`      | HTTP write timeout                           | `3s`                    |
| `APP_FEEDBACK_FORM_URL`       | Feedback form redirect URL                   | `http://localhost:8080` |

Environment variables are loaded directly into the application (no ConfigMap/Secret required for this stateless app).

## Development

### Prerequisites

- Go 1.25+
- Docker
- Make

### Commands

```bash
# Build
make build

# Run
make run

# Test
make test

# Lint
make lint
```

## Production Deployment

### Manual Deployment

1. **Deploy**:
   ```bash
   helm upgrade --install homyak ./helm \
     --namespace homyak \
     --create-namespace \
     --wait \
     --timeout 5m
   ```

3. **Verify**:
   ```bash
   kubectl get all -n homyak
   kubectl logs -n homyak -l app.kubernetes.io/name=homyak
   ```

### Automatic Deployment

Push to `main` branch triggers:

1. Docker image build and push
2. Kubernetes deployment via Helm
3. Rollout verification

See [docs/secrets.md](docs/secrets.md) for required GitHub secrets setup.

## Monitoring

### Health Checks

- **Liveness**: `GET /healthz` - Container restart if fails
- **Readiness**: `GET /readyz` - Traffic stops if fails
- **Metrics**: `GET /metrics` - Prometheus metrics

### Kubernetes Resources

```bash
# Check pod status
kubectl get pods -n homyak

# View logs
kubectl logs -n homyak -l app.kubernetes.io/name=homyak --tail=100 -f

# Check resource usage
kubectl top pods -n homyak
```

## Troubleshooting

### Pod Issues

```bash
# Describe pod for events
kubectl describe pod <pod-name> -n homyak

# View pod logs
kubectl logs <pod-name> -n homyak
```

### Deployment Issues

```bash
# Check rollout status
kubectl rollout status deployment homyak -n homyak

# View history
helm history homyak -n homyak

# Rollback
helm rollback homyak -n homyak
```

### TLS Certificate Issues

```bash
# Check certificate
kubectl get certificate -n homyak

# Check cert-manager logs
kubectl logs -n cert-manager -l app.kubernetes.io/name=cert-manager
```

See [docs/operations.md](docs/operations.md) for detailed troubleshooting.

## Security

- Runs as non-root user (UID 1000)
- Read-only root filesystem
- No privilege escalation
- All capabilities dropped
- Service account token not mounted
- TLS via Let's Encrypt (cert-manager)

## License

[MIT License](LICENSE)

## Contact

- GitHub: https://github.com/victorzhuk/homyak
- Website: https://victorzh.uk

## Contributing

Contributions welcome! Please read the contribution guidelines before submitting PRs.
