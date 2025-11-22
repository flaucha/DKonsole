# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)
![Version](https://img.shields.io/badge/version-1.0.5-green.svg)

**DKonsole** is a modern, lightweight Kubernetes dashboard built entirely with **Artificial Intelligence**. It provides an intuitive interface to manage your cluster resources, view logs, execute commands in pods, and monitor historical metrics with Prometheus integration.

## 🤖 Built with AI

This entire project, from backend to frontend and infrastructure code, was generated using advanced AI agents. It demonstrates the power of AI in modern software development.

## ✨ Features

- 🎯 **Resource Management**: View and manage Deployments, Pods, Services, ConfigMaps, Secrets, and more
- 📊 **Prometheus Integration**: Historical metrics for Pods with customizable time ranges (1h, 6h, 12h, 1d, 7d, 15d)
- 📝 **Live Logs**: Stream logs from containers in real-time
- 💻 **Terminal Access**: Execute commands directly in pod containers
- ✏️ **YAML Editor**: Edit resources with a built-in YAML editor
- 🔐 **Secure Authentication**: Argon2 password hashing and JWT-based sessions
- 🌐 **Multi-Cluster Support**: Manage multiple Kubernetes clusters from a single interface

## 🚀 Quick Start

### 1. Deploy with Helm

```bash
# Add the repo (if applicable) or clone
git clone https://github.com/flaucha/DKonsole.git
cd DKonsole

# Checkout the latest stable version
git checkout v1.0.5

# Install
helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
```

## ⚙️ Configuration

The `values.yaml` file is designed to be simple. You only need to configure the essentials:

### 1. Authentication (Required)
You must provide an `admin` username and an Argon2 `passwordHash`. You also need a `jwtSecret` for session security.

```yaml
admin:
  username: admin
  passwordHash: "$argon2id$..." # Generate with argon2 tool
jwtSecret: "..." # Generate with openssl rand -base64 32
```

**Generate password hash:**
```bash
echo -n "yourpassword" | argon2 $(openssl rand -base64 16) -id -t 3 -m 12 -p 1 -l 32 -e
```

### 2. Ingress (Required for external access)
Configure your domain and TLS settings to access the dashboard.

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: dkonsole.example.com
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
        - dkonsole.example.com

# Optional: Restrict WebSocket origins (CORS)
allowedOrigins: "https://dkonsole.example.com"
```

### 3. Prometheus Integration (Optional)
Enable historical metrics by configuring your Prometheus endpoint.

```yaml
prometheusUrl: "http://prometheus-server.monitoring.svc.cluster.local:9090"
```

**Features enabled with Prometheus:**
- Historical CPU and memory metrics for Pods
- Time range selector (1 hour, 6 hours, 12 hours, 1 day, 7 days, 15 days)
- Metrics tab in Pod details view

**Note:** If `prometheusUrl` is not configured, the Metrics tab will not be displayed.

### 4. Docker Images (Optional)
By default, it uses the official images. You can change tags or repositories if needed.

```yaml
image:
  backend:
    repository: dkonsole/dkonsole-backend
    tag: "1.0.5"
  frontend:
    repository: dkonsole/dkonsole-frontend
    tag: "1.0.5"
```

## 🐳 Docker Images

The official images are available at:

- **Backend**: `dkonsole/dkonsole-backend:1.0.5`
- **Frontend**: `dkonsole/dkonsole-frontend:1.0.5`

## 📊 Prometheus Metrics

DKonsole integrates with Prometheus to provide historical metrics visualization. The following PromQL queries are used:

**CPU Usage (millicores):**
```promql
sum(rate(container_cpu_usage_seconds_total{namespace="<namespace>",pod="<pod-name>",container!=""}[5m])) * 1000
```

**Memory Usage (MiB):**
```promql
sum(container_memory_working_set_bytes{namespace="<namespace>",pod="<pod-name>",container!=""}) / 1024 / 1024
```

## 💰 Support the Project

If you find this project useful, consider donating to support development.

**BSC (Binance Smart Chain) Wallet:**
`0x9baf648fa316030e12b15cbc85278fdbd82a7d20`

## 📧 Contact

For questions or feedback, please contact: **flaucha@gmail.com**

## 🛠️ Development

To run locally:

```bash
# Backend
cd backend && go run main.go

# Frontend
cd frontend && npm run dev
```

## 📝 Changelog

### v1.0.5 (2025-11-22)
- 🔧 Improved build and release scripts
- 📝 Added `build.sh` for simple Docker builds
- 📝 Added `release.sh` for automated releases with git tagging
- 📚 Added `SCRIPTS.md` documentation
- 🔄 Deprecated `deploy.sh` in favor of new scripts

### v1.0.4 (2025-11-22)
- ✨ Added Prometheus integration for Pod metrics
- 📊 Historical metrics with time range selector (1h, 6h, 12h, 1d, 7d, 15d)
- 🎨 Added Metrics tab to Pod details
- 🔧 Removed Metrics tab from Deployment details (kept for future use)
- 🐛 Fixed namespace display for cluster-scoped resources (Nodes, ClusterRoles, etc.)

### v1.0.3
- Initial stable release

## License

MIT License
