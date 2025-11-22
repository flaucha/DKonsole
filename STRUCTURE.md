# DKonsole Project Structure

```
DKonsole/
├── backend/                    # Go backend application
│   ├── Dockerfile             # Backend container image
│   ├── main.go                # Application entry point
│   ├── auth.go                # Authentication handlers
│   ├── handlers.go            # API handlers
│   ├── go.mod                 # Go dependencies
│   └── go.sum                 # Go checksums
│
├── frontend/                   # React frontend application
│   ├── Dockerfile             # Frontend container image
│   ├── nginx.conf             # Nginx configuration
│   ├── src/                   # React source code
│   │   ├── components/        # React components
│   │   ├── context/           # Context providers
│   │   └── App.jsx            # Main app component
│   ├── public/                # Static assets
│   ├── package.json           # NPM dependencies
│   └── vite.config.js         # Vite configuration
│
├── helm/                       # Helm chart
│   └── dkonsole/
│       ├── Chart.yaml          # Chart metadata
│       ├── values.yaml         # Default values
│       ├── values-production.yaml.example  # Production example
│       └── templates/          # Kubernetes templates
│           ├── _helpers.tpl    # Template helpers
│           ├── deployment-backend.yaml
│           ├── deployment-frontend.yaml
│           ├── service.yaml
│           ├── ingress.yaml
│           ├── serviceaccount.yaml
│           ├── rbac.yaml       # ClusterRole & ClusterRoleBinding
│           ├── secret.yaml     # Admin credentials
│           └── pvc.yaml        # Persistent storage
│
├── README.md                   # English documentation
├── README.es.md                # Spanish documentation
├── DEPLOYMENT.md               # Deployment guide
├── migrate.sh                  # Migration script (Linux/Mac)
├── migrate.bat                 # Migration script (Windows)
├── .gitignore                  # Git ignore rules
└── LICENSE                     # Project license
```

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Web Framework**: net/http (stdlib)
- **Authentication**: JWT with Argon2 password hashing
- **K8s Client**: client-go, dynamic client
- **WebSocket**: gorilla/websocket

### Frontend
- **Framework**: React 18
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Editor**: Monaco Editor
- **Icons**: Lucide React
- **HTTP Client**: Fetch API

### Infrastructure
- **Container Runtime**: Docker
- **Orchestration**: Kubernetes
- **Package Manager**: Helm 3
- **Ingress**: Nginx Ingress Controller (recommended)
- **TLS**: cert-manager (optional)

## Key Features

### Security
- JWT-based authentication
- Argon2 password hashing
- RBAC integration
- Secure WebSocket connections
- TLS support via Ingress

### Kubernetes Integration
- ClusterRole with full permissions
- ServiceAccount for pod access
- Dynamic resource discovery
- Metrics API support
- CRD support

### UI/UX
- Modern, responsive design
- Real-time updates via WebSocket
- YAML editing with syntax highlighting
- Resource filtering and search
- Multi-cluster support

## Environment Variables

### Backend
- `JWT_SECRET`: Secret key for JWT signing (required)
- `ADMIN_USER`: Admin username (default: admin)
- `ADMIN_PASSWORD`: Argon2 password hash (required)

### Frontend
(All configuration is done through environment at build time via Vite)

## Build & Deploy Workflow

1. **Development**
   ```bash
   # Backend
   cd backend && go run main.go
   
   # Frontend
   cd frontend && npm run dev
   ```

2. **Build Images**
   ```bash
   docker build -t dkonsole-backend:latest backend/
   docker build -t dkonsole-frontend:latest frontend/
   ```

3. **Push to Registry**
   ```bash
   docker push your-registry/dkonsole-backend:latest
   docker push your-registry/dkonsole-frontend:latest
   ```

4. **Deploy with Helm**
   ```bash
   helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
   ```

## RBAC Permissions

The ServiceAccount has permissions for:

- **Cluster-scoped**: nodes, namespaces, PVs, StorageClasses, CRDs
- **Namespace-scoped**: pods, services, deployments, statefulsets, daemonsets, jobs, cronjobs, configmaps, secrets, PVCs, ingresses, network policies, roles, rolebindings, resource quotas, limit ranges

## Storage

- **PVC**: Used for logo uploads (optional)
- **Size**: 1Gi default (configurable)
- **Access Mode**: ReadWriteOnce
- **Storage Class**: Configurable

## Networking

- **Backend Port**: 8080
- **Frontend Port**: 80
- **Ingress**: Path-based routing
  - `/api` → backend service
  - `/` → frontend service

## Monitoring & Logging

- **Health Checks**: Liveness and readiness probes
- **Logs**: kubectl logs or your logging solution
- **Metrics**: Prometheus-compatible (if metrics-server is installed)

## License

MIT License - see LICENSE file for details
