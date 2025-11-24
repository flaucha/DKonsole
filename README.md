# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)
![Version](https://img.shields.io/badge/version-1.1.4-green.svg)

<img width="1906" height="947" alt="image" src="https://github.com/user-attachments/assets/99030972-04db-4990-8faa-de41079b671c" />

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
git checkout v1.1.4

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
        - path: /
          pathType: Prefix
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

### 4. Docker Image (Optional)
By default, it uses the official image. You can change tag or repository if needed.

```yaml
image:
  repository: dkonsole/dkonsole
  tag: "1.1.3"
```

## 🐳 Docker Image

The official image is available at:

- **Unified**: `dkonsole/dkonsole:1.1.3`

**Note:** Starting from v1.1.0, DKonsole uses a unified container architecture where the backend serves the frontend static files. This improves security by reducing the attack surface and eliminating inter-container communication.

## 📝 Changelog

### v1.1.3 (2025-01-23)
**🚀 Helm Charts Manager & UI Enhancements**

- ✅ **Helm Charts Manager**: Complete Helm release management interface
- ✅ **Pod Events & Timeline**: New events tab in pod details
- ✅ **UI State Persistence**: Browser remembers selected namespace and current view
- ✅ **Responsive Tables**: Scroll horizontal when window is narrow
- ✅ **Pod Table Enhancements**: Added Ready and Restarts columns
- ✅ **PVC Enhancements**: Added Size column with sorting
- ✅ **Date Formatting**: Standardized across all components

### v1.1.2 (2025-01-23)
**✨ Resource Quota Manager Improvements**

- ✅ **Namespace Selector**: Added namespace filter with All/Selected toggle
- ✅ **Automatic Refresh**: Resources refresh automatically after operations
- ✅ **UI Consistency**: Reorganized layout to match other pages
- ✅ **Menu Simplification**: Streamlined to Edit YAML and Delete only
- ✅ **YAML Editor Fix**: Now uses kubectl apply (Server-Side Apply)
- ✅ **Create Menu**: Changed from hover to click interaction
- ✅ **Template Namespace**: New resources use selected namespace
- ✅ **Color Consistency**: Removed colorful elements for gray-scale design

### v1.1.1 (2024-12-19)
**🐛 Bug Fixes and Improvements**

- ✅ **Fixed Resource Import**: Now accepts ANY Kubernetes resource (including CRDs)
- ✅ **Fixed HPA YAML Viewing**: Resolved issue with HorizontalPodAutoscaler YAML viewing
- ✅ **Added Resource Aliases**: Support for common aliases (HPA, PVC, PV, etc.)
- ✅ **Dynamic Resource Discovery**: Uses Kubernetes Discovery API for automatic resource resolution
- ✅ **Enhanced RBAC**: Full access to all resources for import functionality

### v1.1.0 (2024-12-19)

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

**Buy me a coffee:**
https://buymeacoffee.com/flaucha

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

### v1.1.3 (2025-01-23)
**🚀 Helm Charts Manager & UI Enhancements**

- ✅ **Helm Charts Manager**: Complete Helm release management interface
- ✅ **Pod Events & Timeline**: New events tab in pod details
- ✅ **UI State Persistence**: Browser remembers selected namespace and current view
- ✅ **Responsive Tables**: Scroll horizontal when window is narrow
- ✅ **Pod Table Enhancements**: Added Ready and Restarts columns
- ✅ **PVC Enhancements**: Added Size column with sorting
- ✅ **Date Formatting**: Standardized across all components

### v1.1.2 (2025-01-23)
**✨ Resource Quota Manager Improvements**

- ✅ **Namespace Selector**: Added namespace filter with All/Selected toggle
- ✅ **Automatic Refresh**: Resources refresh automatically after operations
- ✅ **UI Consistency**: Reorganized layout to match other pages
- ✅ **Menu Simplification**: Streamlined to Edit YAML and Delete only
- ✅ **YAML Editor Fix**: Now uses kubectl apply (Server-Side Apply)
- ✅ **Create Menu**: Changed from hover to click interaction
- ✅ **Template Namespace**: New resources use selected namespace
- ✅ **Color Consistency**: Removed colorful elements for gray-scale design

### v1.1.1 (2024-12-19)
**🐛 Bug Fixes and Improvements**

- ✅ **Fixed Resource Import**: Now accepts ANY Kubernetes resource (including CRDs and custom resources)
- ✅ **Fixed HPA YAML Viewing**: Resolved issue with HorizontalPodAutoscaler YAML viewing
- ✅ **Added Resource Aliases**: Support for common aliases (HPA, PVC, PV, SC, SA, CR, CRB, RB)
- ✅ **Dynamic Resource Discovery**: Uses Kubernetes Discovery API for automatic resource resolution
- ✅ **Enhanced RBAC**: Full access to all resources for import functionality
- ✅ **Improved Error Handling**: Better error messages and automatic fallback between API versions

### v1.1.0 (2024-12-19)
**🏗️ Unified Architecture Release**

This release introduces a major architectural improvement with enhanced security:

**Architecture Changes:**
- 🎯 **Unified Container**: Backend and Frontend integrated in a single Docker image
- 🔒 **Improved Security**: Reduced attack surface by eliminating inter-container communication
- 🚀 **Simplified Deployment**: Single service, single deployment, single port (8080)
- 📦 **Easier Management**: One image to build, version, and deploy

**Technical Improvements:**
- Backend now serves frontend static files directly
- Simplified Helm chart with unified deployment
- Reduced resource overhead (single container instead of two)
- Single ingress path configuration

**Migration Notes:**
- Docker image changed from `dkonsole/dkonsole-backend` and `dkonsole/dkonsole-frontend` to `dkonsole/dkonsole`
- Helm values updated: `image.backend` and `image.frontend` replaced with single `image` configuration
- Ingress paths simplified: single path `/` instead of separate `/api` and `/` paths

### v1.0.7 (2025-11-23)
**🔒 Security Hardening Release**

This release focuses on addressing critical security vulnerabilities and implementing enterprise-grade security measures:

**Critical Security Fixes:**
- 🛡️ **PromQL Injection Prevention**: Added strict input validation for all Prometheus queries to prevent injection attacks
- 🔐 **Cookie-Based Authentication**: Migrated from localStorage to HttpOnly cookies for JWT tokens, eliminating XSS token theft risks
- ✅ **Kubernetes Name Validation**: Implemented RFC 1123 validation for all resource names and namespaces to prevent command injection
- 🚫 **Input Sanitization**: Added comprehensive validation across all API endpoints (GetResourceYAML, DeleteResource, ScaleResource, StreamPodLogs, ExecIntoPod)

**Security Enhancements:**
- 📝 **Audit Logging**: Implemented comprehensive audit middleware logging all API requests with user attribution, timestamps, and response codes
- ⏱️ **Rate Limiting**: Added intelligent rate limiting (300 req/min per IP) to prevent DoS attacks and brute force attempts
- 🔒 **CSP Updates**: Enhanced Content Security Policy to allow Monaco Editor while maintaining security
- 🔑 **JWT Secret Enforcement**: Strict validation requiring JWT_SECRET to be set and minimum 32 characters

**Bug Fixes:**
- 🐛 Fixed table sorting functionality in WorkloadList component
- 🐛 Fixed YAML editor infinite loading issue caused by CSP restrictions
- 🐛 Fixed Secret data display (now properly decodes and shows values)
- 🐛 Removed duplicate mux initialization in main.go

**Architecture Improvements:**
- 🏗️ Unified middleware chain for consistent security policy application
- 🎯 Separated public and secure route handlers
- 📊 Enhanced error handling and validation across all handlers

**Developer Experience:**
- 📚 Improved code organization with dedicated middleware.go
- 🧹 Removed external validation dependency, using native regex for better portability
- 🔧 Better separation of concerns between authentication, auditing, and rate limiting

### v1.0.6 (2025-11-22)
- ✨ Enhanced Pod metrics with Network RX/TX and PVC usage
- 🎨 Fixed visual issues in Pod metrics tabs (removed extra border, smooth transitions)
- 📊 Improved Cluster Overview with Prometheus integration
- 📈 Added real-time node metrics table with CPU, Memory, Disk, and Network stats
- 🔧 Added cluster-wide statistics (avg CPU/Memory, network traffic, trends)
- 💫 Added fadeIn animation for smooth tab transitions
- 🎯 Conditional rendering of metrics based on data availability

### v1.0.5 (2025-11-22)
- 🔧 Improved build and release scripts
- 📝 Added build and release scripts in `scripts/` directory
- 📚 Scripts documentation available in `scripts/SCRIPTS.md`
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
