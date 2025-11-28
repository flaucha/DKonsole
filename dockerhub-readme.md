# DKonsole

![Version](https://img.shields.io/badge/version-1.2.7-green.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)

**DKonsole** is a modern, lightweight Kubernetes dashboard built entirely with **Artificial Intelligence**. It provides an intuitive web interface to manage your cluster resources, view logs, execute commands in pods, and monitor historical metrics with Prometheus integration.

## 🚀 Quick Start

### Using Docker

```bash
# Run DKonsole locally (for testing)
docker run -d \
  --name dkonsole \
  -p 8080:8080 \
  -v ~/.kube/config:/home/app/.kube/config:ro \
  dkonsole/dkonsole:latest
```

Access the dashboard at `http://localhost:8080`

### Using Helm (Recommended)

```bash
# Clone the repository
git clone https://github.com/flaucha/DKonsole.git
cd DKonsole

# Checkout the latest stable version
git checkout v1.2.7

# Configure ingress and allowedOrigins
vim ./helm/dkonsole/values.yaml

# Install
helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
```

After installation, access the web interface to complete the initial setup.

## ✨ Features

- 🎯 **Resource Management**: View and manage Deployments, Pods, Services, ConfigMaps, Secrets, and more
- 📊 **Prometheus Integration**: Historical metrics for Pods with customizable time ranges (1h, 6h, 12h, 1d, 7d, 15d)
- 📝 **Live Logs**: Stream logs from containers in real-time
- 💻 **Terminal Access**: Execute commands directly in pod containers via WebSocket
- ✏️ **YAML Editor**: Edit resources with a built-in YAML editor
- 🔐 **Secure Authentication**: Argon2 password hashing and JWT-based sessions
- 🎨 **Modern UI**: Clean, responsive interface built with React
- 🔄 **Setup Mode**: Web-based initial configuration - no manual secret creation needed

## 📋 Image Details

- **Base Image**: `alpine:3.19`
- **Port**: `8080`
- **User**: Non-root user (UID 1000)
- **Health Check**: `/healthz` endpoint available

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `JWT_SECRET` | Secret key for JWT token signing (min 32 chars) | No* | Auto-generated in setup mode |
| `ADMIN_USER` | Admin username | No* | Set via web interface |
| `ADMIN_PASSWORD` | Admin password (Argon2 hash) | No* | Set via web interface |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | No | Same-origin only |
| `PROMETHEUS_URL` | Prometheus server URL for metrics | No | Metrics disabled |

\* *Not required in setup mode - configured via web interface*

### Setup Mode

DKonsole automatically detects if the authentication secret doesn't exist and enters **Setup Mode**:

1. Deploy the Helm chart (no authentication configuration needed)
2. Access the web interface
3. Complete the setup form to create the admin account
4. The secret is created automatically in Kubernetes
5. Login with your configured credentials

The setup creates a Kubernetes secret (`{release-name}-auth`) with:
- Admin username
- Argon2-hashed password
- JWT secret for session security

## 📦 Helm Chart

The recommended way to deploy DKonsole is using the included Helm chart:

```bash
helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
```

### Minimal Configuration

At minimum, configure the ingress:

```yaml
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: dkonsole.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: dkonsole-tls
      hosts:
        - dkonsole.example.com

allowedOrigins: "https://dkonsole.example.com"
```

## 🔒 Security

- **Password Hashing**: Argon2id with secure random salt generation
- **JWT Sessions**: HTTP-only cookies for secure session management
- **RBAC**: Full Kubernetes RBAC support - configure permissions via ServiceAccount
- **Non-root**: Container runs as non-root user (UID 1000)
- **Security Headers**: Built-in security headers middleware

## 📊 Prometheus Integration

To enable historical metrics visualization:

1. Set `PROMETHEUS_URL` environment variable or configure in Helm values
2. Ensure Prometheus can scrape your cluster metrics
3. The Metrics tab will appear in Pod details with time range selector

Example:
```yaml
prometheusUrl: "http://prometheus-server.monitoring.svc.cluster.local:9090"
```

## 🏗️ Architecture

- **Backend**: Go 1.24 with Kubernetes client-go
- **Frontend**: React with Vite
- **Authentication**: JWT with Argon2 password hashing
- **Storage**: Kubernetes Secrets for credentials, optional PVC for custom logos

## 📚 Documentation

- **GitHub Repository**: https://github.com/flaucha/DKonsole
- **Full Documentation**: See README.md in the repository
- **Helm Chart**: Included in `./helm/dkonsole/`

## 🤝 Support

For questions or feedback:
- **Email**: flaucha@gmail.com
- **GitHub Issues**: https://github.com/flaucha/DKonsole/issues

## 📝 License

MIT License - see LICENSE file for details

## 🙏 Credits

This project was built entirely with Artificial Intelligence, demonstrating the power of AI in modern software development.

---

**Tags**: `latest`, `1.2.7`, `1.2.6`, `1.2.5` and more. See [Docker Hub tags](https://hub.docker.com/r/dkonsole/dkonsole/tags) for all available versions.
