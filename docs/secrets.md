# GitHub Secrets Configuration

This document describes the GitHub secrets required for automated deployment to Kubernetes.

## Required Secrets

### KUBECONFIG_BASE64

The base64-encoded kubeconfig file for accessing your Kubernetes cluster.

### How to Generate KUBECONFIG_BASE64

1. **Export your kubeconfig**:
   ```bash
   cat ~/.kube/config | base64 -w 0
   ```

   Or on macOS (without `-w 0` option):
   ```bash
   cat ~/.kube/config | base64 | tr -d '\n'
   ```

2. **Copy the output** - this is your base64-encoded kubeconfig.

3. **Add the secret to GitHub**:
   - Go to: https://github.com/victorzhuk/homyak/settings/secrets/actions
   - Click: "New repository secret"
   - Name: `KUBECONFIG_BASE64`
   - Value: Paste the base64 string from step 2
   - Click: "Add secret"

### Security Best Practices

1. **Use service accounts**: Create a dedicated Kubernetes service account with minimal permissions.
   ```bash
   kubectl create serviceaccount github-actions
   kubectl create clusterrolebinding github-actions \
     --clusterrole=cluster-admin \
     --serviceaccount=default:github-actions
   ```

2. **Limit namespace access**: Grant access only to the `homyak` namespace.
   ```bash
   kubectl create role homyak-admin \
     --namespace=homyak \
     --verb="*" \
     --resource="*"
   
   kubectl create rolebinding github-actions-homyak \
     --namespace=homyak \
     --role=homyak-admin \
     --serviceaccount=default:github-actions
   ```

3. **Use short-lived credentials**: Rotate your kubeconfig regularly (every 90 days).

4. **Monitor access**: Audit GitHub Actions logs regularly.

5. **Never commit secrets**: Never add kubeconfig or secrets to the repository.

### Testing the Secret

After adding the secret, test it by triggering the deployment workflow:

1. Go to: https://github.com/victorzhuk/homyak/actions
2. Select "deploy" workflow
3. Click "Run workflow" button
4. Monitor the workflow execution

If the secret is correctly configured, the workflow should:
- Successfully connect to the cluster
- Verify cluster info
- Deploy the application

### Troubleshooting

#### Error: "Invalid kubeconfig"

- Verify the base64 string is complete (no line breaks)
- Re-generate the base64 string
- Ensure the secret name is exactly `KUBECONFIG_BASE64`

#### Error: "Unauthorized"

- Check if the kubeconfig token is expired
- Verify the service account has proper permissions
- Regenerate the kubeconfig from a fresh login

#### Error: "Cluster connection failed"

- Verify cluster is accessible from your local machine
- Check network connectivity
- Verify kubeconfig points to the correct cluster

### Secret Rotation Process

1. Generate new base64 kubeconfig
2. Update the secret in GitHub
3. Trigger a test deployment
4. Verify deployment succeeds
5. Remove old kubeconfig (if applicable)

### Additional Secrets (Optional)

If you need to pass additional secrets to the application:

1. **Create Kubernetes secrets manually**:
   ```bash
   kubectl create secret generic homyak-secrets \
     --namespace=homyak \
     --from-literal=database-url="postgresql://..." \
     --from-literal=api-key="your-api-key"
   ```

2. **Or use GitHub Secrets** (requires workflow modification):
   - Add secret to GitHub
   - Reference in workflow using `${{ secrets.SECRET_NAME }}`
   - Pass via `--set-file` in helm command

### References

- [GitHub Actions Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Kubernetes Authentication Documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/)
- [Helm Security Best Practices](https://helm.sh/docs/chart_best_practices/security/)
