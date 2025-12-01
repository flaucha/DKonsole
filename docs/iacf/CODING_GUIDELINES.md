# IACF Coding Guidelines

Guía base, agnóstica de lenguaje, para escribir y mantener código bajo el IACF. Usa los perfiles como ejemplos, no como obligación.

## 0. Cómo usar esta guía
- Identifica si el repo es **fullstack**, **solo backend** o **solo frontend** siguiendo `IA_GUIDELINES.md`.
- Aplica los principios base y de arquitectura a cualquier stack.
- Toma los perfiles como referencia y sustitúyelos por equivalentes de tu stack. Si falta un perfil, créalo a partir de estos ejemplos.

## 1. Principios base
- **AI-First & Clean Code**: Código claro, autosuficiente y con patrones repetibles.
- **Dominios > capas**: Agrupa por dominio (`auth`, `billing`), no solo por tipo de archivo.
- **Dependencias explícitas**: Inversión de dependencias/DI; evita singletons ocultos.
- **Errores con contexto**: Propaga con mensajes claros y envoltura (`%w` o equivalente).
- **Logging estructurado**: Sin `print` en producción; logger común con claves consistentes.
- **Config externa**: Nunca hardcodees secrets/credenciales; usa config por entorno.
- **Testing por defecto**: Cada cambio significativo viene con pruebas o notas de brecha.

## 2. Arquitectura general (patrón estratificado)
Aplica este flujo, sea cual sea el stack:
- **Entrada/Handler**: Traducir transporte a modelo de dominio, validar inputs.
- **Servicio**: Orquestar reglas de negocio. Evita acoplarte a detalles de transporte/UI.
- **Repositorio/Cliente**: Acceso a datos o servicios externos. Encapsula IO aquí.
- **Fábrica/Composición**: Punto único para inyectar dependencias y construir servicios.
- **Contratos de datos**: Define tipos/DTOs que separen capas y eviten dependencias accidentales.

## 3. Patrones transversales
- **Configuración**: Variables de entorno o secret manager; defaults seguros; separación por entorno.
- **Errores y fallos**: Diferencia entre errores de dominio y fallos técnicos; no silencies excepciones.
- **Idempotencia y efectos**: Para operaciones con side effects, documenta idempotencia y reintentos.
- **Concurrencia**: Protege recursos compartidos; usa primitivas del lenguaje/framework; evita race conditions.

## 4. Perfiles de ejemplo (adapta a tu stack)
### Backend (Go)
- Estructura sugerida: `backend/internal/<dominio>/` con `factories.go`, `service.go`, `repository.go`; compartidos en `backend/internal/models/` y utilidades en `backend/internal/utils/`.
- Factories gestionan DI; no instancies servicios en handlers.
- Services reciben `context.Context`; devuelven datos/errores; el handler gestiona transporte.
- Repositories: todo IO (DB, APIs externas, K8s) vive aquí.
- Logging: slog/zap; evita `fmt.Println`/`log.Println`.
- Tests: mocks vía interfaces; cubre caminos críticos (auth, IO).

### Backend (Node/TypeScript)
- Estructura sugerida: `src/domains/<dominio>/` con `service.ts`, `repository.ts`, `controller.ts` o adaptadores HTTP/queue.
- DI/Contenedores: usa inyección explícita (factories o contenedor ligero) para servicios/repos.
- Errores: clases/objetos de error de dominio + mapeo a HTTP/gRPC en los adaptadores.
- Logging: pino/winston con formato JSON; sin `console.log` en prod.
- Tests: unit con Jest/Vitest; mocks de IO (`nock`/`msw`); contrato/API si aplica.

### Frontend (React como ejemplo de SPA)
- Componentes funcionales con Hooks; agrupa en `src/components` (shared) y `src/features/<dominio>`.
- Estado: Context/Redux/RTK Query para estado global; `useState`/`useReducer` local para UI. Evita duplicar estado con cache de datos.
- Estilos: sistema de diseño o utilidades (Tailwind/CSS Modules); evita inline salvo casos dinámicos.
- Datos/API: cliente centralizado (fetch/axios); maneja `loading`/`error` explícitamente y deduplica llamadas.
- Accesibilidad: semántica y focus gestionado en flujos críticos.
- Tests: RTL + Jest/Vitest para componentes; mocks de API (`msw`); E2E opcional (Playwright/Cypress).

## 5. Git & workflow
- Commits con **Conventional Commits** (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`).
- Versionado **SemVer** (MAJOR.MINOR.PATCH). Si hay `VERSION` o doc de release, úsalo y actualízalo.
- Antes de cerrar una tarea: lint + tests del alcance, revisión de seguridad básica y actualización de docs.

## 6. Checklist de nuevas funcionalidades
- [ ] Confirmar alcance (fullstack/backend/frontend) y stack activo.
- [ ] Definir contratos y tipos compartidos (ej. `pkg/models`, `src/types`).
- [ ] Implementar repositorio/cliente de datos (IO encapsulado).
- [ ] Implementar servicio (reglas de negocio).
- [ ] Exponer por handler/API o UI (según alcance).
- [ ] Añadir frontend (si aplica) y wiring al cliente API.
- [ ] Añadir pruebas (unidad/integración/contrato según riesgo).
- [ ] Actualizar notas de cambio o docs relevantes.
