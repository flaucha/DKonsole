# Prompts de Mejora para DKonsole

Este archivo contiene una serie de prompts diseñados para ser ejecutados secuencialmente por un asistente de IA para mejorar el proyecto DKonsole.

> **Nota:** Se ha priorizado la refactorización sobre el testing global. Sin embargo, se recomienda encarecidamente **agregar tests unitarios a cada nuevo módulo** a medida que se extrae.

## 1. Refactorización Backend (Arquitectura)

### 1.1. Dividir el Monolito `handlers.go` - Paso 1: Estructura
> **Estado:** ✅ COMPLETADO
> **Nota:** La estructura de carpetas `backend/internal` ya existe y está poblada.

### 1.2. Dividir el Monolito `handlers.go` - Paso 2: Migración de Auth
> **Estado:** ✅ COMPLETADO
> **Nota:** La lógica de autenticación ya está en `backend/internal/auth`.

### 1.3. Dividir el Monolito `handlers.go` - Paso 3: Migración de K8s Handlers
> **Estado:** ✅ COMPLETADO
> **Nota:** La mayoría de los handlers se han movido a servicios en `backend/internal`. `handlers.go` ahora es principalmente un wrapper de compatibilidad.

## 2. Refactorización Frontend (React)


### 2.1. Modularizar `WorkloadList.jsx`
> **Objetivo:** Dividir el componente más grande del frontend para hacerlo mantenible.

```text
Actúa como un experto en React y Clean Code. El componente `frontend/src/components/WorkloadList.jsx` es demasiado grande.

1.  Analiza el componente e identifica sub-componentes lógicos (ej: Tabla de Pods, Tabla de Deployments).
2.  Extrae el código de la tabla de Pods a `frontend/src/components/workloads/PodTable.jsx`.
3.  Extrae el código de la tabla de Deployments a `frontend/src/components/workloads/DeploymentTable.jsx`.
4.  Refactoriza `WorkloadList.jsx` para usar estos nuevos componentes.
```

## 3. Infraestructura y Calidad

### 3.1. Configuración de CI/CD (GitHub Actions)
> **Objetivo:** Automatizar verificaciones ahora que el código está más ordenado.

```text
Actúa como un ingeniero DevOps. Configura un pipeline de CI/CD básico.

1.  Crea `.github/workflows/ci.yaml`.
2.  Define un job que corra en cada push a `main`:
    *   Instalar Go y Node.js.
    *   Ejecutar `go vet ./...` y `go test ./...` (ahora deberían correr los tests de los nuevos paquetes).
    *   Ejecutar `npm run lint` en el frontend.
```

## 4. Documentación

### 4.1. Diagrama de Arquitectura y Guía de Contribución
> **Objetivo:** Documentar la *nueva* arquitectura refactorizada.

```text
Actúa como un Technical Writer.

1.  Crea `CONTRIBUTING.md` con pautas de desarrollo.
2.  Genera una descripción Mermaid de la **nueva** arquitectura del sistema (con los paquetes `internal/auth`, `internal/k8s`, etc.) para el README.
```

## 5. Mejoras de Código y Seguridad (Deep Dive)

### 5.1. Uso Correcto de Contextos (Performance/Cancelación)
> **Objetivo:** Evitar "goroutine leaks" y respetar los timeouts de las peticiones HTTP.

```text
Actúa como un experto en Go. He notado que en `backend/handlers.go` y otros archivos se usa mucho `context.TODO()` o `context.Background()` para las llamadas a Kubernetes.

1.  Reemplaza todos los `context.TODO()` y `context.Background()` dentro de los http.Handlers por `r.Context()`.
2.  Esto asegurará que si el cliente cancela la petición, la llamada a Kubernetes también se cancele.
3.  Verifica que `backend/internal/api/api.go` también use el contexto correcto.
```

### 5.2. Optimización de Listados (Performance)
> **Objetivo:** Evitar que el servidor explote de memoria en clusters grandes.

```text
Actúa como un experto en Kubernetes y Go. Actualmente, las llamadas `List` (ej: `client.CoreV1().Pods(...).List(...)`) traen TODOS los recursos sin paginación ni límites.

1.  Modifica los handlers de listado (en `backend/internal/k8s` si ya refactorizaste, o en `handlers.go`) para aceptar parámetros de paginación (`limit` y `continue`) desde el query string.
2.  Pasa estos parámetros a `metav1.ListOptions`.
3.  Si no se provee límite, establece un límite por defecto razonable (ej: 500) para proteger la memoria.
```

## 6. Seguridad y Concurrencia (Second Deep Dive)

### 6.1. Seguridad en Prometheus (TLS/Timeouts)
> **Objetivo:** Endurecer la integración con Prometheus.

```text
Actúa como un experto en Seguridad y Go. Analiza `backend/prometheus.go`.

1.  La función `createSecureHTTPClient` permite saltarse la verificación TLS (`InsecureSkipVerify`) mediante una variable de entorno. Asegúrate de que esto loguee una ADVERTENCIA CLARA al inicio si está activado.
2.  El timeout del cliente HTTP es fijo (30s). Hazlo configurable vía variable de entorno `PROMETHEUS_TIMEOUT` con un default de 30s.
3.  Revisa `validatePromQLParam`. Aunque usa regex, asegúrate de que se aplique a TODOS los parámetros que se interpolan en las queries PromQL en `GetPrometheusMetrics`.
```

### 6.2. Revisión de Concurrencia en Mapas
> **Objetivo:** Prevenir "fatal error: concurrent map writes".

```text
Actúa como un experto en Go.
1.  Revisa el struct `Handlers` en `backend/main.go`. Tiene mapas como `Clients`, `Dynamics`, etc.
2.  Si estos mapas se modifican *después* de que arranca el servidor (ej: al añadir un cluster dinámicamente), necesitas protegerlos con un `sync.RWMutex`.
3.  Si solo se leen, está bien. Pero si `AddCluster` modifica estos mapas, DEBES implementar un Mutex para evitar pánicos por condiciones de carrera.
```

```
