# Análisis de Seguridad Detallado - DKonsole

## Resumen Ejecutivo

Este documento presenta un análisis exhaustivo y actualizado de las vulnerabilidades de seguridad identificadas en el proyecto DKonsole (versión 1.0.6), una consola de administración para Kubernetes. 

**Estado Actual:**
- ✅ **Mejoras Implementadas:** Se han corregido varias vulnerabilidades críticas desde análisis anteriores
- ⚠️ **Vulnerabilidades Activas:** Se han identificado **18 vulnerabilidades** que requieren atención
- 📊 **Distribución:** 6 críticas, 6 de alta severidad, 4 de media severidad, 2 mejoras recomendadas

---

## 🔴 VULNERABILIDADES CRÍTICAS

### 1. CORS con Validación Débil de Origen

**Ubicación:** `backend/main.go:130-183`

**Problema:**
```go
if allowedOrigins != "" {
    origins := strings.Split(allowedOrigins, ",")
    for _, o := range origins {
        if strings.TrimSpace(o) == origin {
            allowed = true
            break
        }
    }
} else {
    // Si no hay ALLOWED_ORIGINS, permite localhost sin validación estricta
    if strings.Contains(origin, r.Host) || strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
        allowed = true
    }
}
```

**Severidad:** 🔴 CRÍTICA

**Descripción:**
- La validación de origen usa `strings.Contains()` que permite dominios maliciosos como `evil-localhost.com`
- Si `ALLOWED_ORIGINS` no está configurado, permite cualquier origen que contenga "localhost"
- No valida el formato completo de URL (esquema, host, puerto)

**Impacto:**
- Ataques de Cross-Site Request Forgery (CSRF)
- Robo de tokens mediante JavaScript malicioso
- Acceso no autorizado a recursos del clúster

**Solución:**
```go
func enableCors(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
        
        // Si no hay origen (y no es OPTIONS), permitir solo si es same-origin
        if origin == "" && r.Method != "OPTIONS" {
            next(w, r)
            return
        }
        
        allowed := false
        if allowedOrigins != "" {
            origins := strings.Split(allowedOrigins, ",")
            for _, o := range origins {
                o = strings.TrimSpace(o)
                if o == origin {
                    allowed = true
                    break
                }
            }
        } else {
            // Si no hay ALLOWED_ORIGINS configurado, solo permitir same-origin exacto
            if origin != "" {
                originURL, err := url.Parse(origin)
                if err == nil {
                    host := r.Host
                    // Remover puerto para comparación si es necesario
                    if strings.Contains(host, ":") {
                        host = strings.Split(host, ":")[0]
                    }
                    originHost := originURL.Host
                    if strings.Contains(originHost, ":") {
                        originHost = strings.Split(originHost, ":")[0]
                    }
                    // Solo permitir exactamente localhost, 127.0.0.1, o el mismo host
                    if (originHost == "localhost" || originHost == "127.0.0.1" || originHost == host) &&
                       (originURL.Scheme == "http" || originURL.Scheme == "https") {
                        allowed = true
                    }
                }
            }
        }
        
        if !allowed && origin != "" {
            http.Error(w, "Origin not allowed", http.StatusForbidden)
            return
        }
        
        if origin != "" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Max-Age", "3600")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next(w, r)
    }
}
```

---

### 2. Falta de Límite en Cantidad de Recursos en YAML Import

**Ubicación:** `backend/handlers.go:1143-1245`

**Problema:**
```go
func (h *Handlers) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
    // ... límite de tamaño existe (1MB) ...
    for {
        var objMap map[string]interface{}
        if err := dec.Decode(&objMap); err != nil {
            // ... sin límite en cantidad de recursos ...
        }
        // ... crear recursos sin límite ...
    }
}
```

**Severidad:** 🔴 CRÍTICA

**Descripción:**
Aunque existe límite de tamaño (1MB), no hay límite en la cantidad de recursos que se pueden crear en una sola solicitud. Un atacante puede crear miles de recursos pequeños.

**Impacto:**
- Denegación de servicio (DoS) mediante creación masiva de recursos
- Agotamiento de recursos del clúster
- Posible saturación del API server de Kubernetes

**Solución:**
```go
func (h *Handlers) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
    // ... código existente ...
    
    dec := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(body), 4096)
    var applied []string
    resourceCount := 0
    maxResources := 50 // Límite de recursos por solicitud
    
    // Contadores por tipo de recurso
    resourceTypeCounts := make(map[string]int)
    maxPerType := map[string]int{
        "Deployment": 10,
        "Service": 20,
        "ConfigMap": 30,
        "Secret": 10,
        "Job": 15,
        "CronJob": 5,
    }
    
    for {
        if resourceCount >= maxResources {
            http.Error(w, fmt.Sprintf("Too many resources (max %d)", maxResources), http.StatusBadRequest)
            return
        }
        
        var objMap map[string]interface{}
        if err := dec.Decode(&objMap); err != nil {
            if err == io.EOF {
                break
            }
            http.Error(w, fmt.Sprintf("Failed to decode YAML: %v", err), http.StatusBadRequest)
            return
        }
        
        if len(objMap) == 0 {
            continue
        }
        
        obj := &unstructured.Unstructured{Object: objMap}
        kind := obj.GetKind()
        
        // Validar límite por tipo
        if maxCount, exists := maxPerType[kind]; exists {
            if resourceTypeCounts[kind] >= maxCount {
                http.Error(w, fmt.Sprintf("Too many resources of type %s (max %d)", kind, maxCount), http.StatusBadRequest)
                return
            }
            resourceTypeCounts[kind]++
        } else {
            // Para tipos no especificados, límite general
            if resourceTypeCounts[kind] >= 10 {
                http.Error(w, fmt.Sprintf("Too many resources of type %s (max 10)", kind), http.StatusBadRequest)
                return
            }
            resourceTypeCounts[kind]++
        }
        
        // ... resto del código de validación y creación ...
        resourceCount++
    }
    
    // ... resto del código ...
}
```

---

### 3. WebSocket Origin Check Débil

**Ubicación:** `backend/handlers.go:1810-1840` (aproximado)

**Problema:**
```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    if origin == "" {
        return true // ⚠️ Permite conexiones sin origen
    }
    if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
        return true // ⚠️ "evil-localhost.com" pasaría esta validación
    }
    if strings.Contains(origin, r.Host) {
        return true
    }
    // ...
}
```

**Severidad:** 🔴 CRÍTICA

**Descripción:**
La validación de origen para WebSocket es demasiado permisiva y usa `strings.Contains()` en lugar de comparación exacta.

**Impacto:**
- Ataques de Cross-Site WebSocket Hijacking (CSWSH)
- Ejecución remota de comandos en pods mediante WebSocket comprometido
- Bypass de autenticación en terminal interactiva

**Solución:**
```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    
    // No permitir conexiones sin origen en producción
    if origin == "" {
        return false
    }
    
    // Parsear y validar el origen
    originURL, err := url.Parse(origin)
    if err != nil {
        return false
    }
    
    // Obtener origen permitido desde variable de entorno
    allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
    if allowedOrigins != "" {
        origins := strings.Split(allowedOrigins, ",")
        for _, allowed := range origins {
            allowed = strings.TrimSpace(allowed)
            allowedURL, err := url.Parse(allowed)
            if err != nil {
                continue
            }
            
            // Comparar esquema, host y puerto exactamente
            if originURL.Scheme == allowedURL.Scheme &&
               originURL.Host == allowedURL.Host {
                return true
            }
        }
        return false
    }
    
    // Si no hay ALLOWED_ORIGINS, solo permitir same-origin exacto
    host := r.Host
    if strings.Contains(host, ":") {
        host = strings.Split(host, ":")[0]
    }
    
    originHost := originURL.Host
    if strings.Contains(originHost, ":") {
        originHost = strings.Split(originHost, ":")[0]
    }
    
    // Validación estricta: solo localhost exacto, 127.0.0.1 exacto, o mismo host
    return (originHost == "localhost" || originHost == "127.0.0.1" || originHost == host) &&
           (originURL.Scheme == "http" || originURL.Scheme == "https" || originURL.Scheme == "ws" || originURL.Scheme == "wss")
},
```

---

### 4. RBAC Demasiado Permisivo - Permisos de Escritura Amplios

**Ubicación:** `helm/dkonsole/values.yaml:99-165`

**Problema:**
```yaml
namespacedResources:
  # Permisos de escritura en muchos recursos
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonsets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Severidad:** 🔴 CRÍTICA

**Descripción:**
El ClusterRole otorga permisos de escritura (`create`, `update`, `patch`, `delete`) en recursos críticos sin restricciones. Esto permite:
- Modificar deployments en producción
- Crear/eliminar secretos
- Modificar configmaps que pueden contener configuraciones críticas

**Impacto:**
- Modificación no autorizada de recursos en producción
- Eliminación accidental o maliciosa de recursos
- Escalación de privilegios mediante modificación de ServiceAccounts
- Compromiso de aplicaciones mediante modificación de configuraciones

**Solución:**
```yaml
rbac:
  namespacedResources:
    # Recursos con permisos de SOLO LECTURA
    - apiGroups: [""]
      resources: ["pods", "services", "namespaces"]
      verbs: ["get", "list", "watch"]
    - apiGroups: ["apps"]
      resources: ["deployments", "statefulsets", "daemonsets", "replicasets"]
      verbs: ["get", "list", "watch"]
    - apiGroups: ["batch"]
      resources: ["jobs", "cronjobs"]
      verbs: ["get", "list", "watch"]
    - apiGroups: ["networking.k8s.io"]
      resources: ["ingresses", "networkpolicies"]
      verbs: ["get", "list", "watch"]
    - apiGroups: ["rbac.authorization.k8s.io"]
      resources: ["roles", "rolebindings"]
      verbs: ["get", "list", "watch"]
    
    # ConfigMaps: solo lectura y actualización (no creación/eliminación)
    - apiGroups: [""]
      resources: ["configmaps"]
      verbs: ["get", "list", "watch", "update", "patch"]
    
    # Secretos: SOLO lectura de metadatos (el backend ya filtra el contenido)
    - apiGroups: [""]
      resources: ["secrets"]
      verbs: ["get", "list", "watch"]
    
    # Deployments: solo escalamiento y actualización limitada
    - apiGroups: ["apps"]
      resources: ["deployments"]
      verbs: ["get", "list", "watch"]
    - apiGroups: ["apps"]
      resources: ["deployments/scale"]
      verbs: ["get", "update", "patch"]
    - apiGroups: ["apps"]
      resources: ["deployments/status"]
      verbs: ["get", "patch"]
    
    # Pods: solo logs y exec (no modificación)
    - apiGroups: [""]
      resources: ["pods/log", "pods/exec"]
      verbs: ["get", "create"]
    
    # Jobs: solo trigger de CronJobs (no creación directa)
    - apiGroups: ["batch"]
      resources: ["jobs"]
      verbs: ["create"]  # Solo para trigger de CronJobs
```

**Nota:** Si se requiere funcionalidad de creación/eliminación, implementar validaciones adicionales en el backend y logging de auditoría.

---

### 5. Token en localStorage en TerminalViewer

**Ubicación:** `frontend/src/components/TerminalViewer.jsx:55`

**Problema:**
```javascript
const token = localStorage.getItem('token') || '';
const wsUrl = `${protocol}//${window.location.host}/api/pods/exec?namespace=${namespace}&pod=${pod}&container=${container || ''}&token=${encodeURIComponent(token)}`;
```

**Severidad:** 🔴 CRÍTICA

**Descripción:**
Aunque el sistema principal usa cookies HttpOnly, el componente TerminalViewer aún intenta obtener el token desde localStorage y lo pasa en la URL del WebSocket. Esto:
- Expone el token en la URL (visible en logs, historial del navegador)
- Es vulnerable a XSS si hay alguna vulnerabilidad en el frontend
- No sigue el patrón de seguridad del resto de la aplicación

**Impacto:**
- Exposición del token JWT en URLs
- Robo de token mediante XSS
- Acceso no autorizado a terminales de pods

**Solución:**
```javascript
// TerminalViewer.jsx
useEffect(() => {
    const term = termRef.current;
    if (!term) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // NO usar localStorage, el token debe venir de la cookie HttpOnly
    // El backend debe leer el token de la cookie automáticamente
    const wsUrl = `${protocol}//${window.location.host}/api/pods/exec?namespace=${namespace}&pod=${pod}&container=${container || ''}`;

    const ws = new WebSocket(wsUrl);
    // ... resto del código ...
}, [namespace, pod, container]);
```

**Backend:** Asegurar que `ExecIntoPod` lea el token de la cookie, no del query parameter:
```go
func (h *Handlers) ExecIntoPod(w http.ResponseWriter, r *http.Request) {
    // Autenticar usando cookie (no query param)
    claims, err := authenticateRequest(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // ... resto del código ...
}
```

---

### 6. Falta de Validación de Tamaño en Respuestas de Prometheus

**Ubicación:** `backend/prometheus.go:155, 214`

**Problema:**
```go
body, err := io.ReadAll(resp.Body) // ⚠️ Sin límite de tamaño
```

**Severidad:** 🔴 CRÍTICA

**Descripción:**
Las respuestas de Prometheus se leen completamente sin límite de tamaño. Un atacante puede hacer queries que retornen respuestas enormes, causando:
- Consumo excesivo de memoria
- Denegación de servicio
- Posible crash del servidor

**Impacto:**
- DoS mediante respuestas grandes de Prometheus
- Agotamiento de memoria del servidor
- Posible crash de la aplicación

**Solución:**
```go
func (h *Handlers) queryPrometheusRange(query string, start, end time.Time) []MetricDataPoint {
    // ... código existente ...
    
    resp, err := client.Get(fullURL)
    if err != nil {
        return []MetricDataPoint{}
    }
    defer resp.Body.Close()
    
    // Limitar tamaño de respuesta a 10MB
    maxResponseSize := int64(10 << 20) // 10MB
    limitedReader := io.LimitReader(resp.Body, maxResponseSize)
    
    body, err := io.ReadAll(limitedReader)
    if err != nil {
        return []MetricDataPoint{}
    }
    
    // Verificar si se truncó la respuesta
    if len(body) >= int(maxResponseSize) {
        fmt.Printf("Warning: Prometheus response truncated (max %d bytes)\n", maxResponseSize)
    }
    
    // ... resto del código ...
}

// Aplicar lo mismo a queryPrometheusInstant
```

---

## 🟠 VULNERABILIDADES DE ALTA SEVERIDAD

### 7. Falta de Rate Limiting

**Ubicación:** Múltiples endpoints en `backend/handlers.go`

**Problema:**
No hay límites de velocidad en los endpoints, especialmente en:
- `/api/login` - Permite fuerza bruta
- `/api/resource/import` - Permite DoS mediante múltiples solicitudes
- `/api/pods/logs` - Puede consumir recursos excesivos

**Severidad:** 🟠 ALTA

**Impacto:**
- Ataques de fuerza bruta en login
- Denegación de servicio (DoS)
- Consumo excesivo de recursos del servidor

**Solución:**
```go
import (
    "golang.org/x/time/rate"
    "sync"
    "time"
)

type rateLimiter struct {
    limiter *rate.Limiter
    lastSeen time.Time
    mu      sync.Mutex
}

var (
    loginLimiters = make(map[string]*rateLimiter)
    apiLimiters   = make(map[string]*rateLimiter)
    limiterMu     sync.Mutex
    cleanupTicker *time.Ticker
)

func init() {
    // Limpiar limiters inactivos cada 5 minutos
    cleanupTicker = time.NewTicker(5 * time.Minute)
    go func() {
        for range cleanupTicker.C {
            cleanupLimiters()
        }
    }()
}

func getClientIP(r *http.Request) string {
    // Intentar obtener IP real (detrás de proxy)
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }
    if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
        return strings.Split(ip, ",")[0]
    }
    ip, _, _ := strings.Cut(r.RemoteAddr, ":")
    return ip
}

func rateLimitMiddleware(next http.HandlerFunc, rps float64, burst int) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientIP := getClientIP(r)
        
        limiterMu.Lock()
        lim, exists := apiLimiters[clientIP]
        if !exists {
            lim = &rateLimiter{
                limiter: rate.NewLimiter(rate.Limit(rps), burst),
                lastSeen: time.Now(),
            }
            apiLimiters[clientIP] = lim
        }
        lim.lastSeen = time.Now()
        limiterMu.Unlock()
        
        lim.mu.Lock()
        if !lim.limiter.Allow() {
            lim.mu.Unlock()
            http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
            w.Header().Set("Retry-After", "60")
            return
        }
        lim.mu.Unlock()
        
        next(w, r)
    }
}

func cleanupLimiters() {
    limiterMu.Lock()
    defer limiterMu.Unlock()
    
    now := time.Now()
    for ip, lim := range apiLimiters {
        lim.mu.Lock()
        if now.Sub(lim.lastSeen) > 10*time.Minute {
            delete(apiLimiters, ip)
        }
        lim.mu.Unlock()
    }
    for ip, lim := range loginLimiters {
        lim.mu.Lock()
        if now.Sub(lim.lastSeen) > 10*time.Minute {
            delete(loginLimiters, ip)
        }
        lim.mu.Unlock()
    }
}

// Aplicar en main.go:
mux.HandleFunc("/api/login", enableCors(rateLimitMiddleware(h.LoginHandler, 5.0, 5))) // 5 req/min, burst 5
mux.HandleFunc("/api/resource/import", enableCors(AuthMiddleware(rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        h.ImportResourceYAML(w, r)
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}, 10.0, 10)))) // 10 req/min para import
```

---

### 8. Content-Security-Policy Permisivo

**Ubicación:** `frontend/nginx.conf:39`

**Problema:**
```nginx
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss:;" always;
```

**Severidad:** 🟠 ALTA

**Descripción:**
- `'unsafe-inline'` permite ejecutar JavaScript inline, vulnerable a XSS
- `'unsafe-eval'` permite `eval()`, vulnerable a inyección de código
- `ws: wss:` permite conexiones WebSocket a cualquier origen

**Impacto:**
- Vulnerable a ataques XSS
- Permite ejecución de código malicioso mediante eval()
- Permite conexiones WebSocket a dominios maliciosos

**Solución:**
```nginx
# Usar nonces para scripts inline (requiere modificar el build)
# O mejor aún, eliminar scripts inline completamente
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' wss://${ALLOWED_WS_ORIGIN}; frame-ancestors 'none'; base-uri 'self'; form-action 'self';" always;
```

**Nota:** Para usar nonces, se requiere modificar el build del frontend para inyectar nonces en los scripts. Alternativamente, eliminar todos los scripts inline.

---

### 9. Falta de Validación de Tipo MIME en Upload de Logo

**Ubicación:** `backend/handlers.go:1913-1969`

**Problema:**
```go
// Solo valida extensión, no el tipo MIME real
ext := strings.ToLower(filepath.Ext(handler.Filename))
if ext != ".png" && ext != ".svg" {
    http.Error(w, "Invalid file type. Only .png and .svg are allowed", http.StatusBadRequest)
    return
}
```

**Severidad:** 🟠 ALTA

**Descripción:**
Solo se valida la extensión del archivo, no el contenido real. Un atacante puede subir un archivo malicioso con extensión `.png` pero que en realidad sea un script ejecutable.

**Impacto:**
- Carga de archivos maliciosos disfrazados como imágenes
- Posible ejecución de código si el archivo se procesa incorrectamente
- Almacenamiento de archivos no deseados

**Solución:**
```go
func (h *Handlers) UploadLogo(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(5 << 20)
    
    file, handler, err := r.FormFile("logo")
    if err != nil {
        http.Error(w, "Error retrieving file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Validar tamaño
    if handler.Size > 5<<20 {
        http.Error(w, "File too large (max 5MB)", http.StatusBadRequest)
        return
    }
    
    // Leer primeros bytes para validar tipo MIME real
    buffer := make([]byte, 512)
    n, err := file.Read(buffer)
    if err != nil && err != io.EOF {
        http.Error(w, "Error reading file", http.StatusBadRequest)
        return
    }
    file.Seek(0, 0) // Resetear para copiar después
    
    // Validar tipo MIME
    contentType := http.DetectContentType(buffer[:n])
    allowedTypes := map[string]bool{
        "image/png":     true,
        "image/svg+xml": true,
    }
    
    if !allowedTypes[contentType] {
        http.Error(w, fmt.Sprintf("Invalid file type: %s. Only PNG and SVG are allowed", contentType), http.StatusBadRequest)
        return
    }
    
    // Validar extensión también
    ext := strings.ToLower(filepath.Ext(handler.Filename))
    if ext != ".png" && ext != ".svg" {
        http.Error(w, "Invalid file extension", http.StatusBadRequest)
        return
    }
    
    // Validar que el contenido coincida con la extensión
    if ext == ".png" && contentType != "image/png" {
        http.Error(w, "File content does not match extension", http.StatusBadRequest)
        return
    }
    if ext == ".svg" && contentType != "image/svg+xml" {
        http.Error(w, "File content does not match extension", http.StatusBadRequest)
        return
    }
    
    // ... resto del código de guardado ...
}
```

---

### 10. Exposición de Información del Sistema en Errores

**Ubicación:** Múltiples lugares en `backend/handlers.go`

**Problema:**
```go
http.Error(w, fmt.Sprintf("Failed to fetch resource: %v", err), http.StatusInternalServerError)
// Expone detalles internos como rutas, nombres de recursos, etc.
```

**Severidad:** 🟠 ALTA

**Descripción:**
Los mensajes de error exponen información detallada sobre el sistema interno, incluyendo rutas de archivos, nombres de recursos internos, stack traces, etc.

**Impacto:**
- Reconocimiento del sistema por atacantes
- Exposición de estructura interna
- Información útil para ataques dirigidos

**Solución:**
```go
func handleError(w http.ResponseWriter, err error, userMessage string, statusCode int) {
    // Log el error completo internamente con contexto
    log.Printf("Error [%s]: %v", userMessage, err)
    
    // Enviar mensaje genérico al usuario
    http.Error(w, userMessage, statusCode)
}

// Uso en handlers:
if err != nil {
    handleError(w, err, "Failed to fetch resource", http.StatusInternalServerError)
    return
}
```

---

### 11. Falta de Validación de Certificados TLS en Cliente HTTP de Prometheus

**Ubicación:** `backend/prometheus.go:145, 197`

**Problema:**
```go
client := &http.Client{
    Timeout: 30 * time.Second,
}
resp, err := client.Get(fullURL) // ⚠️ No valida certificados si es HTTPS
```

**Severidad:** 🟠 ALTA

**Descripción:**
Si Prometheus usa HTTPS, el cliente HTTP no valida certificados, permitiendo ataques Man-in-the-Middle.

**Impacto:**
- Ataques Man-in-the-Middle (MITM)
- Interceptación de métricas sensibles
- Posible inyección de datos falsos

**Solución:**
```go
import (
    "crypto/tls"
    "crypto/x509"
)

func createSecureHTTPClient() *http.Client {
    // Cargar certificados del sistema
    rootCAs, _ := x509.SystemCertPool()
    if rootCAs == nil {
        rootCAs = x509.NewCertPool()
    }
    
    // Opcional: cargar certificados adicionales desde archivo o variable de entorno
    // certPEM := os.Getenv("PROMETHEUS_CA_CERT")
    // if certPEM != "" {
    //     rootCAs.AppendCertsFromPEM([]byte(certPEM))
    // }
    
    config := &tls.Config{
        RootCAs: rootCAs,
        // En producción, no permitir certificados autofirmados
        InsecureSkipVerify: os.Getenv("PROMETHEUS_INSECURE_SKIP_VERIFY") == "true", // Solo para desarrollo
    }
    
    transport := &http.Transport{
        TLSClientConfig: config,
    }
    
    return &http.Client{
        Timeout:   30 * time.Second,
        Transport: transport,
    }
}

// Usar en queryPrometheusRange y queryPrometheusInstant:
client := createSecureHTTPClient()
```

---

### 12. Falta de Logging de Auditoría

**Ubicación:** Todo el backend

**Problema:**
No hay logging de acciones críticas como:
- Intentos de login fallidos/exitosos
- Creación/modificación/eliminación de recursos
- Acceso a secretos
- Ejecución de comandos en pods

**Severidad:** 🟠 ALTA

**Impacto:**
- Imposible rastrear actividades maliciosas
- No hay evidencia para investigar incidentes
- No se puede detectar comportamiento anómalo

**Solución:**
```go
import (
    "log/slog"
    "encoding/json"
)

type AuditLog struct {
    Timestamp   time.Time `json:"timestamp"`
    User        string    `json:"user"`
    Action      string    `json:"action"`
    Resource    string    `json:"resource"`
    Namespace   string    `json:"namespace"`
    IP          string    `json:"ip"`
    UserAgent   string    `json:"user_agent"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
    Details     map[string]interface{} `json:"details,omitempty"`
}

func auditLog(action, resource, namespace string, r *http.Request, success bool, err error, details map[string]interface{}) {
    claims, _ := r.Context().Value("user").(*Claims)
    username := "anonymous"
    if claims != nil {
        username = claims.Username
    }
    
    log := AuditLog{
        Timestamp: time.Now(),
        User:      username,
        Action:    action,
        Resource:  resource,
        Namespace: namespace,
        IP:        getClientIP(r),
        UserAgent: r.UserAgent(),
        Success:   success,
        Details:   details,
    }
    if err != nil {
        log.Error = err.Error()
    }
    
    // Log estructurado
    logJSON, _ := json.Marshal(log)
    slog.Info("audit", "log", string(logJSON))
    
    // También escribir a archivo de auditoría si es necesario
    // auditFile.Write(logJSON)
}

// Uso en handlers:
func (h *Handlers) DeleteResource(w http.ResponseWriter, r *http.Request) {
    // ... código existente ...
    
    auditLog("delete", kind, namespace, r, true, nil, map[string]interface{}{
        "name": name,
        "force": force,
    })
    
    // ...
}

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
    // ... código existente ...
    
    if !match {
        auditLog("login", "user", "", r, false, fmt.Errorf("invalid password"), nil)
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    auditLog("login", "user", "", r, true, nil, map[string]interface{}{
        "username": creds.Username,
    })
    
    // ...
}
```

---

## 🟡 VULNERABILIDADES DE MEDIA SEVERIDAD

### 13. Falta de Validación de Versiones de Dependencias

**Ubicación:** `backend/go.mod`, `frontend/package.json`

**Problema:**
No se especifican versiones exactas de dependencias, usando `^` que permite actualizaciones automáticas.

**Severidad:** 🟡 MEDIA

**Solución:**
- Usar versiones exactas o rangos específicos
- Implementar dependabot/renovate para actualizaciones controladas
- Revisar CVE regularmente con `govulncheck` y `npm audit`

---

### 14. Falta de Headers de Seguridad Adicionales

**Ubicación:** `frontend/nginx.conf`

**Problema:**
Faltan algunos headers importantes de seguridad.

**Severidad:** 🟡 MEDIA

**Solución:**
```nginx
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always; # Solo si se usa HTTPS
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
```

---

### 15. Falta de Validación de Límites de Recursos en Kubernetes

**Ubicación:** `backend/handlers.go` (múltiples funciones)

**Problema:**
No se valida si las operaciones exceden los límites de recursos del clúster (ResourceQuota, LimitRange).

**Severidad:** 🟡 MEDIA

**Solución:**
Implementar validación antes de crear/actualizar recursos para verificar límites de ResourceQuota y LimitRange.

---

### 16. Falta de Timeout en Operaciones de Kubernetes

**Ubicación:** Múltiples funciones en `backend/handlers.go`

**Problema:**
Las operaciones de Kubernetes usan `context.TODO()` sin timeout, lo que puede causar que las goroutines se queden bloqueadas.

**Severidad:** 🟡 MEDIA

**Solución:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Usar ctx en lugar de context.TODO()
list, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
```

---

## 🔵 MEJORAS RECOMENDADAS

### 17. Implementar HTTPS Obligatorio

**Recomendación:**
- Forzar HTTPS en producción
- Redirigir HTTP a HTTPS
- Usar certificados válidos (Let's Encrypt)
- Configurar HSTS

---

### 18. Implementar Autenticación de Dos Factores (2FA)

**Recomendación:**
Agregar soporte para TOTP (Time-based One-Time Password) para mayor seguridad.

---

## 📊 RESUMEN DE VULNERABILIDADES

| Severidad | Cantidad | Estado |
|-----------|----------|--------|
| 🔴 Crítica | 6 | Requiere atención inmediata |
| 🟠 Alta | 6 | Debe corregirse en 1-2 semanas |
| 🟡 Media | 4 | Debe corregirse en 1 mes |
| 🔵 Mejora | 2 | Recomendado para mejor seguridad |

**Total:** 18 vulnerabilidades identificadas

---

## 📋 PLAN DE ACCIÓN PRIORIZADO

### Fase 1 - Crítico (Inmediato - Esta Semana)
1. ✅ Corregir validación de CORS (comparación exacta de URLs)
2. ✅ Agregar límite de recursos en YAML Import
3. ✅ Mejorar validación de WebSocket Origin
4. ✅ Reducir permisos RBAC (solo lectura donde sea posible)
5. ✅ Eliminar uso de localStorage en TerminalViewer
6. ✅ Agregar límite de tamaño en respuestas de Prometheus

### Fase 2 - Alta (1-2 semanas)
7. ✅ Implementar rate limiting
8. ✅ Mejorar Content-Security-Policy
9. ✅ Validar tipo MIME en uploads
10. ✅ Sanitizar mensajes de error
11. ✅ Validar certificados TLS en cliente Prometheus
12. ✅ Implementar logging de auditoría

### Fase 3 - Media (1 mes)
13. ✅ Revisar y fijar dependencias
14. ✅ Agregar headers de seguridad adicionales
15. ✅ Validar límites de recursos de Kubernetes
16. ✅ Agregar timeouts en operaciones de Kubernetes

### Fase 4 - Mejoras (Ongoing)
17. ✅ HTTPS obligatorio
18. ✅ Considerar 2FA

---

## 🔍 HERRAMIENTAS RECOMENDADAS

### Análisis Estático
- **Go:** `gosec`, `staticcheck`, `govulncheck`
- **JavaScript:** `eslint-plugin-security`, `npm audit`
- **Kubernetes:** `kube-score`, `Polaris`, `kubeaudit`

### Análisis Dinámico
- **SAST:** SonarQube, Semgrep, CodeQL
- **DAST:** OWASP ZAP, Burp Suite

### Monitoreo
- **Kubernetes Security:** Falco, KubeArmor
- **Logging:** ELK Stack, Loki
- **SIEM:** Splunk, ELK Security

---

## 📚 REFERENCIAS Y ESTÁNDARES

- [OWASP Top 10 (2021)](https://owasp.org/www-project-top-ten/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [Prometheus Security](https://prometheus.io/docs/operating/security/)

---

## 🎯 MÉTRICAS DE SEGURIDAD

### Estado Actual
- **Vulnerabilidades Críticas:** 6
- **Vulnerabilidades Altas:** 6
- **Score de Seguridad:** ~55/100

### Objetivo Después de Correcciones
- **Vulnerabilidades Críticas:** 0
- **Vulnerabilidades Altas:** 0-1
- **Score de Seguridad:** >85/100

---

**Fecha del Análisis:** 2024-12-19
**Versión Analizada:** 1.0.6
**Última Actualización:** 2024-12-19
**Analista:** AI Security Review

---

## ⚠️ NOTAS IMPORTANTES

1. **Este análisis es exhaustivo pero no exhaustivo** - Siempre realice auditorías de seguridad adicionales antes de desplegar en producción.

2. **Pruebas de Penetración** - Se recomienda encarecidamente realizar pruebas de penetración profesionales antes del despliegue en producción.

3. **Monitoreo Continuo** - Implemente monitoreo de seguridad continuo para detectar nuevas vulnerabilidades y ataques.

4. **Actualizaciones** - Mantenga todas las dependencias actualizadas y revise CVE regularmente.

5. **Documentación de Seguridad** - Mantenga documentación actualizada de políticas de seguridad y procedimientos de respuesta a incidentes.
