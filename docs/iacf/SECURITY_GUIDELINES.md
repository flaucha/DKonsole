# Security Guidelines

Estándares de seguridad agnósticos para proyectos que usan el IACF. Ejemplos concretos (p.ej., K8s) se dan como referencia, no como requisito.

## 1. Principios y alcance
- **Agnóstico a lenguaje/infra**: aplica a backend/frontend/infra; si usas K8s/cloud, trátalo como un perfil de ejemplo.
- **Config externa, nunca en código**: secrets y credenciales siempre fuera del repositorio (secret manager/vars de entorno). No hardcode.
- **Mínimo privilegio**: procesos, servicios y cuentas con permisos mínimos (p.ej., RBAC K8s si aplica). Usa separación por rol/entorno.
- **Superficie de ataque controlada**: expón solo lo necesario; valida inputs en la entrada y encoda outputs en la salida.

## 2. Controles por capa
- **Entrada/validación**: validar y normalizar datos en handlers/APIs (tipos, rangos, listas blancas). Rechazar lo que no se entienda.
- **Salida/encoding**: escapar/encodar en el contexto adecuado (HTML/JS/JSON). Cabeceras anti-XSS y CSP en front/web. CORS restrictivo.
- **Autenticación/Autorización**: flujos de login seguros, MFA si aplica, sessions/tokens con expiración corta y rotación; autorización por recurso/acción.
- **Datos y tokens**: clasifica datos (público/interno/sensible). Cifrado en tránsito (TLS) y en reposo según sensibilidad. No reuso de refresh tokens; TTL y revocación documentados.

## 3. OWASP Top 10 (cheatsheet agnóstico)
1. **Broken Access Control**: aplica control de acceso por recurso; pruebas negativas. (Ej. revisar policies/RBAC si usas K8s).
2. **Cryptographic Failures**: sin secretos en repos; cifrado en tránsito/at rest; rotación de llaves; algoritmos y modos seguros.
3. **Injection**: inputs parametrizados (SQL/NoSQL/LDAP/commands); sanitiza rutas/IDs; evita concatenar comandos/queries.
4. **Insecure Design**: threat model básico en el plan; identifica abusos y límites.
5. **Security Misconfiguration**: defaults seguros; deshabilita endpoints de debug; configura cabeceras de seguridad; si usas Helm/K8s, valores seguros por defecto.
6. **Vulnerable/Outdated Components**: usa lockfiles/version pinning; escanea dependencias regularmente; registries/imagenes de confianza.
7. **Identification & Authentication Failures**: manejo robusto de sesiones/JWT (aud/iss/exp/nbf), políticas de contraseñas/cookies (HttpOnly, Secure, SameSite).
8. **Software & Data Integrity Failures**: builds reproducibles, firma/verificación de artefactos; control de integridad en pipelines; no ejecutar código no verificado.
9. **Security Logging & Monitoring Failures**: logs estructurados de eventos críticos (authn/authz, cambios de config, acciones privilegiadas), con retención y alertas básicas. No loggear secrets/PII.
10. **SSRF**: valida y restringe URLs externas (listas blancas, timeouts, no acceso a metadata interna); separa redes si es posible.

## 4. Supply chain y pipelines seguros
- **SAST/DAST/dep scanning** obligatorio en CI según el stack.
- **Lockfiles y pinning**: mantén y revisa `go.sum`/`package-lock`/`poetry.lock`/`pnpm-lock`/`pom.xml`/`Cargo.lock`, etc.
- **Imágenes y artefactos confiables**: bases mínimas y firmadas; verifica firmas/digests en despliegues.
- **Build reproducible**: evita scripts no deterministas; documenta pasos; si se firma, verifica en deploy.

## 5. Observabilidad y auditoría
- **Qué loggear**: authn/authz (success/fail), cambios de configuración, acciones privilegiadas, operaciones sobre datos sensibles.
- **Formato y retención**: logs estructurados; retención acorde a normativa; proteger logs de escritura indebida; mascarar/omitir PII y secretos.
- **Alertas mínimas**: fallos repetidos de login, escaladas de permisos, anomalías de tráfico (rate limit superado).

## 6. Superficie externa y resiliencia
- **Cabeceras de seguridad**: HSTS, X-Content-Type-Options, X-Frame-Options/Frame-Ancestors, Referrer-Policy, CSP apropiada.
- **Rate limiting y anti-abuso**: límites por IP/usuario/token en endpoints públicos; protección básica contra DoS/botting.
- **Egress control**: restringir salidas a dominios permitidos; timeouts y retries acotados.

## 7. Multi-tenant y segregación de datos (si aplica)
- Separar datos por tenant/organización; controles de autorización en cada consulta/escritura.
- Aislar recursos de infraestructura por tenant cuando sea posible; límites por tenant para evitar ruido lateral.

## 8. Proceso de análisis y checklist por cambio
1. **Scan**: revisar el cambio contra OWASP y dependencias.
2. **Checklist por cambio**: inputs validados, outputs encodados, authn/authz aplicados, secretos fuera del código, dependencias escaneadas/pineadas, logs de eventos críticos, pruebas de seguridad (unitarias/contrato) cuando aplique.
3. **Report**: documentar hallazgos con severidad (High/Medium/Low) y plan/due date de remediación.
4. **Score**: asignar puntaje de 0-100 basado en hallazgos.
5. **Cadencia**: revisar/actualizar esta guía cuando cambie el stack o al menos por release mayor.

## 9. SWOT Analysis (Security Focus)
Incluye un SWOT en reportes de seguridad:
- **Strengths**: controles existentes (auth robusta, defaults seguros, monitoreo).
- **Weaknesses**: brechas actuales (tests faltantes, dependencias viejas, falta de rate limiting).
- **Opportunities**: mejoras de seguridad (nuevos escaneos, rotación automática de claves, sandboxing adicional).
- **Threats**: riesgos externos (zero-days, cambios de APIs cloud, exposición de endpoints).
