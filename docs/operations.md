# Operations Guide

This guide covers daily operations, troubleshooting, and maintenance procedures for the homyak deployment.

## Daily Operations

### Check Application Status

```bash
kubectl get all -n homyak
```

### View Logs

```bash
# All pods
kubectl logs -n homyak -l app.kubernetes.io/name=homyak --tail=100 -f

# Specific pod
kubectl logs -n homyak <pod-name> --tail=100 -f
```

### Scale Application

```bash
# Scale up to 4 replicas
kubectl scale deployment homyak --replicas=4 -n homyak

# Scale down to 2 replicas
kubectl scale deployment homyak --replicas=2 -n homyak
```

### Restart Pods

```bash
# Restart all pods (rolling restart)
kubectl rollout restart deployment homyak -n homyak

# Watch rollout progress
kubectl rollout status deployment homyak -n homyak
```

### View Resource Usage

```bash
# Pod resource usage
kubectl top pods -n homyak

# Node resource usage
kubectl top nodes
```

## Deployment Operations

### Deploy New Version

**Via GitHub Actions** (recommended):
1. Push changes to `main` branch
2. Monitor workflow at: https://github.com/victorzhuk/homyak/actions
3. Deployment happens automatically

**Via Helm** (manual):
```bash
# Get latest image tag from GitHub Container Registry
helm upgrade homyak ./helm \
  --namespace homyak \
  --set image.tag=v1.2.3 \
  --wait \
  --timeout 5m
```

### Rollback Deployment

```bash
# List deployment history
helm history homyak -n homyak

# Rollback to previous version
helm rollback homyak -n homyak

# Rollback to specific revision
helm rollback homyak 2 -n homyak
```

### Upgrade Helm Chart

```bash
# Update chart dependencies
helm dependency update ./helm

# Apply chart changes
helm upgrade homyak ./helm \
  --namespace homyak \
  --reuse-values \
  --wait
```

### Uninstall Application

```bash
helm uninstall homyak -n homyak
kubectl delete namespace homyak
```

## Troubleshooting

### Pods Not Starting

1. **Check pod status**:
   ```bash
   kubectl describe pod <pod-name> -n homyak
   ```

2. **Check events**:
   ```bash
   kubectl get events -n homyak --sort-by='.lastTimestamp'
   ```

3. **Common issues**:
   - **Image pull errors**: Check if image tag exists, verify registry credentials
   - **CrashLoopBackOff**: Check logs for application errors
   - **Pending**: Check resource requests vs available node capacity

### Application Not Responding

1. **Check pod health**:
   ```bash
   kubectl describe pod <pod-name> -n homyak | grep -A 10 "Health"
   ```

2. **Test from inside cluster**:
   ```bash
   kubectl run debug --rm -it --image=busybox --restart=Never -- wget -qO- http://homyak.homyak.svc.cluster.local:8080/healthz
   ```

3. **Check service connectivity**:
   ```bash
   kubectl get svc homyak -n homyak
   kubectl get endpoints homyak -n homyak
   ```

### Ingress/TLS Issues

1. **Check ingress status**:
   ```bash
   kubectl get ingress homyak -n homyak -o yaml
   ```

2. **Check certificate**:
   ```bash
   kubectl get certificate -A
   kubectl get secret homyak-tls -n homyak -o yaml
   ```

3. **Check cert-manager logs**:
   ```bash
   kubectl logs -n cert-manager -l app.kubernetes.io/name=cert-manager
   ```

4. **DNS verification**:
   ```bash
   dig victorzh.uk +short
   nslookup victorzh.uk
   ```

### High Resource Usage

1. **Identify high consumers**:
   ```bash
   kubectl top pods -n homyak
   ```

2. **Check resource limits**:
   ```bash
   kubectl get deployment homyak -n homyak -o jsonpath='{.spec.template.spec.containers[0].resources}'
   ```

3. **Increase resources**:
   ```bash
   kubectl set resources deployment homyak \
     --requests=cpu=200m,memory=256Mi \
     --limits=cpu=500m,memory=512Mi \
     -n homyak
   ```

### Database Connection Issues

1. **Check secret**:
   ```bash
   kubectl get secret homyak-secret -n homyak -o yaml
   ```

2. **Verify connection from pod**:
   ```bash
   kubectl exec -it <pod-name> -n homyak -- nc -zv <database-host> <database-port>
   ```

3. **Update database URL**:
   ```bash
   kubectl create secret generic homyak-secret \
     --from-literal=database-url="postgresql://..." \
     --dry-run=client -o yaml | kubectl apply -f -
   
   kubectl rollout restart deployment homyak -n homyak
   ```

## Backup and Recovery

### Backup Configuration

```bash
# Backup Helm values
helm get values homyak -n homyak > homyak-values-backup.yaml

# Backup secrets
kubectl get secret homyak-secret -n homyak -o yaml > homyak-secret-backup.yaml

# Backup deployment state
kubectl get deployment homyak -n homyak -o yaml > homyak-deployment-backup.yaml
```

### Restore Configuration

```bash
# Restore from backup
helm upgrade homyak ./helm \
  --namespace homyak \
  --values homyak-values-backup.yaml \
  --wait
```

## Monitoring and Alerting

### Setup Metrics

Enable Prometheus metrics (if available):
```bash
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
```

### Health Check Endpoints

- **Liveness**: `GET /healthz` - Container restarts if this fails
- **Readiness**: `GET /readyz` - Traffic stops if this fails
- **Metrics**: `GET /metrics` - Prometheus metrics endpoint

### Log Aggregation

```bash
# Export logs to file
kubectl logs -n homyak -l app.kubernetes.io/name=homyak --tail=1000 > homyak-logs.txt

# Filter by log level
kubectl logs -n homyak -l app.kubernetes.io/name=homyak | grep "ERROR"
```

### Alerts to Monitor

- Pods in CrashLoopBackOff state
- High CPU/memory usage (above 80%)
- Failed readiness probes
- Certificate expiration (within 7 days)
- High error rate in logs

## Maintenance

### Regular Maintenance Tasks

**Daily**:
- Check pod status: `kubectl get pods -n homyak`
- Review recent logs for errors

**Weekly**:
- Review resource usage: `kubectl top pods -n homyak`
- Check certificate expiration
- Review deployment history: `helm history homyak -n homyak`

**Monthly**:
- Rotate kubeconfig (update KUBECONFIG_BASE64 secret)
- Review and update image tags
- Audit access logs
- Clean up old Helm releases: `helm list -A`

### Node Maintenance

**Drain node for maintenance**:
```bash
kubectl cordon <node-name>
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
```

**Uncordon after maintenance**:
```bash
kubectl uncordon <node-name>
```

## Security

### Security Audits

```bash
# Check for privileged containers
kubectl get pods -n homyak -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].securityContext.privileged}{"\n"}{end}'

# Check for secrets in environment variables
kubectl get pods -n homyak -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{range .spec.containers[*].env[*]}{.name}{"\n"}{end}{end}'
```

### Network Policies

Apply network policy to restrict traffic:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: homyak-network-policy
  namespace: homyak
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: homyak
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: homyak
    ports:
    - protocol: TCP
      port: 8080
```

## Useful Commands Reference

```bash
# Quick status
kubectl get all -n homyak

# Watch pods
watch kubectl get pods -n homyak

# Port forward to localhost
kubectl port-forward -n homyak deployment/homyak 8080:8080

# Execute command in pod
kubectl exec -it <pod-name> -n homyak -- /bin/sh

# Copy files from pod
kubectl cp <pod-name>:/path/to/file ./local-file -n homyak

# Edit deployment
kubectl edit deployment homyak -n homyak

# Get pod YAML
kubectl get pod <pod-name> -n homyak -o yaml
```

## Escalation

If you encounter issues beyond the scope of this guide:

1. Check application logs for error details
2. Review Kubernetes events: `kubectl get events -n homyak`
3. Consult Helm chart documentation: `helm show readme ./helm`
4. Review GitHub Actions workflow logs
5. Check cluster resources: `kubectl describe node <node-name>`

## Contact

For issues or questions about this deployment, contact:
- GitHub Issues: https://github.com/victorzhuk/homyak/issues
- Email: victor@victorzh.uk
