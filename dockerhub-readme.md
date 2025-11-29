# DKonsole

Modern Kubernetes dashboard built with AI. Provides a web interface to manage cluster resources, view logs, execute commands in pods, and monitor historical metrics with Prometheus integration.

## Image Information

- **Repository**: `dkonsole/dkonsole`
- **Base Image**: `alpine:3.19`
- **Port**: `8080`
- **User**: Non-root (UID 1000, GID 1000)
- **Health Check**: `/healthz` endpoint
- **Architecture**: Multi-stage build (Go backend + React frontend)

## Quick Run

```bash
docker run -d \
  --name dkonsole \
  -p 8080:8080 \
  -v ~/.kube/config:/home/app/.kube/config:ro \
  dkonsole/dkonsole:latest
```

Access at `http://localhost:8080`

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `JWT_SECRET` | Secret key for JWT token signing (minimum 32 characters) | No* | Auto-generated in setup mode |
| `ADMIN_USER` | Admin username | No* | Set via web interface |
| `ADMIN_PASSWORD` | Admin password (Argon2 hash format) | No* | Set via web interface |
| `AUTH_SECRET_NAME` | Kubernetes secret name for authentication | No | `dkonsole-auth` |
| `POD_NAMESPACE` | Kubernetes namespace (auto-detected from service account) | No | Auto-detected |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | No | Same-origin only |
| `PROMETHEUS_URL` | Prometheus server URL for historical metrics | No | Metrics disabled |

*Not required in setup mode - configured via web interface during first access*

## Setup Mode

If the authentication secret does not exist, DKonsole automatically enters Setup Mode:

1. Access the web interface
2. Complete the setup form (admin username, password, optional JWT secret)
3. Secret is created automatically in Kubernetes
4. Service reloads configuration without pod restart
5. Login with configured credentials

The secret name follows the pattern: `{release-name}-auth` (default: `dkonsole-auth`)

## Kubernetes Deployment

For production deployments, use the Helm chart. See the [GitHub repository](https://github.com/flaucha/DKonsole) for installation instructions and full documentation.

The Helm chart handles:
- ServiceAccount creation with RBAC permissions
- Ingress configuration
- Persistent volume for custom logos (optional)
- Resource limits and requests
- Autoscaling (optional)

## Image Layers

- **Frontend Builder**: Node.js 18 Alpine - builds React application
- **Backend Builder**: Go 1.24 Alpine - compiles Go backend
- **Runtime**: Alpine 3.19 - minimal production image

## Security

- Runs as non-root user (UID 1000)
- Argon2id password hashing with secure random salt
- JWT sessions stored in HTTP-only cookies
- Security headers middleware enabled
- RBAC support via Kubernetes ServiceAccount

## Volumes

- `/home/app/data`: Persistent storage for custom logos (optional, requires PVC)
- `/home/app/.kube/config`: Kubernetes config file (for local testing, read-only recommended)

## Health Checks

- **Liveness Probe**: `GET /healthz` - checks every 10s after 30s initial delay
- **Readiness Probe**: `GET /healthz` - checks every 5s after 5s initial delay

## Resource Requirements

Default Helm values:
- **CPU Request**: 100m
- **CPU Limit**: 500m
- **Memory Request**: 128Mi
- **Memory Limit**: 512Mi

## Tags

- `latest`: Points to the most recent stable release
- Version tags: `1.3.0`, `1.2.8`, `1.2.7`, etc.

See [Docker Hub tags](https://hub.docker.com/r/dkonsole/dkonsole/tags) for all available versions.

## Documentation

For installation instructions, configuration options, and detailed documentation, visit the [GitHub repository](https://github.com/flaucha/DKonsole).

## License

MIT License
