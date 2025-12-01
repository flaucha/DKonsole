# Testing Guidelines

Estrategia de pruebas agnóstica al lenguaje para proyectos que usan el IACF. El foco mínimo es asegurar pruebas unitarias en el build; el resto es riesgo-dirigido.

## 1. Principios base (aplican a cualquier stack)
- **Unit tests son bloqueantes**: deben ejecutarse en cada build o verificación del proyecto y pasar. Ubica los tests junto al código o en la convención del stack.
- **Riesgo primero**: integra pruebas adicionales según criticidad (integración, contrato/API, E2E) para caminos críticos (auth, pagos, datos sensibles).
- **Mocks y dobles de prueba**: aísla dependencias externas (IO, red, filesystem) para unidades; reserva entornos reales para integración controlada.
- **Cobertura**: prioriza >80% en lógica de negocio y 100% en flujos críticos; reporta brechas si no se alcanza y justifica.
- **Reproducible**: comandos de test deben ser determinísticos y ejecutarse sin dependencias manuales; si falta script, créalo o proponlo.

## 2. Cómo elegir y ejecutar (agnóstico)
1. Detecta el stack con `IA_GUIDELINES.md` (artefactos como `go.mod`, `package.json`, `pyproject.toml`, `pom.xml`, `Cargo.toml`, etc.).
2. Usa la herramienta estándar del stack para unit tests:
   - Go: `go test ./...`
   - Node/TS: `npm test` / `pnpm test` / `yarn test`
   - Python: `pytest`
   - Java: `mvn test` / `gradle test`
   - Rust: `cargo test`
3. Si no hay script `test`/`check`, define uno mínimo y documenta la convención adoptada.
4. Para integración/contrato/E2E, solo habilita lo necesario según riesgo y deja claro prerequisitos (servicios, contenedores, seeds).

## 3. Perfiles de referencia (ejemplos, adapta a tu stack)
- **Backend (Go)**: `_test.go` junto al código; mocks vía interfaces; integración opcional en `backend/tests/integration`; `go test ./... -cover`.
- **Backend (Node/TS)**: `src/**/*.test.ts|js`; Jest/Vitest + `ts-jest`/`ts-node`; mocks de IO con `nock`/`msw`/dobles manuales.
- **Backend (Python)**: `tests/` + `pytest`; fixtures para IO; marca integración con `-m`.
- **Frontend (React/Vue/SPA)**: tests de componentes con RTL + Jest/Vitest; mock de fetch/API con `msw`; E2E opcional con Playwright/Cypress.

## 4. Responsabilidad de la IA
- **Features**: agrega unit tests para la lógica tocada; si es crítico, añade integración o contrato.
- **Fixes**: incluye test de regresión que falle antes del fix.
- **Refactors**: asegura que la suite existente pase; no reduzcas cobertura en rutas críticas.
- **Scripts**: si no existe comando de test para el stack detectado, proponlo/crea uno mínimo y documenta la convención en las guías del proyecto.
