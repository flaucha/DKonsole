# Progreso de Refactorización

## ✅ Completado

### 1. Estructura de Directorios
- ✅ `internal/models/` - Tipos compartidos
- ✅ `internal/utils/` - Funciones auxiliares
- ✅ `internal/cluster/` - Gestión de clusters
- ✅ `internal/k8s/` - (Estructura creada, pendiente implementación)
- ✅ `internal/api/` - (Estructura creada, pendiente implementación)
- ✅ `internal/helm/` - (Estructura creada, pendiente implementación)
- ✅ `internal/pod/` - (Estructura creada, pendiente implementación)
- ✅ `internal/auth/` - (Estructura creada, pendiente implementación)

### 2. Paquete `models` ✅
**Archivo:** `internal/models/models.go`

Contiene todos los tipos compartidos:
- `Handlers` - Estructura principal con métodos de acceso al mutex
- `ClusterConfig`, `Namespace`, `Resource`, `DeploymentDetails`
- `APIResourceInfo`, `APIResourceObject`
- `HelmRelease`, `ClusterStats`
- Tipos de Prometheus: `PrometheusQueryResult`, `MetricDataPoint`, etc.
- `Credentials`
- Funciones: `ResolveGVR()`, `NormalizeKind()`
- Variables: `ResourceMetaMap`, `KindAliases`

### 3. Paquete `utils` ✅
**Archivo:** `internal/utils/utils.go`

Funciones auxiliares compartidas:
- `HandleError()` - Manejo de errores
- `CreateTimeoutContext()` - Contextos con timeout
- `IsSystemNamespace()` - Validación de namespaces del sistema
- `ValidateK8sName()` - Validación de nombres K8s
- `CheckQuotaLimits()`, `CheckStorageQuota()` - Validación de quotas
- `GetClientIP()` - Extracción de IP del cliente
- `AuditLog()` - Logging de auditoría

### 4. Módulo `cluster` ✅
**Archivo:** `internal/cluster/cluster.go`

Handlers implementados:
- `GetClusters()` - Lista de clusters configurados
- `AddCluster()` - Agregar nuevo cluster
- `GetClient()` - Obtener cliente Kubernetes
- `GetDynamicClient()` - Obtener cliente dinámico
- `GetMetricsClient()` - Obtener cliente de métricas

## ⏳ Pendiente

### Módulo `k8s` (Recursos Kubernetes)
**Handlers a mover:**
- `GetNamespaces()` - Línea ~332
- `GetResources()` - Línea ~361
- `GetResourceYAML()` - Línea ~1125
- `UpdateResourceYAML()` - Línea ~1358
- `ImportResourceYAML()` - Línea ~1505
- `DeleteResource()` - Línea ~1748
- `ScaleResource()` - Línea ~1840
- `WatchResources()` - Línea ~1909
- `GetClusterStats()` - Línea ~2192
- `TriggerCronJob()` - Línea ~2944
- `validateResourceQuota()` - Línea ~3025
- `validateLimitRange()` - Línea ~3143

**Dependencias:**
- Usa `models.Handlers`, `models.Resource`, `models.Namespace`, etc.
- Usa `utils.HandleError()`, `utils.CreateTimeoutContext()`, etc.
- Usa `cluster.Service` para obtener clientes

### Módulo `api` (API Resources y CRDs)
**Handlers a mover:**
- `ListAPIResources()` - Línea ~2011
- `ListAPIResourceObjects()` - Línea ~2045
- `GetAPIResourceYAML()` - Línea ~2120
- `GetCRDs()` - Línea ~2735
- `GetCRDResources()` - Línea ~2799
- `GetCRDYaml()` - Línea ~2877

**Dependencias:**
- Usa `models.Handlers`, `models.APIResourceInfo`, `models.APIResourceObject`
- Usa `utils.*`
- Usa `cluster.Service` para obtener clientes

### Módulo `helm` (Helm Releases)
**Handlers a mover:**
- `GetHelmReleases()` - Línea ~3213
- `DeleteHelmRelease()` - Línea ~3430
- `UpgradeHelmRelease()` - Línea ~3527
- `InstallHelmRelease()` - Línea ~3960

**Dependencias:**
- Usa `models.Handlers`, `models.HelmRelease`
- Usa `utils.*`
- Usa `cluster.Service` para obtener clientes

### Módulo `pod` (Operaciones de Pods)
**Handlers a mover:**
- `StreamPodLogs()` - Línea ~2247
- `GetPodEvents()` - Línea ~2325
- `ExecIntoPod()` - Línea ~2394

**Dependencias:**
- Usa `models.Handlers`
- Usa `utils.*`
- Usa `cluster.Service` para obtener clientes
- Requiere WebSocket para `ExecIntoPod`

### Módulo `auth` (Autenticación)
**Handlers a mover desde `auth.go`:**
- `LoginHandler()` - Línea ~69
- `LogoutHandler()` - Línea ~139
- `MeHandler()` - Línea ~155
- `AuthMiddleware()` - Línea ~220
- `authenticateRequest()` - Línea ~239
- `verifyPassword()` - Línea ~168

**Dependencias:**
- Usa `models.Credentials`, `models.Claims` (necesita mover Claims a models)
- JWT handling

## 📝 Notas Importantes

1. **Claims Type**: El tipo `Claims` está en `auth.go` pero se usa en múltiples lugares. Debería moverse a `models` o crear un paquete `auth/models`.

2. **HealthHandler**: Este handler es simple y puede quedarse en `main.go` o moverse a un módulo `health`.

3. **Logo Handlers**: `GetLogo()` y `UploadLogo()` pueden ir a un módulo `ui` o `config`.

4. **Prometheus Handlers**: Ya están en archivos separados (`prometheus.go`, `prometheus_pod.go`, `prometheus_cluster.go`), solo necesitan actualizar imports.

5. **Middleware**: Los middlewares en `middleware.go` pueden quedarse ahí o moverse a `internal/middleware`.

## 🔄 Próximos Pasos

1. Crear estructura base de cada módulo con `Service` struct
2. Mover handlers uno por uno, actualizando imports
3. Actualizar `main.go` para usar los nuevos servicios
4. Actualizar `handlers.go` para usar los servicios (o eliminarlo si todo se mueve)
5. Ejecutar tests y corregir errores de compilación
6. Implementar tests unitarios para cada módulo




