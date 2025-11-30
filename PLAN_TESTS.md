# Plan de Implementación de Tests - DKonsole

Este documento detalla el plan paso a paso para remediar la falta de testing en el proyecto DKonsole, basado en el análisis de `analisis_tests_unitarios.md`.

## Estado Actual

- **Cobertura Global:** ~10-15%
- **Módulos sin tests:** 9 módulos completos
- **Módulos con cobertura insuficiente:** 4 módulos

## Estrategia

Seguir las fases definidas en el análisis, priorizando seguridad y funcionalidad crítica.

---

## FASE 1: CRÍTICA - Seguridad y Permisos (Semana 1-2)

### ✅ Objetivo: Tests de seguridad y funcionalidad core

### 1. `internal/permissions` - 🚨 MÁXIMA PRIORIDAD
**Razón:** Sistema de permisos es crítico para seguridad. Sin tests, existe riesgo de escalación de privilegios.

**Tests a implementar:**
- [ ] `TestGetUserFromContext` - Extracción de usuario del contexto
- [ ] `TestHasNamespaceAccess` - Verificación de acceso a namespace
- [ ] `TestGetPermissionLevel` - Obtención de nivel de permiso
- [ ] `TestCanPerformAction` - Verificación de acciones permitidas
- [ ] `TestFilterAllowedNamespaces` - Filtrado de namespaces
- [ ] `TestGetAllowedNamespaces` - Lista de namespaces permitidos
- [ ] `TestValidateNamespaceAccess` - Validación de acceso
- [ ] `TestValidateAction` - Validación de acción
- [ ] `TestFilterResources` - Filtrado de recursos
- [ ] `TestIsAdmin` - Verificación de admin
- [ ] `TestRequireAdmin` - Requisito de admin

**Archivo:** `backend/internal/permissions/service_test.go`

### 2. `internal/auth` - Completar tests
**Razón:** Autenticación es crítica para seguridad.

**Tests faltantes:**
- [ ] `TestAuthHandlers_LoginHandler` - Handler de login
- [ ] `TestAuthHandlers_LogoutHandler` - Handler de logout
- [ ] `TestAuthHandlers_MeHandler` - Handler de usuario actual
- [ ] `TestAuthService_LoginWithLDAP` - Login con LDAP
- [ ] `TestAuthService_ChangePassword` - Cambio de contraseña
- [ ] `TestAuthSetup_SetupAuth` - Configuración de auth
- [ ] `TestAuthSetup_VerifyOrCreateAdminUser` - Verificación/creación de admin

**Archivos:**
- `backend/internal/auth/auth_test.go` (handlers)
- `backend/internal/auth/service_test.go` (agregar tests faltantes)

### 3. `internal/middleware` - Completar tests
**Razón:** Middleware de seguridad y rate limiting.

**Tests faltantes:**
- [ ] `TestCSRFMiddleware` - CSRF protection
- [ ] `TestRateLimitMiddleware` - Rate limiting
- [ ] `TestWebSocketLimiter` - Limite de conexiones WebSocket
- [ ] `TestAuditMiddleware` - Auditoría de requests

**Archivos:**
- `backend/internal/middleware/csrf_test.go`
- `backend/internal/middleware/ratelimit_test.go`
- `backend/internal/middleware/websocket_limiter_test.go`
- `backend/internal/middleware/audit_test.go`

### 4. `internal/ldap` - Tests completos
**Razón:** Autenticación enterprise, bugs pueden bloquear acceso.

**Tests a implementar:**
- [ ] `TestLDAPService_AuthenticateUser` - Autenticación LDAP
- [ ] `TestLDAPService_GetUserGroups` - Obtención de grupos
- [ ] `TestLDAPService_ValidateUserInGroup` - Validación de grupo
- [ ] `TestLDAPClient_Connect` - Conexión LDAP
- [ ] `TestLDAPClient_Close` - Cierre de conexión
- [ ] `TestLDAPRepository_GetConfig` - Obtención de configuración
- [ ] `TestLDAPRepository_SaveConfig` - Guardado de configuración

**Archivos:**
- `backend/internal/ldap/service_test.go`
- `backend/internal/ldap/client_test.go`
- `backend/internal/ldap/repository_test.go`

---

## FASE 2: ALTA PRIORIDAD - Funcionalidad Core K8s (Semana 3-4)

### ✅ Objetivo: Estabilidad de operaciones críticas de Kubernetes

### 5. `internal/k8s` - Completar tests
**Razón:** Core de operaciones Kubernetes.

**Tests faltantes:**
- [ ] `TestDeploymentService_RestartDeployment` - Reinicio de deployment
- [ ] `TestDeploymentService_ScaleDeployment` - Escalado de deployment
- [ ] `TestNamespaceService_GetNamespaces` - Listado de namespaces
- [ ] `TestNamespaceService_CreateNamespace` - Creación de namespace
- [ ] `TestNamespaceService_DeleteNamespace` - Eliminación de namespace
- [ ] `TestCronJobService_TriggerCronJob` - Trigger de cronjob
- [ ] `TestImportService_ImportFromYAML` - Importación desde YAML
- [ ] `TestWatchService_WatchResources` - Watch de recursos
- [ ] `TestClusterStatsService_GetClusterStats` - Estadísticas de cluster

**Archivos:**
- `backend/internal/k8s/deployment_service_test.go`
- `backend/internal/k8s/namespace_service_test.go`
- `backend/internal/k8s/cronjob_service_test.go`
- `backend/internal/k8s/import_service_test.go`
- `backend/internal/k8s/watch_service_test.go`
- `backend/internal/k8s/clusterstats_service_test.go`

### 6. `internal/pod` - Completar tests
**Razón:** Operaciones core de pods.

**Tests faltantes:**
- [ ] `TestPodService_GetPods` - Listado de pods
- [ ] `TestPodService_GetPodDetails` - Detalles de pod
- [ ] `TestLogService_GetPodLogs` - Logs de pod
- [ ] `TestLogService_StreamLogs` - Stream de logs
- [ ] `TestExecService_CreateExecutor` - Creación de executor

**Archivos:**
- `backend/internal/pod/service_test.go`
- `backend/internal/pod/log_service_test.go`
- `backend/internal/pod/exec_service_test.go` (completar)

### 7. `internal/api` - Implementar tests
**Razón:** Acceso dinámico a recursos de Kubernetes y CRDs.

**Tests a implementar:**
- [ ] `TestAPIService_ListAPIResources` - Listado de recursos API
- [ ] `TestAPIService_ListAPIResourceObjects` - Listado de objetos
- [ ] `TestAPIService_GetResourceYAML` - Obtención de YAML
- [ ] `TestCRDService_GetCRDs` - Listado de CRDs
- [ ] `TestCRDService_GetCRDResources` - Listado de recursos CRD

**Archivo:** `backend/internal/api/api_service_test.go`

### 8. `internal/cluster` - Implementar tests
**Razón:** Maneja la conexión multi-cluster.

**Tests a implementar:**
- [ ] `TestClusterService_GetClient` - Obtención de cliente
- [ ] `TestClusterService_GetDynamicClient` - Cliente dinámico
- [ ] `TestClusterService_GetMetricsClient` - Cliente de métricas
- [ ] `TestClusterService_GetRESTConfig` - Configuración REST

**Archivo:** `backend/internal/cluster/cluster_test.go`

---

## FASE 3: FUNCIONALIDAD PREMIUM (Semana 5-6)

### ✅ Objetivo: Features premium y observabilidad

### 9. `internal/helm` - Implementar tests
**Razón:** Gestión de Helm es funcionalidad premium.

**Tests a implementar:**
- [ ] `TestHelmInstallService_InstallHelmRelease` - Instalación de release
- [ ] `TestHelmUpgradeService_UpgradeHelmRelease` - Upgrade de release
- [ ] `TestHelmReleaseService_GetReleases` - Listado de releases
- [ ] `TestHelmReleaseService_GetReleaseDetails` - Detalles de release
- [ ] `TestHelmReleaseService_DeleteRelease` - Eliminación de release
- [ ] `TestHelmJobService_CreateHelmJob` - Creación de job
- [ ] `TestHelmJobService_GetJobStatus` - Estado de job
- [ ] `TestHelmJobService_CreateValuesConfigMap` - Creación de ConfigMap

**Archivos:**
- `backend/internal/helm/helm_install_service_test.go`
- `backend/internal/helm/helm_upgrade_service_test.go`
- `backend/internal/helm/helm_release_service_test.go`
- `backend/internal/helm/helm_job_service_test.go`

### 10. `internal/prometheus` - Implementar tests
**Razón:** Métricas son core feature premium.

**Tests a implementar:**
- [ ] `TestPrometheusService_GetDeploymentMetrics` - Métricas de deployment
- [ ] `TestPrometheusService_GetPodMetrics` - Métricas de pod
- [ ] `TestPrometheusService_GetClusterOverview` - Overview del cluster
- [ ] `TestPrometheusService_isControlPlaneNode` - Detección de control plane
- [ ] `TestPrometheusService_calculateClusterStats` - Cálculo de stats
- [ ] `TestPrometheusUtils_ParseMetricResponse` - Parsing de respuesta
- [ ] `TestPrometheusUtils_FormatMemoryValue` - Formateo de memoria

**Archivos:**
- `backend/internal/prometheus/service_test.go`
- `backend/internal/prometheus/utils_test.go`

### 11. `internal/logo` - Implementar tests
**Razón:** Seguridad importante - validación de uploads.

**Tests a implementar:**
- [ ] `TestLogoService_UploadLogo` - Upload de logo
- [ ] `TestLogoService_GetLogoPath` - Obtención de path
- [ ] `TestLogoValidator_ValidateFile` - Validación de archivo
- [ ] `TestLogoStorage_Save` - Guardado de logo
- [ ] `TestLogoStorage_Get` - Obtención de logo
- [ ] `TestLogoStorage_RemoveAll` - Eliminación de logos

**Archivos:**
- `backend/internal/logo/service_test.go`
- `backend/internal/logo/validator_test.go`
- `backend/internal/logo/storage_test.go`

### 12. `internal/settings` - Implementar tests
**Razón:** Configuración crítica pero de bajo riesgo.

**Tests a implementar:**
- [ ] `TestSettingsService_GetPrometheusURLHandler` - Obtención de URL
- [ ] `TestSettingsService_UpdatePrometheusURLHandler` - Actualización de URL

**Archivo:** `backend/internal/settings/service_test.go`

---

## FASE 4: COMPLETAR COBERTURA (Semana 7+)

### ✅ Objetivo: 80%+ cobertura global

### 13. `internal/health` - Implementar tests
**Razón:** Función simple pero crítica para monitoreo.

**Tests a implementar:**
- [ ] `TestHealthHandler` - Handler de health check

**Archivo:** `backend/internal/health/health_test.go`

### 14. Completar `internal/utils`
**Razón:** Ya tiene 67.2% cobertura, completar casos edge.

**Tests adicionales:**
- [ ] `TestJSONResponse` - Response JSON
- [ ] `TestErrorResponse` - Response de error
- [ ] `TestLogInfo/LogError/LogWarning` - Logging
- [ ] Más casos edge en validaciones

**Archivo:** `backend/internal/utils/utils_test.go` (agregar)

---

## Ajustes al build.sh

El script `build.sh` ya ejecuta tests en la línea 106:
```bash
go test -v -coverprofile=coverage.out ./...
```

**Verificaciones necesarias:**
- ✅ Los tests se ejecutan automáticamente en el build
- [ ] Verificar que la cobertura se genere correctamente
- [ ] Asegurar que todos los módulos nuevos se incluyan
- [ ] Verificar que los mocks necesarios estén disponibles

---

## Métricas de Éxito

### Objetivos por Fase

| Fase | Cobertura Objetivo | Módulos Completados |
|------|-------------------|---------------------|
| Fase 1 | 40-50% | 4 módulos críticos |
| Fase 2 | 60-70% | 8 módulos críticos |
| Fase 3 | 75-80% | 12 módulos |
| Fase 4 | 85%+ | Todos |

### KPIs
- **Cobertura de Línea:** Objetivo 85%+
- **Cobertura de Branches:** Objetivo 75%+
- **Tests Pasando:** 100%
- **Tests Performance:** Todos < 100ms (unit tests)

---

## Patrones de Testing

### Estructura de Tests
```go
// Patrón recomendado: Table-driven tests
func TestService_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        mock    func() *MockDependency
        want    OutputType
        wantErr bool
    }{
        // casos de prueba
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // setup
            // execute
            // assert
        })
    }
}
```

### Mocking
Usar interfaces para dependencias y crear mocks simples:
```go
type mockDependency struct {
    func func(...) (...)
}
```

---

## Próximos Pasos

1. ✅ Crear plan detallado
2. 🔄 Implementar tests de `internal/permissions` (en progreso)
3. ⏳ Continuar con Fase 1 en orden de prioridad
4. ⏳ Verificar que build.sh ejecute todos los tests correctamente
5. ⏳ Generar reportes de cobertura periódicamente

---

**Última actualización:** 2025-01-27
**Estado:** En progreso - Fase 1 iniciada
