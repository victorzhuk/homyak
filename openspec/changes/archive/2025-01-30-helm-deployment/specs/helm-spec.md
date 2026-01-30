# Helm Chart Specification

## Overview

Create a production-ready Helm chart for deploying the homyak application to Kubernetes with best practices for security, scalability, and maintainability.

## Requirements

### FR-1: Chart Metadata
**Given** a new Helm chart for homyak
**When** the Chart.yaml is created
**Then** it contains:
- API version: v2
- Name: homyak
- Version: 0.1.0
- Description: Helm chart for homyak application
- Type: application

### FR-2: Default Values
**Given** a values.yaml file
**When** the chart is created
**Then** it contains sensible defaults:
- replicaCount: 2
- Image repository: ghcr.io/victorzhuk/homyak
- Image pullPolicy: IfNotPresent
- Service port: 8080
- Resource limits: 256Mi memory, 200m CPU
- Resource requests: 128Mi memory, 100m CPU

### FR-3: Deployment Template
**Given** a deployment.yaml template
**When** rendered with default values
**Then** the resulting deployment:
- Has the specified replica count (default: 2)
- Uses the specified container image
- Sets resource requests and limits
- Configures health probes (liveness and readiness)
- Uses RollingUpdate strategy with maxUnavailable: 0
- Has container security context (runAsNonRoot, readOnlyRootFilesystem)

### FR-4: Service Template
**Given** a service.yaml template
**When** rendered with default values
**Then** the resulting service:
- Type: ClusterIP
- Exposes port 8080 as HTTP
- Selects pods with app: homyak label
- No external access (handled by Ingress)

### FR-5: Ingress Template
**Given** an ingress.yaml template
**When** rendered with TLS enabled
**Then** the resulting ingress:
- Routes traffic for victorzh.uk to the service
- Configures TLS with cert-manager annotation
- Uses pathType: Prefix for path matching
- Has proper ingress class (assumed: nginx)

### FR-6: ServiceAccount Template
**Given** a serviceaccount.yaml template
**When** rendered
**Then** the resulting ServiceAccount:
- Has name: homyak
- Has automountServiceAccountToken: false (security best practice)
- Is used by the deployment pods

### FR-7: Helm Lint
**Given** the completed Helm chart
**When** running `helm lint ./helm`
**Then** the lint passes with no errors or warnings

### FR-8: Template Validation
**Given** the completed Helm chart
**When** running `helm template --dry-run homyak ./helm`
**Then** valid Kubernetes YAML manifests are produced

## Technical Details

### Chart.yaml

```yaml
apiVersion: v2
name: homyak
description: Helm chart for homyak application
type: application
version: 0.1.0
appVersion: "1.0.0"
keywords:
  - go
  - web
  - api
maintainers:
  - name: Victor Zhuk
    email: victor@victorzh.uk
```

### values.yaml

```yaml
# Default values for homyak
replicaCount: 2

image:
  repository: ghcr.io/victorzhuk/homyak
  pullPolicy: IfNotPresent
  tag: ""

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations: {}

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000

securityContext:
  runAsNonRoot: true
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: victorzh.uk
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: homyak-tls
      hosts:
        - victorzh.uk

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi

autoscaling:
  enabled: false
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

# Application configuration via environment variables
config:
  app:
    debug: false
    env: production
    feedbackFormUrl: "https://victorzh.uk"
  http:
    addr: ":8080"
    maxHeaderSizeMb: 1
    readTimeout: "3s"
    writeTimeout: "3s"
```

### Deployment Template

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "homyak.fullname" . }}
  labels:
    {{- include "homyak.labels" . | nindent 4 }}
    spec:
      {{- if not .Values.autoscaling.enabled }}
      replicas: {{ .Values.replicaCount }}
      {{- end }}
      selector:
        matchLabels:
          {{- include "homyak.selectorLabels" . | nindent 6 }}
      strategy:
        type: RollingUpdate
        rollingUpdate:
          maxSurge: 1
          maxUnavailable: 0
      template:
        metadata:
          annotations:
            {{- with .Values.podAnnotations }}
            {{- toYaml . | nindent 8 }}
            {{- end }}
      labels:
        {{- include "homyak.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "homyak.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.port }}
              protocol: TCP
          env:
          - name: APP_DEBUG
            value: {{ .Values.config.app.debug | quote }}
          - name: APP_ENV
            value: {{ .Values.config.app.env | quote }}
          - name: APP_FEEDBACK_FORM_URL
            value: {{ .Values.config.app.feedbackFormUrl | quote }}
          - name: APP_HTTP_ADDR
            value: {{ .Values.config.http.addr | quote }}
          - name: APP_HTTP_MAX_HEADER_SIZE_MB
            value: {{ .Values.config.http.maxHeaderSizeMb | quote }}
          - name: APP_HTTP_READ_TIMEOUT
            value: {{ .Values.config.http.readTimeout | quote }}
          - name: APP_HTTP_WRITE_TIMEOUT
            value: {{ .Values.config.http.writeTimeout | quote }}
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

### Files to Create

| File | Description |
|------|-------------|
| `helm/Chart.yaml` | Chart metadata |
| `helm/values.yaml` | Default configuration values |
| `helm/templates/_helpers.tpl` | Template helpers |
| `helm/templates/deployment.yaml` | Deployment resource |
| `helm/templates/service.yaml` | Service resource |
| `helm/templates/ingress.yaml` | Ingress with TLS |
| `helm/templates/serviceaccount.yaml` | ServiceAccount |
| `helm/.helmignore` | Files to exclude |
| `helm/README.md` | Usage documentation |

### Commands

```bash
# Create chart directory structure
mkdir -p helm/templates

# Lint chart
helm lint ./helm

# Test template rendering
helm template --dry-run homyak ./helm

# Install chart (dry run)
helm install --dry-run --debug homyak ./helm

# Install chart (real)
helm install homyak ./helm --namespace homyak --create-namespace

# Upgrade chart
helm upgrade homyak ./helm --namespace homyak

# Rollback
helm rollback homyak

# Uninstall
helm uninstall homyak --namespace homyak
```

## Acceptance Criteria

- [ ] `helm lint ./helm` passes without errors
- [ ] `helm template --dry-run homyak ./helm` produces valid YAML
- [ ] Chart follows Helm best practices
- [ ] All resources have proper labels and annotations
- [ ] Deployment uses RollingUpdate strategy
- [ ] Health probes are configured
- [ ] Security context is set (non-root, read-only)
- [ ] Resource limits and requests are defined
- [ ] Ingress has cert-manager annotation
- [ ] Application config uses direct environment variables
- [ ] README.md with usage examples is included

## Dependencies

- Kubernetes cluster (kubespray-managed)
- Helm 3.x installed locally
- cert-manager installed in cluster
- Ingress controller (assumed nginx)

## References

- [Helm Chart Best Practices](https://helm.sh/docs/chart_best_practices/)
- [Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
