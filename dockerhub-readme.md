# DKonsole

Modern Kubernetes dashboard for managing cluster resources, viewing logs, executing commands in pods, and monitoring historical metrics with optional Prometheus integration.

## Image Information

- **Repository**: `dkonsole/dkonsole`
- **Base Image**: `alpine:3.22`
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
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | Yes** | (none) |
| `PROMETHEUS_URL` | Prometheus server URL for historical metrics | No | Metrics disabled |

*Not required in setup mode - configured via web interface during first access*
**Required for browser access in production (configure your ingress URL origin).**

## Setup Mode

If the authentication secret does not exist, DKonsole automatically enters Setup Mode:

1. Access the web interface
2. Complete the setup form (admin username, password, optional JWT secret, and a Kubernetes token)
3. Secret is created automatically in Kubernetes
4. Service reloads configuration without pod restart
5. Login with configured credentials

The secret name defaults to `dkonsole-auth` and can be overridden with `AUTH_SECRET_NAME`.

## Kubernetes Deployment

For production deployments, use the single manifest:

```bash
kubectl apply -f https://raw.githubusercontent.com/flaucha/DKonsole/v2.0.0/deploy/dkonsole.yaml
```

If you want an ingress, define `DKONSOLE_DOMAIN` and render the manifest:

```bash
export DKONSOLE_DOMAIN=dkonsole.example.com
bash <(curl -fsSL https://raw.githubusercontent.com/flaucha/DKonsole/v2.0.0/scripts/render-manifest.sh) \
  https://raw.githubusercontent.com/flaucha/DKonsole/v2.0.0/deploy/dkonsole.yaml \
  | kubectl apply -f -
```

## Image Layers

- **Frontend Builder**: Node.js 22 Alpine - builds React application
- **Backend Builder**: Go 1.25.8 Alpine 3.22 - compiles Go backend
- **Runtime**: Alpine 3.22 - minimal production image

## Security

- Runs as non-root user (UID 1000)
- Argon2id password hashing with secure random salt
- JWT sessions stored in HTTP-only cookies
- Security headers middleware enabled
- RBAC depends on the Kubernetes token configured during setup

## Volumes

- `/home/app/.kube/config`: Kubernetes config file for local testing (read-only recommended)
- `/home/app/data`: Optional only for non-Kubernetes/local fallback logo storage

## Health Checks

- **Liveness Probe**: `GET /healthz` - checks every 10s after 30s initial delay
- **Readiness Probe**: `GET /healthz` - checks every 5s after 5s initial delay

## Resource Requirements

Default manifest resource values:
- **CPU Request**: 100m
- **CPU Limit**: 500m
- **Memory Request**: 128Mi
- **Memory Limit**: 512Mi

## Tags

- `latest`: Points to the most recent stable release
- Version tags: `2.0.0`, `1.6.0`, `1.5.7`, `1.5.6`, etc.

See [Docker Hub tags](https://hub.docker.com/r/dkonsole/dkonsole/tags) for all available versions.

## Documentation

For installation instructions, configuration options, and detailed documentation, visit the [GitHub repository](https://github.com/flaucha/DKonsole).

## License

MIT License
