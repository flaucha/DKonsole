# DKonsole - Kubernetes Dashboard

![License](https://img.shields.io/badge/license-MIT-blue.svg)

A modern, lightweight Kubernetes dashboard for cluster management and monitoring. DKonsole provides an intuitive interface to manage deployments, services, pods, and other Kubernetes resources.

## Features

- 🚀 **Resource Management**: View, create, edit, and delete Kubernetes resources
- 📊 **Cluster Overview**: Real-time cluster metrics and statistics
- 🔍 **API Explorer**: Discover and interact with Kubernetes APIs
- 📦 **Namespace Management**: Manage namespaces, resource quotas, and limit ranges
- 🔐 **Secure Authentication**: JWT-based authentication with Argon2 password hashing
- 📝 **YAML Editor**: Built-in Monaco editor for YAML manifests
- 📈 **Resource Scaling**: Scale deployments directly from the UI
- 🔌 **Pod Exec**: Execute commands in pods via WebSocket
- 📜 **Log Streaming**: Real-time pod log viewing

## Prerequisites

- **Kubernetes Cluster**: v1.19+ recommended
- **Helm**: v3.0+
- **kubectl**: Configured to access your cluster
- **Metrics Server** (optional): For pod/node metrics
- **Ingress Controller** (optional): For external access (e.g., nginx-ingress)
- **cert-manager** (optional): For automatic TLS certificates

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/DKonsole.git
cd DKonsole
```

### 2. Generate Admin Credentials

Generate a secure password hash using Argon2:

```bash
# Install argon2 if not available
# Ubuntu/Debian: apt-get install argon2
# macOS: brew install argon2

# Generate hash
echo -n "your-secure-password" | argon2 $(openssl rand -base64 16) -id -t 3 -m 12 -p 1 -l 32 -e
```

Save the output (starts with `$argon2id$...`)

### 3. Generate JWT Secret

```bash
openssl rand -base64 32
```

### 4. Create values file

Create a `my-values.yaml` file:

```yaml
admin:
  username: admin
  passwordHash: "$argon2id$v=19$m=4096,t=3,p=1$<your-hash-here>"

jwtSecret: "<your-jwt-secret-here>"

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: dkonsole.yourdomain.com
      paths:
        - path: /api
          pathType: Prefix
          backend: backend
        - path: /
          pathType: Prefix
          backend: frontend
  tls:
    - secretName: dkonsole-tls
      hosts:
        - dkonsole.yourdomain.com

image:
  backend:
    repository: your-registry/dkonsole-backend
    tag: "latest"
  frontend:
    repository: your-registry/dkonsole-frontend
    tag: "latest"
```

### 5. Install with Helm

```bash
# Create namespace
kubectl create namespace dkonsole

# Install chart
helm install dkonsole ./helm/dkonsole \
  --namespace dkonsole \
  --values my-values.yaml
```

### 6. Access the Dashboard

```bash
# If using Ingress
open https://dkonsole.yourdomain.com

# Or port-forward for local access
kubectl port-forward -n dkonsole svc/dkonsole-frontend 8080:80
open http://localhost:8080
```

Login with the username and password you configured.

## Building Docker Images

### Backend

```bash
cd backend
docker build -t your-registry/dkonsole-backend:latest .
docker push your-registry/dkonsole-backend:latest
```

### Frontend

```bash
cd frontend
docker build -t your-registry/dkonsole-frontend:latest .
docker push your-registry/dkonsole-frontend:latest
```

## Configuration

### RBAC Permissions

DKonsole requires cluster-wide permissions to manage resources. The Helm chart automatically creates:

- **ServiceAccount**: `dkonsole`
- **ClusterRole**: With permissions for all managed resources
- **ClusterRoleBinding**: Binds the role to the service account

To customize permissions, edit `values.yaml` under `rbac.clusterResources` and `rbac.namespacedResources`.

### Persistence

Logo uploads are persisted using a PersistentVolumeClaim:

```yaml
persistence:
  enabled: true
  storageClass: "standard"  # Your storage class
  size: 1Gi
```

### Resource Limits

Adjust resource requests and limits in `values.yaml`:

```yaml
resources:
  backend:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

### Ingress Configuration

For different ingress controllers, adjust annotations:

```yaml
ingress:
  annotations:
    # For Traefik
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    
    # For ALB (AWS)
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
```

## Security Considerations

1. **Change Default Credentials**: Always use secure, unique passwords
2. **Use HTTPS**: Enable TLS for production deployments
3. **Rotate Secrets**: Regularly rotate JWT secrets and passwords
4. **RBAC**: Review and restrict permissions as needed
5. **Network Policies**: Consider implementing network policies to restrict traffic
6. **Pod Security**: The chart uses restrictive security contexts by default

## Upgrading

```bash
helm upgrade dkonsole ./helm/dkonsole \
  --namespace dkonsole \
  --values my-values.yaml
```

## Uninstalling

```bash
helm uninstall dkonsole --namespace dkonsole
kubectl delete namespace dkonsole
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl get pods -n dkonsole

# View logs
kubectl logs -n dkonsole deployment/dkonsolek-backend
kubectl logs -n dkonsole deployment/dkonsole-frontend
```

### Authentication issues

Verify secret is created correctly:

```bash
kubectl get secret -n dkonsole dkonsole-auth -o yaml
```

### RBAC permissions

Check if ServiceAccount has proper permissions:

```bash
kubectl auth can-i --as=system:serviceaccount:dkonsole:dkonsole get pods --all-namespaces
```

## Development

### Local Development

```bash
# Backend
cd backend
go run main.go

# Frontend
cd frontend
npm install
npm run dev
```

### Environment Variables

Backend requires:
- `JWT_SECRET`: Secret for JWT signing
- `ADMIN_USER`: Admin username
- `ADMIN_PASSWORD`: Argon2 password hash

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- GitHub Issues: https://github.com/yourusername/DKonsole/issues
- Discussions: https://github.com/yourusername/DKonsole/discussions
