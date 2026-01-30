# cert-manager Setup Guide

This guide covers cert-manager installation and configuration for TLS certificate provisioning.

## Prerequisites

- Kubernetes cluster 1.23+
- kubectl configured
- NGINX Ingress Controller installed
- Domain with DNS pointing to cluster ingress IP

## Installation

### Install cert-manager

```bash
# Add cert-manager Helm repository
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Install cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.14.0 \
  --set installCRDs=true \
  --set prometheus.enabled=true \
  --set prometheus.servicemonitor.enabled=true
```

### Verify Installation

```bash
# Check cert-manager pods
kubectl get pods -n cert-manager

# Expected output:
# NAME                                      READY   STATUS    RESTARTS   AGE
# cert-manager-5c6866597-zg8p9           1/1     Running   0          2m
# cert-manager-cainjector-565f695868-9c2x2  1/1     Running   0          2m
# cert-manager-webhook-6486c7d6f9-tqh9v      1/1     Running   0          2m
```

## ClusterIssuer Configuration

### Create Let's Encrypt Production Issuer

Save as `clusterissuer.yaml`:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: victor@victorzh.uk
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### Create Let's Encrypt Staging Issuer (for testing)

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: victor@victorzh.uk
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: nginx
```

### Apply ClusterIssuers

```bash
kubectl apply -f clusterissuer.yaml

# Verify issuers are ready
kubectl get clusterissuer
```

Expected output:
```
NAME                 READY   AGE
letsencrypt-prod     True    2m
letsencrypt-staging   True    2m
```

## DNS Requirements

Before requesting certificates, ensure:

1. **DNS A record exists** pointing to your cluster's ingress IP:
   ```
   victorzh.uk  A  <your-cluster-ingress-ip>
   ```

2. **DNS propagation is complete**:
   ```bash
   dig victorzh.uk +short
   ```

3. **Can resolve from internet**:
   ```bash
   curl -I https://victorzh.uk
   ```

## Certificate Request

Certificates are automatically provisioned when you create an Ingress with the `cert-manager.io/cluster-issuer` annotation.

### Example Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: homyak
  annotations:
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

After creating the Ingress, cert-manager automatically:
1. Creates a Certificate resource
2. Requests certificate from Let's Encrypt
3. Performs HTTP-01 challenge validation
4. Stores the certificate in the TLS Secret

### Verify Certificate

```bash
# Check certificate status
kubectl get certificate -n homyak

# Describe certificate for details
kubectl describe certificate homyak-tls -n homyak

# View certificate details
kubectl get secret homyak-tls -n homyak -o yaml
```

## Troubleshooting

### Certificate Pending

If certificate remains in "Pending" state:

```bash
# Check Certificate resource
kubectl describe certificate <certificate-name> -n <namespace>

# Common causes:
# 1. DNS not pointing to cluster
# 2. Ingress not accessible from internet
# 3. Port 80 blocked
# 4. HTTP-01 challenge failing
```

### Certificate Failed

```bash
# Check Certificate resource
kubectl describe certificate <certificate-name> -n <namespace>

# Check Order resource
kubectl get order -n <namespace>
kubectl describe order <order-name> -n <namespace>

# Check Challenge resource
kubectl get challenge -n <namespace>
kubectl describe challenge <challenge-name> -n <namespace>
```

Common failures:
- **DNS propagation delay**: Wait 10-30 minutes after DNS change
- **Port 80 blocked**: Ensure HTTP traffic can reach ingress
- **Rate limiting**: Let's Encrypt rate limits (5 certs/week per domain)
- **Invalid email**: Email address in ClusterIssuer must be valid

### Certificate Renewal

cert-manager automatically renews certificates before expiration (30 days before).

Force renewal:
```bash
kubectl annotate certificate <certificate-name> \
  -n <namespace> \
  cert-manager.io/issue-temporary-certificate="true"

# Then remove annotation
kubectl annotate certificate <certificate-name> \
  -n <namespace> \
  cert-manager.io/issue-temporary-certificate- \
```

### Check cert-manager Logs

```bash
# cert-manager logs
kubectl logs -n cert-manager -l app.kubernetes.io/name=cert-manager --tail=100 -f

# cainjector logs
kubectl logs -n cert-manager -l app.kubernetes.io/name=cainjector --tail=100

# webhook logs
kubectl logs -n cert-manager -l app.kubernetes.io/name=webhook --tail=100
```

### Reset cert-manager

If cert-manager is misconfigured:

```bash
# Delete cert-manager
helm uninstall cert-manager -n cert-manager

# Delete CRDs
kubectl delete crd \
  certificaterequests.cert-manager.io \
  certificates.cert-manager.io \
  challenges.acme.cert-manager.io \
  clusterissuers.cert-manager.io \
  issuers.cert-manager.io \
  orders.acme.cert-manager.io

# Reinstall cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.14.0 \
  --set installCRDs=true
```

## Testing with Staging Issuer

Use Let's Encrypt staging issuer to test configuration without rate limits:

1. Update Ingress annotation:
   ```yaml
   cert-manager.io/cluster-issuer: letsencrypt-staging
   ```

2. Apply change:
   ```bash
   kubectl apply -f ingress.yaml
   ```

3. Wait for certificate
4. Test TLS connection:
   ```bash
   curl -vI https://victorzh.uk
   ```

5. Switch to production issuer:
   ```yaml
   cert-manager.io/cluster-issuer: letsencrypt-prod
   ```

## Certificate Expiration

Check expiration:
```bash
kubectl get certificate -n homyak \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.notAfter}{"\n"}{end}'
```

### Renew Certificate Manually

```bash
# Delete secret (triggers renewal)
kubectl delete secret homyak-tls -n homyak

# Or delete and recreate certificate
kubectl delete certificate homyak-tls -n homyak
kubectl apply -f certificate.yaml
```

## Best Practices

1. **Use staging first**: Test with staging issuer before production
2. **Monitor renewal**: Set up alerts for certificate expiration
3. **Backup secrets**: Export TLS secrets for disaster recovery
4. **Rate limits**: Be aware of Let's Encrypt rate limits
5. **DNS management**: Use reliable DNS provider with fast propagation
6. **Test certificates**: Verify certificates are valid before going live

## Security Considerations

1. **Email privacy**: Email used in ACME challenges is public
2. **Secret access**: TLS secrets should have restricted access
3. **Certificate rotation**: Regular renewal required
4. **Key length**: cert-manager uses 2048-bit RSA by default (can be increased)
5. **Certificate pinning**: Avoid certificate pinning (complicates renewal)

## Monitoring

### Setup Prometheus Metrics

cert-manager exposes Prometheus metrics:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cert-manager
  namespace: cert-manager
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: cert-manager
  endpoints:
  - port: tcp-prometheus-servicemonitor
    interval: 30s
```

### Alert on Certificate Expiration

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cert-manager-alerts
  namespace: cert-manager
spec:
  groups:
  - name: certificate-alerts
    rules:
    - alert: CertificateExpiringSoon
      expr: cert_manager_certificate_expiration_timestamp_seconds < time() + 86400 * 7
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Certificate {{ $labels.name }} expiring in less than 7 days"
        description: "Certificate {{ $labels.name }} in namespace {{ $labels.namespace }} expires in {{ $value | humanizeDuration }}"
```

## References

- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/)
- [ACME Protocol](https://tools.ietf.org/html/rfc8555)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
