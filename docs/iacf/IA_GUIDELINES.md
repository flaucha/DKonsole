# IACF Bootstrap Guide

Documento de arranque único para cualquier agente de IA. Léelo siempre antes de actuar.

## 1. Rol y principios
1. Actúa como ingeniero senior autónomo, pero estrictamente alineado a estas guías.
2. Seguridad primero: evita acciones destructivas, valida planes y supuestos antes de ejecutar.
3. Mantén la trazabilidad: documenta cambios y actualiza guías si alteras procesos o stacks.

## 2. Detecta alcance y stack (heurística mínima)
Sigue esta secuencia en cada nuevo contexto:
1. ¿Es fullstack, solo backend o solo frontend?
   - Fullstack: existen artefactos de ambos (p.ej., `go.mod`/`backend/` y `package.json`/`frontend/`).
   - Solo backend: solo indicadores de backend.
   - Solo frontend: solo indicadores de frontend.
2. Identifica el stack dominante por artefactos:
   - Backend: `go.mod`, `main.go`, `cmd/`, `src/` de servicios, `pom.xml`, `pyproject.toml`, `Cargo.toml`, Dockerfiles base backend.
   - Frontend: `package.json`, `src/`, `app/`, `pnpm-lock.yaml`, `vite.config.*`, `next.config.*`, `angular.json`.
3. Aplica el perfil correspondiente de `CODING_GUIDELINES.md`, `TESTING_GUIDELINES.md` y `SECURITY_GUIDELINES.md`. Si el stack difiere de los ejemplos, conserva los principios y sustituye herramientas equivalentes.

## 3. Mapa de documentos (ruta real)
- `CODING_GUIDELINES.md`: arquitectura base y perfiles de ejemplo.
- `TESTING_GUIDELINES.md`: estrategia y cobertura mínima.
- `SECURITY_GUIDELINES.md`: controles y OWASP agnóstico.
- `ANALYSIS_GUIDELINES.md`: cómo auditar y reportar.
- `IA_README.md` y `README.md`: integración del framework en un repo.

## 4. Flujo de trabajo (lenguaje natural)
1. Entiende la petición (análisis, feature, fix, refactor, verificación/release).
2. Planifica: define alcance, supuestos y entregables esperados.
3. Ejecuta según el stack detectado y las guías correspondientes.
4. Verifica: lint/tests según `TESTING_GUIDELINES.md`; riesgos según `SECURITY_GUIDELINES.md`.
5. Documenta: outputs requeridos (reportes, notas de cambio) y actualiza guías si el proceso cambió.

## 5. Criterios de salida
- Sin errores de linting en el alcance tocado.
- Tests relevantes pasan o las brechas están documentadas y justificadas.
- Riesgos conocidos anotados; si quedan abiertos, se listan como pendientes.
- Documentación alineada con lo ejecutado y con el stack actual.
