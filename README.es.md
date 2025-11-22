# DKonsole - Dashboard de Kubernetes

![License](https://img.shields.io/badge/license-MIT-blue.svg)

Un dashboard moderno y liviano para la gestión y monitore human de clusters de Kubernetes. DKonsole proporciona una interfaz intuitiva para administrar deployments, services, pods y otros recursos de Kubernetes.

## Características

- 🚀 **Gestión de Recursos**: Ver, crear, editar y eliminar recursos de Kubernetes
- 📊 **Vista General del Cluster**: Métricas y estadísticas en tiempo real
- 🔍 **Explorador de APIs**: Descubrir e interactuar con las APIs de Kubernetes
- 📦 **Gestión de Namespaces**: Administrar namespaces, resource quotas y limit ranges
- 🔐 **Autenticación Segura**: Autenticación basada en JWT con hashing Argon2
- 📝 **Editor YAML**: Editor Monaco integrado para manifiestos YAML
- 📈 **Escalado de Recursos**: Escalar deployments directamente desde la UI
- 🔌 **Exec en Pods**: Ejecutar comandos en pods vía WebSocket
- 📜 **Streaming de Logs**: Visualización de logs de pods en tiempo real

## Prerequisitos

- **Cluster Kubernetes**: v1.19+ recomendado
- **Helm**: v3.0+
- **kubectl**: Configurado para acceder a tu cluster
- **Metrics Server** (opcional): Para métricas de pods/nodos
- **Ingress Controller** (opcional): Para acceso externo (ej: nginx-ingress)
- **cert-manager** (opcional): Para certificados TLS automáticos

## Inicio Rápido

### 1. Clonar el Repositorio

```bash
git clone https://github.com/tuusuario/DKonsole.git
cd DKonsole
```

### 2. Generar Credenciales de Administrador

Genera un hash seguro de contraseña usando Argon2:

```bash
# Instalar argon2 si no está disponible
# Ubuntu/Debian: apt-get install argon2
# macOS: brew install argon2

# Generar hash
echo -n "tu-contraseña-segura" | argon2 $(openssl rand -base64 16) -id -t 3 -m 12 -p 1 -l 32 -e
```

Guarda la salida (comienza con `$argon2id$...`)

### 3. Generar Secreto JWT

```bash
openssl rand -base64 32
```

### 4. Crear archivo de valores

Crea un archivo `mis-valores.yaml`:

```yaml
admin:
  username: admin
  passwordHash: "$argon2id$v=19$m=4096,t=3,p=1$<tu-hash-aqui>"

jwtSecret: "<tu-secreto-jwt-aqui>"

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: dkonsole.tudominio.com
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
        - dkonsole.tudominio.com

image:
  backend:
    repository: tu-registry/dkonsole-backend
    tag: "latest"
  frontend:
    repository: tu-registry/dkonsole-frontend
    tag: "latest"
```

### 5. Instalar con Helm

```bash
# Crear namespace
kubectl create namespace dkonsole

# Instalar chart
helm install dkonsole ./helm/dkonsole \
  --namespace dkonsole \
  --values mis-valores.yaml
```

### 6. Acceder al Dashboard

```bash
# Si usas Ingress
open https://dkonsole.tudominio.com

# O port-forward para acceso local
kubectl port-forward -n dkonsole svc/dkonsole-frontend 8080:80
open http://localhost:8080
```

Inicia sesión con el usuario y contraseña que configuraste.

## Construir Imágenes Docker

### Backend

```bash
cd backend
docker build -t tu-registry/dkonsole-backend:latest .
docker push tu-registry/dkonsole-backend:latest
```

### Frontend

```bash
cd frontend
docker build -t tu-registry/dkonsole-frontend:latest .
docker push tu-registry/dkonsole-frontend:latest
```

## Configuración

### Permisos RBAC

DKonsole requiere permisos a nivel de cluster para gestionar recursos. El chart de Helm crea automáticamente:

- **ServiceAccount**: `dkonsole`
- **ClusterRole**: Con permisos para todos los recursos gestionados
- **ClusterRoleBinding**: Vincula el rol a la cuenta de servicio

Para personalizar permisos, edita `values.yaml` bajo `rbac.clusterResources` y `rbac.namespacedResources`.

### Persistencia

Las cargas de logos se persisten usando un PersistentVolumeClaim:

```yaml
persistence:
  enabled: true
  storageClass: "standard"  # Tu storage class
  size: 1Gi
```

### Límites de Recursos

Ajusta los requests y limits de recursos en `values.yaml`:

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

### Configuración de Ingress

Para diferentes controladores de ingress, ajusta las anotaciones:

```yaml
ingress:
  annotations:
    # Para Traefik
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    
    # Para ALB (AWS)
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
```

## Consideraciones de Seguridad

1. **Cambia las Credenciales por Defecto**: Siempre usa contraseñas seguras y únicas
2. **Usa HTTPS**: Habilita TLS para deployments en producción
3. **Rota Secretos**: Rota regularmente los secretos JWT y contraseñas
4. **RBAC**: Revisa y restringe permisos según sea necesario
5. **Políticas de Red**: Considera implementar network policies para restringir tráfico
6. **Seguridad de Pods**: El chart usa security contexts restrictivos por defecto

## Actualización

```bash
helm upgrade dkonsole ./helm/dkonsole \
  --namespace dkonsole \
  --values mis-valores.yaml
```

## Desinstalación

```bash
helm uninstall dkonsole --namespace dkonsole
kubectl delete namespace dkonsole
```

## Resolución de Problemas

### Pods no arrancan

```bash
# Verificar estado de pods
kubectl get pods -n dkonsole

# Ver logs
kubectl logs -n dkonsole deployment/dkonsole-backend
kubectl logs -n dkonsole deployment/dkonsole-frontend
```

### Problemas de autenticación

Verifica que el secret se creó correctamente:

```bash
kubectl get secret -n dkonsole dkonsole-auth -o yaml
```

### Permisos RBAC

Verifica que el ServiceAccount tiene permisos apropiados:

```bash
kubectl auth can-i --as=system:serviceaccount:dkonsole:dkonsole get pods --all-namespaces
```

## Desarrollo

### Desarrollo Local

```bash
# Backend
cd backend
go run main.go

# Frontend
cd frontend
npm install
npm run dev
```

### Variables de Entorno

El backend requiere:
- `JWT_SECRET`: Secreto para firmar JWTs
- `ADMIN_USER`: Usuario administrador
- `ADMIN_PASSWORD`: Hash Argon2 de la contraseña

## Contribuciones

¡Las contribuciones son bienvenidas! Por favor, siéntete libre de enviar un Pull Request.

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - consulta el archivo LICENSE para más detalles.

## Soporte

Para problemas y preguntas:
- GitHub Issues: https://github.com/tuusuario/DKonsole/issues
- Discusiones: https://github.com/tuusuario/DKonsole/discussions
