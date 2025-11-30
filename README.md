# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)
![Version](https://img.shields.io/badge/version-1.3.5-green.svg)

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
- 📱 **LDAP Integration**: LDAP Integration for user authentication

## 🚀 Quick Start

### 1. Deploy with Helm

```bash
# Add the repo (if applicable) or clone
git clone https://github.com/flaucha/DKonsole.git
cd DKonsole

# Checkout the latest stable version
git checkout v1.3.5

# Configure ingress and allowedOrigins (at minimum)
vim ./helm/dkonsole/values.yaml

# Install
helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace

# After installation, access the web interface to complete the initial setup
```

## ⚙️ Configuration

The `values.yaml` file is designed to be simple. You only need to configure the essentials:

### 1. Ingress (Required for external access)
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

# Required for setup mode via ingress (CORS)
allowedOrigins: "https://dkonsole.example.com"
```

### 2. Initial Setup (Web Interface)
After deploying the Helm chart, access the web interface to complete the initial setup:

1. **Deploy the chart** (no authentication configuration needed):
   ```bash
   helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
   ```

2. **Access the web interface** via your ingress URL

3. **Complete the setup form**:
   - Enter admin username
   - Enter admin password (minimum 8 characters)
   - Optionally set a JWT secret (or leave empty for auto-generation)
   - Click "Complete Setup"

4. **Login** with the credentials you configured

The setup creates a Kubernetes secret (`{release-name}-auth`) automatically with:
- Admin username
- Argon2-hashed password
- JWT secret for session security

**Note:** The secret is created automatically by the application. You don't need to configure authentication in Helm values.

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

### 4. Security

#### Dependency Scanning

Este proyecto utiliza escaneo automatizado de vulnerabilidades:

- **Trivy**: Escaneo de contenedores y filesystems
- **govulncheck**: Análisis específico de Go
- **npm audit**: Vulnerabilidades de Node.js

##### Ejecutar manualmente

```bash
# Backend (Go)
cd backend
govulncheck ./...

# Frontend (npm)
cd frontend
npm audit --audit-level=high

# Container
docker build -t dkonsole:test .
trivy image dkonsole:test
```

##### CI/CD

El workflow `.github/workflows/security.yml` ejecuta automáticamente:
- ✅ Escaneo en cada push/PR
- ✅ Escaneo diario programado (2 AM)
- ✅ Generación de SBOM en main
- ✅ Resultados en GitHub Security tab

#### Reportar Vulnerabilidades

Si encuentras una vulnerabilidad de seguridad, por favor reporta a: security@example.com

### 5. Docker Image (Optional)
By default, it uses the official image. You can change tag or repository if needed.

```yaml
image:
  repository: dkonsole/dkonsole
  tag: "1.3.5"
```

## 🐳 Docker Image

The official image is available at:

- **Unified**: `dkonsole/dkonsole:1.3.5`

**Note:** Starting from v1.1.0, DKonsole uses a unified container architecture where the backend serves the frontend static files. This improves security by reducing the attack surface and eliminating inter-container communication.

## 📝 Changelog

### v1.3.5 (2025-11-30)
**🔧 Table Layout Improvements & Actions Fix**

This release reorganizes table columns and fixes action buttons wrapping issues.

- **Table Layout Reorganization**: Reorganized all table columns from right to left for better visual organization
  - Name column always positioned on the left with more space
  - All other columns arranged from right to left based on resource type
  - Improved column spacing and distribution across all resource types
- **Name Column Improvements**: Enhanced Name column handling for long resource names
  - Better truncation and overflow handling
  - Added tooltip to show full name on hover
  - Added 0.5cm left padding to Name column header for better alignment
- **Actions Column Fix**: Fixed action buttons wrapping to second line in all resource tables
  - Adjusted column spans for Actions based on number of buttons
  - Actions properly aligned to the right in all resource types
  - All table rows now maintain single-line layout without wrapping

### v1.3.4 (2025-11-30)
**✨ Table Enhancements & UI Improvements**

This release adds Ready column for Pods, Tag column for Deployments, and improves table alignment.

- **Pod Ready Column**: Added Ready column to Pods table showing container ready status in X/Y format (like kubectl)
  - Displays ready containers count over total containers (e.g., "2/3")
  - Column positioned between Status and Age columns
  - Sortable by ready ratio
- **Deployment Tag Column**: Added Tag column to Deployments table
  - Shows image tag extracted from container image
  - Handles SHA256 digests by showing first 8 characters
  - Falls back to "latest" if no tag specified
- **Table Alignment**: Centered all column texts in all resource tables
  - Consistent visual presentation across all resource types
  - Improved readability and professional appearance
- **Unit Tests**: Added comprehensive unit tests for new table columns

### v1.3.3 (2025-11-30)
**🐛 LDAP Admin Permissions & Logo Storage Migration**

This release fixes LDAP admin permissions and migrates logo storage to ConfigMap.

- **LDAP Admin Permissions**: Fixed LDAP admin users not being able to view cluster resources
  - LDAP admins now correctly receive `role: "admin"` in JWT when they belong to admin groups
  - LDAP admins can now view all resources (deployments, pods, etc.) and cluster overview
- **Settings Password Change**: Fixed password change section showing for LDAP admin users
  - Password change section now only visible for core admins
  - LDAP admins can access Prometheus settings but not password change
- **Logo Storage**: Migrated logo storage from PersistentVolumeClaim to Kubernetes ConfigMap
  - Logos are now stored in ConfigMap `dkonsole-logo` in base64 format
  - Simplified deployment by removing PVC requirement
- **Settings Access**: Added Settings access within Admin Area section

### v1.2.8 (2025-11-29)
**✨ Settings Management & Metrics Fixes**

This release adds settings management and fixes Prometheus metrics functionality.

- **Settings Module**: New settings management system for application configuration
  - Prometheus URL configuration via web interface
  - Password change functionality with confirmation dialog
  - Settings stored in Kubernetes ConfigMap for persistence
  - Dynamic Prometheus service updates without restart
- **Favicon Size**: Increased favicon size from 120x120 to 512x512 for better visibility
- **Password Change UX**: Improved password change flow with confirmation popup
  - Shows warning about automatic logout
  - Requires explicit confirmation before changing password
  - Automatic logout and redirect to login after password change
- **Prometheus Metrics Fix**: Fixed metrics not working after URL update
  - Added thread-safe URL updates with mutex protection
  - Prometheus service now updates dynamically when URL changes
  - ConfigMap is read at startup to load saved Prometheus URL

### v1.2.7 (2025-11-28)
**✨ Setup Mode & Auto-Reload**

This release introduces a web-based setup mode for initial configuration and automatic service reload.

- **Setup Mode**: Initial setup via web interface instead of Helm secrets
  - Automatic detection when `dkonsole-auth` secret doesn't exist
  - Web-based setup form for admin username, password, and JWT secret
  - Auto-generation of JWT secret with manual override option
  - Argon2 password hashing for secure credential storage
- **Auto-Reload After Setup**: Service automatically reloads configuration after setup completion
  - No pod restart required after initial setup
  - Seamless transition from setup mode to normal operation
- **Setup Status Check**: Frontend checks setup status on page load
  - Shows "Setup Completed" message if setup already done
  - Prevents duplicate setup attempts
  - Automatic redirect to login after successful setup
- **Security**: Passwords are now hashed using Argon2id with secure random salt generation

For the complete changelog, see [CHANGELOG.md](./CHANGELOG.md)

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

## 🏗️ Arquitectura

For detailed coding standards and contribution guidelines, please refer to [CODING_GUIDELINES.md](./CODING_GUIDELINES.md).

DKonsole utiliza una arquitectura orientada al dominio en el backend, organizando el código en módulos especializados dentro de `backend/internal/`:

```mermaid
graph TB
    subgraph "Frontend"
        UI[React UI]
    end

    subgraph "Backend - HTTP Server"
        Main[main.go<br/>Router & Middleware]
        AuthMW[Auth Middleware]
        RateLimit[Rate Limiting]
        CORS[CORS Handler]
    end

    subgraph "Backend - Services Layer"
        AuthSvc[auth.Service<br/>Login, Logout, Auth]
        ClusterSvc[cluster.Service<br/>Cluster Management]
        K8sSvc[k8s.Service<br/>Resources, Namespaces]
        ApiSvc[api.Service<br/>API Resources, CRDs]
        HelmSvc[helm.Service<br/>Helm Releases]
        PodSvc[pod.Service<br/>Logs, Exec, Events]
        PromSvc[prometheus.Service<br/>Metrics & Overview]
        LogoSvc[logo.Service<br/>Custom Branding]
        HealthSvc[health.Handler<br/>Health Checks]
    end

    subgraph "Backend - Shared"
        Models[models/<br/>Shared Types]
        Utils[utils/<br/>Utilities]
    end

    subgraph "External Systems"
        K8s[Kubernetes API]
        Prometheus[Prometheus]
        FileSystem[File System]
    end

    UI -->|HTTP Requests| Main
    Main --> AuthMW
    AuthMW --> RateLimit
    RateLimit --> CORS
    CORS --> AuthSvc
    CORS --> K8sSvc
    CORS --> ApiSvc
    CORS --> HelmSvc
    CORS --> PodSvc
    CORS --> PromSvc
    CORS --> LogoSvc
    CORS --> HealthSvc

    AuthSvc --> Models
    ClusterSvc --> Models
    K8sSvc --> Models
    K8sSvc --> ClusterSvc
    ApiSvc --> Models
    ApiSvc --> ClusterSvc
    HelmSvc --> Models
    HelmSvc --> ClusterSvc
    PodSvc --> Models
    PodSvc --> ClusterSvc
    PromSvc --> Models
    PromSvc --> ClusterSvc
    LogoSvc --> Models

    K8sSvc --> Utils
    ApiSvc --> Utils
    HelmSvc --> Utils
    PodSvc --> Utils
    AuthSvc --> Utils
    PromSvc --> Utils
    LogoSvc --> Utils

    ClusterSvc --> K8s
    K8sSvc --> K8s
    ApiSvc --> K8s
    HelmSvc --> K8s
    PodSvc --> K8s
    PodSvc --> Prometheus
    PromSvc --> Prometheus
    LogoSvc --> FileSystem

    style Main fill:#e1f5ff
    style AuthSvc fill:#fff4e1
    style ClusterSvc fill:#fff4e1
    style K8sSvc fill:#fff4e1
    style ApiSvc fill:#fff4e1
    style HelmSvc fill:#fff4e1
    style PodSvc fill:#fff4e1
    style PromSvc fill:#fff4e1
    style LogoSvc fill:#fff4e1
    style HealthSvc fill:#fff4e1
    style Models fill:#e8f5e9
    style Utils fill:#e8f5e9
    style K8s fill:#ffebee
    style Prometheus fill:#ffebee
    style FileSystem fill:#ffebee
```

### Módulos del Backend

- **`models/`**: Tipos compartidos y estructuras de datos (Handlers, ClusterConfig, Resource, etc.)
- **`utils/`**: Funciones auxiliares compartidas (manejo de errores, validaciones, contextos)
- **`auth/`**: Autenticación y autorización (JWT, Argon2, middleware)
- **`cluster/`**: Gestión de múltiples clusters Kubernetes
- **`k8s/`**: Operaciones con recursos estándar de Kubernetes (Namespaces, Resources, YAML)
- **`api/`**: Recursos de API genéricos y CRDs (Custom Resource Definitions)
- **`helm/`**: Gestión de releases de Helm
- **`pod/`**: Operaciones específicas de pods (logs, exec, events, métricas)
- **`prometheus/`**: Integración con Prometheus para métricas históricas
- **`logo/`**: Gestión de logos personalizados
- **`health/`**: Endpoints de health check (liveness/readiness)

## 🛠️ Development

To run locally:

```bash
# Backend
cd backend && go run main.go

# Frontend
cd frontend && npm run dev
```

## License

MIT License
