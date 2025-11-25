# Estructura de Directorios - Arquitectura Orientada al Dominio

Esta estructura organiza el código del backend en módulos orientados al dominio para facilitar el mantenimiento, testing y escalabilidad.

## Estructura Propuesta

```
backend/
├── internal/
│   ├── models/          # Tipos compartidos y estructuras de datos
│   ├── api/            # Handlers de API genéricos (API Resources, CRDs)
│   ├── k8s/            # Handlers de recursos Kubernetes (Resources, Namespaces, etc.)
│   ├── helm/           # Handlers de Helm (Releases)
│   ├── auth/           # Handlers de autenticación
│   ├── cluster/        # Gestión de clusters (GetClusters, AddCluster)
│   └── pod/            # Operaciones específicas de pods (logs, exec, events)
```

## Descripción de Módulos

### `models/`
Contiene todos los tipos compartidos (structs) que son utilizados por múltiples módulos para evitar dependencias circulares:

- **Handlers**: Estructura principal con clientes de Kubernetes
- **ClusterConfig**: Configuración de clusters
- **Namespace, Resource, DeploymentDetails**: Tipos de recursos K8s
- **APIResourceInfo, APIResourceObject**: Tipos para recursos de API descubiertos
- **HelmRelease**: Tipos para releases de Helm
- **ClusterStats**: Estadísticas agregadas del cluster
- **PrometheusQueryResult, MetricDataPoint**: Tipos para métricas de Prometheus
- **Credentials**: Tipos de autenticación
- Funciones auxiliares: `ResolveGVR()`, `NormalizeKind()`
- Variables globales: `ResourceMetaMap`, `KindAliases`

### `api/` (Próximo paso)
Handlers para recursos de API genéricos y CRDs:
- `ListAPIResources`
- `ListAPIResourceObjects`
- `GetAPIResourceYAML`
- `GetCRDs`
- `GetCRDResources`
- `GetCRDYaml`

### `k8s/` (Próximo paso)
Handlers para recursos estándar de Kubernetes:
- `GetNamespaces`
- `GetResources`
- `GetResourceYAML`
- `UpdateResourceYAML`
- `ImportResourceYAML`
- `DeleteResource`
- `ScaleResource`
- `WatchResources`
- `GetClusterStats`
- `TriggerCronJob`
- Validaciones: `validateResourceQuota`, `validateLimitRange`

### `helm/` (Próximo paso)
Handlers para operaciones de Helm:
- `GetHelmReleases`
- `DeleteHelmRelease`
- `UpgradeHelmRelease`
- `InstallHelmRelease`

### `auth/` (Próximo paso)
Handlers de autenticación (actualmente en `auth.go`):
- `LoginHandler`
- `LogoutHandler`
- `MeHandler`
- `AuthMiddleware`
- `authenticateRequest`

### `cluster/` (Próximo paso)
Gestión de múltiples clusters:
- `GetClusters`
- `AddCluster`
- Funciones auxiliares: `getClient()`, `getDynamicClient()`, `getMetricsClient()`

### `pod/` (Próximo paso)
Operaciones específicas de pods:
- `StreamPodLogs`
- `GetPodEvents`
- `ExecIntoPod`

## Beneficios de esta Estructura

1. **Separación de responsabilidades**: Cada módulo tiene un propósito claro
2. **Evita dependencias circulares**: Los tipos compartidos están en `models/`
3. **Facilita el testing**: Cada módulo puede ser testeado independientemente
4. **Mejora la mantenibilidad**: Código más organizado y fácil de navegar
5. **Escalabilidad**: Fácil agregar nuevos módulos sin afectar los existentes

## Próximos Pasos

1. ✅ Crear estructura de directorios
2. ✅ Crear paquete `models` con tipos compartidos
3. ⏳ Mover handlers a sus respectivos módulos
4. ⏳ Actualizar imports en `main.go` y otros archivos
5. ⏳ Implementar tests para cada módulo









