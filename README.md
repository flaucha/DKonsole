# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)
![Version](https://img.shields.io/badge/version-1.4.1-green.svg)

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
git checkout v1.4.1

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
    - host: dkonsole.lan
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: dkonsole-tls
      hosts:
        - dkonsole.lan

# Required for setup mode via ingress (CORS)
allowedOrigins: "https://dkonsole.lan"
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
  tag: "1.4.1"
```

## 🐳 Docker Image

The official image is available at:

- **Unified**: `dkonsole/dkonsole:1.4.1`

**Note:** Starting from v1.1.0, DKonsole uses a unified container architecture where the backend serves the frontend static files. This improves security by reducing the attack surface and eliminating inter-container communication.

## 📝 Changelog

### v1.4.1 (2025-12-03)
**🐛 Bug Fixes & UX Improvements**

This release addresses critical bugs and enhances user experience with faster updates and improved navigation.

- **Namespace Filter "All"**: Fixed bug where selecting "All" namespaces in Resource Quotas/Limit Ranges view showed no data
- **Faster Refresh Rates**: Reduced refetch intervals from 10 seconds to 2 seconds for more responsive UI updates
- **Quotas/Limits Menu**: Consolidated Resource Quotas and Limit Ranges into a single menu item under Admin Area
- **Reorderable AGE Column**: Made the AGE column reorderable in workload lists

### v1.4.0 (2025-12-01)
**✨ Persistent column layouts & documentation refresh**

This release keeps user-defined column order sticky and aligns all docs with the real deployment defaults.

- **Column Layout Persistence**: Workload and Namespace tables now expose drag handles and remember per-user column order using cookies + localStorage.
- **Namespace Defaults**: The dashboard shares a `DEFAULT_NAMESPACE` constant so selectors, quota dialogs, and other flows open the `dkonsole` namespace when no preference is stored.
- **Docs & Helm Samples**: README, Helm values, and dockerhub docs all reference `dkonsole.lan`/`allowedOrigins` to match the deployed environment, alongside a refreshed `docs/iacf` AI automation pack.

### v1.3.6 (2025-11-30)
**✨ Dynamic Column Distribution & Layout Improvements**

This release implements intelligent column sizing and improves table layout responsiveness.

- **Dynamic Column Distribution**: Implemented intelligent column sizing that automatically adjusts based on available space
  - Columns now distribute equitably with minimum space constraints
  - Each column has a minimum width and shares remaining space proportionally
  - Better space utilization across all table types
- **Table Column System**: Migrated from fixed 12-column grid to dynamic CSS Grid with flexible sizing
  - Name column maintains minimum width (200px) and grows with available space
  - All columns use `minmax()` for responsive sizing with minimum constraints
  - Improved column alignment and spacing across all resource types
- **Row Height Optimization**: Made table rows thinner while maintaining readability
  - Minimum height to accommodate 2 lines of AGE text
  - Better vertical space utilization
- **Fixes**: Fixed Provisioner column width, Ports display, Access Mode alignment, and prevented column title wrapping

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
