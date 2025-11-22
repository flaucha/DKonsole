# DKonsole Deployment Guide

## Quick Deployment Script

```bash
#!/bin/bash
set -e

echo "🚀 DKonsole Deployment Script"
echo "=============================="

# Configuration
NAMESPACE="dkonsole"
RELEASE_NAME="dkonsole"
DOMAIN="dkonsole.example.com"

# Generate admin password hash
echo "📝 Generating admin credentials..."
echo -n "Enter admin password: "
read -s ADMIN_PASS
echo
ADMIN_HASH=$(echo -n "$ADMIN_PASS" | argon2 $(openssl rand -base64 16) -id -t 3 -m 12 -p 1 -l 32 -e)

# Generate JWT secret
echo "🔐 Generating JWT secret..."
JWT_SECRET=$(openssl rand -base64 32)

# Create namespace
echo "📦 Creating namespace $NAMESPACE..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Create values file
cat > /tmp/dkonsole-values.yaml <<EOF
admin:
  username: admin
  passwordHash: "$ADMIN_HASH"

jwtSecret: "$JWT_SECRET"

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: $DOMAIN
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
        - $DOMAIN

image:
  backend:
    repository: your-registry/dkonsole-backend
    tag: "latest"
  frontend:
    repository: your-registry/dkonsole-frontend
    tag: "latest"
EOF

# Install/upgrade Helm chart
echo "📊 Installing Helm chart..."
helm upgrade --install $RELEASE_NAME ./helm/dkonsole \
  --namespace $NAMESPACE \
  --values /tmp/dkonsole-values.yaml \
  --wait

# Clean up sensitive file
rm /tmp/dkonsole-values.yaml

echo "✅ Deployment complete!"
echo "🌐 Access: https://$DOMAIN"
echo "👤 Username: admin"
echo "🔑 Password: (the one you entered)"
```

## Manual Step-by-Step

### 1. Prerequisites Check

```bash
# Check kubectl
kubectl version --client

# Check Helm
helm version

# Check cluster access
kubectl cluster-info
```

### 2. Build and Push Images

```bash
# Backend
cd backend
docker build -t your-registry/dkonsole-backend:1.0.0 .
docker push your-registry/dkonsole-backend:1.0.0

# Frontend
cd ../frontend
docker build -t your-registry/dkonsole-frontend:1.0.0 .
docker push your-registry/dkonsole-frontend:1.0.0
```

### 3. Prepare Configuration

```bash
# Generate password hash
echo -n "YourSecurePassword123!" | argon2 $(openssl rand -base64 16) -id -t 3 -m 12 -p 1 -l 32 -e

# Generate JWT secret
openssl rand -base64 32
```

### 4. Create Custom Values

Create `my-values.yaml`:

```yaml
admin:
  username: admin
  passwordHash: "<paste-hash-here>"

jwtSecret: "<paste-jwt-secret-here>"

ingress:
  enabled: true
  hosts:
    - host: dkonsole.yourdomain.com
```

### 5. Deploy

```bash
helm install dkonsole ./helm/dkonsole \
  --namespace dkonsole \
  --create-namespace \
  --values my-values.yaml
```

### 6. Verify Deployment

```bash
# Check pods
kubectl get pods -n dkonsole

# Check services
kubectl get svc -n dkonsole

# Check ingress
kubectl get ingress -n dkonsole

# View logs
kubectl logs -n dkonsole -l app.kubernetes.io/component=backend
```

## Upgrading

```bash
# Pull latest code
git pull

# Rebuild images (if needed)
docker build -t your-registry/dkonsole-backend:1.1.0 backend/
docker build -t your-registry/dkonsole-frontend:1.1.0 frontend/

# Update image tags in values
# Then upgrade
helm upgrade dkonsole ./helm/dkonsole \
  --namespace dkonsole \
  --values my-values.yaml \
  --wait
```

## Backup

```bash
# Backup Helm values
helm get values dkonsole -n dkonsole > dkonsole-backup-$(date +%Y%m%d).yaml

# Backup secret
kubectl get secret -n dkonsole dkonsole-auth -o yaml > secret-backup-$(date +%Y%m%d).yaml
```

## Monitoring

```bash
# Watch pods
kubectl get pods -n dkonsole -w

# Follow logs
kubectl logs -n dkonsole -f deployment/dkonsole-backend

# Port forward for local access
kubectl port-forward -n dkonsole svc/dkonsole-frontend 8080:80
```
