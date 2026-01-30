# Kubernetes Deployment Specification

## Overview

Define Kubernetes resource requirements, configuration, and operational parameters for the homyak application deployment using Helm chart.

## Requirements

### FR-1: Namespace Isolation
**Given** the Kubernetes cluster
**When** deploying homyak
**Then** all resources are deployed to the `homyak` namespace
**And** the namespace is created if it doesn't exist

### FR-2: Pod Replicas
**Given** the deployment is configured
**When** the application is running
**Then** 2 replica pods are running by default
**And** rolling updates ensure zero downtime

### FR-3: Resource Limits
**Given** the deployment pod specifications
**When** pods are scheduled
**Then** each pod has:
- Memory request: 128Mi
- Memory limit: 256Mi
- CPU request: 100m (0.1 CPU)
- CPU limit: 200m (0.2 CPU)

### FR-4: Container Security
**Given** deployment pod specifications
**When** containers are running
**Then** the container:
- Runs as non-root user (uid 1000)
- Has read-only root filesystem
- Has allowPrivilegeEscalation: false
- Drops all Linux capabilities
- Does not mount service account token

### FR-5: Health Probes
**Given** the deployment pod specifications
**When** pods are running
**Then** the container has:
- Liveness probe at /healthz on port 8080
- Readiness probe at /readyz on port 8080
- Initial delay: 10s (liveness), 5s (readiness)
- Period: 10s (liveness), 5s (readiness)
- Timeout: 5s (liveness), 3s (readiness)

### FR-6: Service Configuration
**Given** the Kubernetes service
**When** the application is deployed
**Then** the service:
- Type: ClusterIP
- Port: 8080
- Protocol: TCP
- Selects pods with label `app.kubernetes.io/name: homyak`
- Has no external access (handled by Ingress)

### FR-7: Ingress Configuration
**Given** the Kubernetes ingress
**When** TLS is enabled
**Then** the ingress:
- Host: victorzh.uk
- Path: / (Prefix type)
- Routes traffic to homyak service on port 8080
- Has TLS secret: homyak-tls
- Has cert-manager annotation: cert-manager.io/cluster-issuer: letsencrypt-prod

### FR-8: Certificate Management
**Given** cert-manager is installed
**When** the ingress is created
**Then** cert-manager:
- Automatically provisions Let's Encrypt certificate
- Stores certificate in homyak-tls Secret
- Renews certificate before expiration (30 days before)

### FR-9: Application Configuration
**Given** the application configuration
**When** the deployment is created
**Then** the container has:
- APP_DEBUG: false
- APP_ENV: production
- APP_FEEDBACK_FORM_URL: "https://victorzh.uk"
- APP_HTTP_ADDR: ":8080"
- APP_HTTP_MAX_HEADER_SIZE_MB: "1"
- APP_HTTP_READ_TIMEOUT: "3s"
- APP_HTTP_WRITE_TIMEOUT: "3s"

## Technical Details

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: homyak
  labels:
    name: homyak
    app.kubernetes.io/name: homyak
    app.kubernetes.io/instance: homyak
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: homyak
  namespace: homyak
  labels:
    app.kubernetes.io/name: homyak
    app.kubernetes.io/instance: homyak
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: homyak
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app.kubernetes.io/name: homyak
        app.kubernetes.io/instance: homyak
    spec:
      serviceAccountName: homyak
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      automountServiceAccountToken: false
      containers:
        - name: homyak
          image: ghcr.io/victorzhuk/homyak:latest
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          securityContext:
            runAsNonRoot: true
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
          env:
          - name: APP_DEBUG
            value: "false"
          - name: APP_ENV
            value: "production"
          - name: APP_FEEDBACK_FORM_URL
            value: "https://victorzh.uk"
          - name: APP_HTTP_ADDR
            value: ":8080"
          - name: APP_HTTP_MAX_HEADER_SIZE_MB
            value: "1"
          - name: APP_HTTP_READ_TIMEOUT
            value: "3s"
          - name: APP_HTTP_WRITE_TIMEOUT
            value: "3s"
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
            successThreshold: 1
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
            successThreshold: 1
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 200m
              memory: 256Mi
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: homyak
  namespace: homyak
  labels:
    app.kubernetes.io/name: homyak
    app.kubernetes.io/instance: homyak
spec:
  type: ClusterIP
  ports:
    - port: 8080
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app.kubernetes.io/name: homyak
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: homyak
  namespace: homyak
  labels:
    app.kubernetes.io/name: homyak
    app.kubernetes.io/instance: homyak
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - victorzh.uk
      secretName: homyak-tls
  rules:
    - host: victorzh.uk
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: homyak
                port:
                  number: 8080
```

### ServiceAccount

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: homyak
  namespace: homyak
  labels:
    app.kubernetes.io/name: homyak
    app.kubernetes.io/instance: homyak
  annotations:
    eks.amazonaws.com/role-arn: <optional-iam-role>  # If using EKS
automountServiceAccountToken: false
```

## Operational Procedures

### Deployment

```bash
# Using Helm
helm upgrade --install homyak ./helm --namespace homyak --create-namespace

# Using kubectl (from template output)
kubectl apply -f manifests/
```

### Rollback

```bash
# Using Helm
helm rollback homyak --namespace homyak

# Using kubectl
kubectl rollout undo deployment/homyak --namespace homyak
```

### Scaling

```bash
# Scale up to 4 replicas
kubectl scale deployment/homyak --replicas=4 --namespace homyak

# Using Helm
helm upgrade homyak ./helm --set replicaCount=4 --namespace homyak
```

### Viewing Logs

```bash
# View all pod logs
kubectl logs -l app.kubernetes.io/name=homyak --namespace homyak --all-containers=true --tail=100

# View specific pod logs
kubectl logs -f deployment/homyak --namespace homyak

# View previous pod logs (after crash)
kubectl logs -l app.kubernetes.io/name=homyak --namespace homyak --previous
```

### Debugging

```bash
# Get pod details
kubectl describe pod -l app.kubernetes.io/name=homyak --namespace homyak

# Get events
kubectl get events --namespace homyak --sort-by='.lastTimestamp'

# Exec into pod (if shell available)
kubectl exec -it <pod-name> --namespace homyak -- sh

# Port forward for local testing
kubectl port-forward deployment/homyak 8080:8080 --namespace homyak
```

### Health Checks

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=homyak --namespace homyak

# Check deployment status
kubectl rollout status deployment/homyak --namespace homyak

# Describe deployment
kubectl describe deployment/homyak --namespace homyak
```

## Monitoring

### Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| Pod CPU Usage | CPU consumption per pod | > 180m (90% of limit) |
| Pod Memory Usage | Memory consumption per pod | > 230Mi (90% of limit) |
| Pod Restart Count | Number of pod restarts | > 0 in 5 minutes |
| 5xx Response Rate | HTTP 500+ errors | > 5% of requests |
| Response Time | Average response time | > 1 second |
| Pod Ready Time | Time to become ready | > 60 seconds |

### Logging Strategy

- **Structured Logs**: JSON format with correlation IDs
- **Log Level**: Configurable via environment variables
- **Log Aggregation**: Use cluster log aggregation (if available)
- **Retention**: Follow cluster retention policy (default 7 days)

## Acceptance Criteria

- [ ] Namespace `homyak` is created
- [ ] 2 replica pods are running and ready
- [ ] Pods have correct resource requests and limits
- [ ] Pods run as non-root user with read-only filesystem
- [ ] Health probes are configured and passing
- [ ] Service is accessible on port 8080
- [ ] Ingress routes traffic from victorzh.uk
- [ ] TLS certificate is valid and provisioned by cert-manager
- [ ] Application config is set via environment variables
- [ ] Application responds at https://victorzh.uk

## Dependencies

- Kubernetes cluster (kubespray-managed)
- Ingress controller (nginx)
- cert-manager
- DNS A record: victorzh.uk → cluster ingress IP
- Docker image: ghcr.io/victorzhuk/homyak
- Application: /healthz and /readyz endpoints

## Prerequisites for Cluster

Before deploying, ensure:

1. **Ingress Controller**: nginx-ingress-controller is running
2. **cert-manager**: Installed and working
3. **ClusterIssuer**: `letsencrypt-prod` exists
4. **RBAC**: Service account has appropriate permissions
5. **DNS**: victorzh.uk points to cluster ingress IP
6. **Storage**: If persistent storage is needed (not for this deployment)

## References

- [Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- [cert-manager](https://cert-manager.io/docs/)
- [Kubernetes Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
