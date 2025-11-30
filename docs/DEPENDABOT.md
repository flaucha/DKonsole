# 🤖 Dependabot - Guía Completa

## 📖 ¿Qué es Dependabot?

**Dependabot** es un bot automatizado de GitHub que mantiene las dependencias de tu proyecto actualizadas y seguras. Funciona de forma completamente automática, creando Pull Requests cuando detecta actualizaciones disponibles.

### Características principales:
- ✅ **Detección automática** de dependencias desactualizadas
- ✅ **Creación de Pull Requests** para cada actualización
- ✅ **Validación automática** mediante workflows de CI/CD
- ✅ **Mantenimiento de seguridad** actualizando dependencias vulnerables
- ✅ **Schedule configurable** (diario, semanal, mensual)

---

## 🏗️ Arquitectura y Funcionamiento

### 1. **Proceso de Escaneo**

```
┌─────────────────────────────────────────┐
│  Dependabot escanea según schedule      │
│  (Lunes para Go/npm, Mensual para CI)   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Detecta dependencias en:                │
│  • backend/go.mod (Go modules)           │
│  • frontend/package.json (npm)           │
│  • .github/workflows/*.yml (Actions)     │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Compara versiones actuales vs          │
│  versiones disponibles en registros      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Crea Pull Request por cada actualización│
│  (hasta el límite configurado)           │
└─────────────────────────────────────────┘
```

### 2. **Flujo de un Pull Request**

1. **Dependabot crea el PR**
   - Actualiza el archivo de dependencias (`go.mod`, `package.json`, etc.)
   - Usa mensajes de commit con prefijo "chore"
   - Aplica labels automáticos (`dependencies`, `go`, `javascript`, etc.)
   - Asigna reviewers configurados

2. **Workflow de CI se ejecuta automáticamente**
   - `pr-checks.yml` se dispara en cada PR
   - Ejecuta tests del backend (Go)
   - Ejecuta tests del frontend (Node.js)
   - Valida linting y seguridad
   - Sube coverage a Codecov

3. **Validación y aprobación**
   - Si los checks pasan ✅ → PR listo para revisión
   - Si los checks fallan ❌ → PR necesita atención
   - Reviewer puede aprobar y mergear

4. **Merge automático (opcional)**
   - Si está configurado, puede mergear automáticamente cuando:
     - Los checks pasan
     - El PR está aprobado
     - No hay conflictos

---

## ⚙️ Configuración Actual

El archivo `.github/dependabot.yml` controla todo el comportamiento de Dependabot:

### 📦 Backend - Go Modules
```yaml
- package-ecosystem: "gomod"
  directory: "/backend"
  schedule:
    interval: "weekly"      # Cada lunes
    day: "monday"
  open-pull-requests-limit: 5  # Máximo 5 PRs abiertos
  labels:
    - "dependencies"
    - "go"
  reviewers:
    - "flaucha"
```

**Qué actualiza:**
- Todas las dependencias en `backend/go.mod`
- Incluye dependencias directas e indirectas
- Respeta las restricciones de versión en `go.mod`

### 📦 Frontend - npm
```yaml
- package-ecosystem: "npm"
  directory: "/frontend"
  schedule:
    interval: "weekly"      # Cada lunes
    day: "monday"
  open-pull-requests-limit: 5
  ignore:
    - dependency-name: "*"
      update-types: ["version-update:semver-patch"]  # Ignora patches de devDependencies
  labels:
    - "dependencies"
    - "javascript"
```

**Qué actualiza:**
- Dependencias en `frontend/package.json`
- Solo actualizaciones `major` y `minor`
- Ignora `patch` de devDependencies (para reducir ruido)

### 🔧 GitHub Actions
```yaml
- package-ecosystem: "github-actions"
  directory: "/"
  schedule:
    interval: "monthly"    # Una vez al mes
  labels:
    - "dependencies"
    - "ci"
```

**Qué actualiza:**
- Versiones de acciones en `.github/workflows/*.yml`
- Ejemplo: `actions/checkout@v4` → `actions/checkout@v6`

---

## 🛠️ Comandos Útiles

### Ver PRs abiertos de Dependabot
```bash
gh pr list --author "app/dependabot" --state open
```

### Ver detalles de un PR específico
```bash
gh pr view <PR_NUMBER> --repo flaucha/DKonsole
```

### Cerrar todos los PRs de Dependabot
```bash
cd /home/flaucha/repos/DKonsole
./scripts/close-dependabot-prs.sh
```

### Aprobar un PR de Dependabot
```bash
gh pr review <PR_NUMBER> --approve --repo flaucha/DKonsole
```

### Mergear un PR de Dependabot
```bash
gh pr merge <PR_NUMBER> --squash --repo flaucha/DKonsole
```

### Ver configuración actual
```bash
cat .github/dependabot.yml
```

### Ver historial de PRs de Dependabot
```bash
gh pr list --author "app/dependabot" --state all --limit 20
```

---

## 📅 Schedule y Timing

### Cuándo se ejecuta Dependabot:

| Ecosistema | Frecuencia | Día/Hora | Límite PRs |
|------------|------------|----------|------------|
| Go modules | Semanal | Lunes | 5 |
| npm | Semanal | Lunes | 5 |
| GitHub Actions | Mensual | Variable | Sin límite |

**Nota:** Dependabot puede tardar hasta 24 horas después del schedule en crear los PRs.

---

## 🔍 Validación Automática

Cuando Dependabot crea un PR, automáticamente se ejecuta el workflow `.github/workflows/pr-checks.yml`:

### Backend Checks:
- ✅ `go mod tidy` - Limpia dependencias
- ✅ `go vet` - Análisis estático
- ✅ `golangci-lint` - Linting avanzado
- ✅ `govulncheck` - Escaneo de vulnerabilidades
- ✅ Tests unitarios con coverage

### Frontend Checks:
- ✅ `npm install` - Instala dependencias
- ✅ `npm audit` - Escaneo de vulnerabilidades
- ✅ `npm run lint` - Linting
- ✅ `npm run test` - Tests con coverage

**Si todos los checks pasan:** El PR está listo para revisión y merge.

---

## ⚙️ Personalización Avanzada

### Cambiar frecuencia de actualizaciones

```yaml
schedule:
  interval: "daily"    # Opciones: daily, weekly, monthly
  day: "monday"        # Solo para weekly
  time: "09:00"        # Hora UTC (opcional)
```

### Ajustar límite de PRs

```yaml
open-pull-requests-limit: 10  # Aumentar o disminuir según necesidad
```

### Ignorar dependencias específicas

```yaml
ignore:
  # Ignorar una dependencia completamente
  - dependency-name: "nombre-del-paquete"

  # Ignorar solo actualizaciones mayores
  - dependency-name: "otro-paquete"
    update-types: ["version-update:semver-major"]

  # Ignorar versiones específicas
  - dependency-name: "paquete-problematico"
    versions: [">= 2.0.0, < 3.0.0"]
```

### Agrupar actualizaciones (reducir número de PRs)

```yaml
groups:
  production-dependencies:
    patterns:
      - "express"
      - "react"
      - "lodash"
  dev-dependencies:
    patterns:
      - "vitest"
      - "eslint"
```

### Configurar auto-merge (avanzado)

Requiere configuración adicional en GitHub Settings:
1. Settings → Actions → General
2. Habilitar "Allow GitHub Actions to create and approve pull requests"
3. Configurar branch protection rules

---

## 🚨 Solución de Problemas

### ❌ PRs fallando constantemente

**Síntomas:** Los PRs de Dependabot fallan en los checks de CI

**Soluciones:**
1. Verifica que `pr-checks.yml` esté funcionando:
   ```bash
   gh workflow view pr-checks.yml --repo flaucha/DKonsole
   ```

2. Revisa los logs del workflow:
   - Ve a GitHub → Actions → PR Checks
   - Revisa qué step está fallando

3. Ejecuta los tests localmente:
   ```bash
   cd backend && go test ./...
   cd ../frontend && npm test
   ```

4. Verifica que las dependencias sean compatibles:
   - Revisa los changelogs de las dependencias actualizadas
   - Puede haber breaking changes

### 📈 Demasiados PRs

**Síntomas:** Dependabot crea muchos PRs y es difícil mantenerlos

**Soluciones:**
1. Reducir el límite:
   ```yaml
   open-pull-requests-limit: 3
   ```

2. Aumentar el intervalo:
   ```yaml
   schedule:
     interval: "monthly"  # En lugar de weekly
   ```

3. Ignorar más tipos de actualizaciones:
   ```yaml
   ignore:
     - dependency-name: "*"
       update-types: ["version-update:semver-patch"]
   ```

4. Agrupar dependencias relacionadas (ver sección anterior)

### 🔇 PRs no se crean

**Síntomas:** Dependabot no está creando PRs aunque hay actualizaciones

**Soluciones:**
1. Verifica que Dependabot esté habilitado:
   - Settings → Security → Dependabot
   - Asegúrate de que "Dependabot version updates" esté activo

2. Verifica el archivo de configuración:
   ```bash
   cat .github/dependabot.yml
   ```
   - Debe estar en la rama `main`
   - Debe tener sintaxis YAML válida

3. Espera al próximo schedule:
   - Dependabot puede tardar hasta 24 horas
   - Los schedules semanales se ejecutan el día configurado

4. Revisa los logs de Dependabot:
   - Settings → Security → Dependabot → Insights
   - Busca errores o advertencias

### 🔄 PRs se recrean constantemente

**Síntomas:** Cierras un PR y Dependabot lo vuelve a crear

**Causa:** La dependencia sigue desactualizada

**Soluciones:**
1. Mergear el PR en lugar de cerrarlo
2. Ignorar la dependencia si no quieres actualizarla:
   ```yaml
   ignore:
     - dependency-name: "nombre-del-paquete"
   ```

---

## 📊 Monitoreo y Estadísticas

### Ver actividad de Dependabot

```bash
# PRs creados en el último mes
gh pr list --author "app/dependabot" --state all --limit 50

# PRs abiertos actualmente
gh pr list --author "app/dependabot" --state open

# PRs cerrados (no mergeados)
gh pr list --author "app/dependabot" --state closed --limit 20
```

### Dashboard de GitHub

1. Ve a: `https://github.com/flaucha/DKonsole/security/dependabot`
2. Revisa:
   - Alertas de seguridad
   - PRs pendientes
   - Estadísticas de actualizaciones

---

## 🔐 Seguridad

### Dependabot Security Updates

Además de las actualizaciones de versión, Dependabot también crea PRs automáticos para vulnerabilidades críticas:

- **Automático:** No requiere configuración
- **Prioritario:** Se crean inmediatamente, sin esperar schedule
- **Etiquetado:** Con label `security`

### Ver vulnerabilidades

```bash
# Ver alertas de seguridad
gh api repos/flaucha/DKonsole/dependabot/alerts

# Ver PRs de seguridad
gh pr list --label "security" --author "app/dependabot"
```

---

## 📚 Recursos y Referencias

### Documentación Oficial
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [dependabot.yml Configuration](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
- [Dependabot Security Updates](https://docs.github.com/en/code-security/dependabot/dependabot-security-updates)

### Scripts Útiles
- `./scripts/close-dependabot-prs.sh` - Cerrar todos los PRs de Dependabot

### Archivos Relacionados
- `.github/dependabot.yml` - Configuración principal
- `.github/workflows/pr-checks.yml` - Workflow de validación
- `docs/DEPENDABOT.md` - Esta documentación

---

## 💡 Mejores Prácticas

1. **Revisa regularmente los PRs**
   - No dejes que se acumulen demasiados
   - Mergea los que pasan los checks

2. **Mantén el límite de PRs bajo**
   - 5 PRs es un buen balance
   - Facilita la revisión y merge

3. **Ignora dependencias problemáticas**
   - Si una dependencia causa problemas constantemente
   - Agrégalo a la lista de `ignore`

4. **Agrupa dependencias relacionadas**
   - Reduce el número de PRs
   - Facilita el testing conjunto

5. **Revisa los changelogs**
   - Antes de mergear, revisa breaking changes
   - Especialmente en actualizaciones mayores

---

## 🎯 Resumen Rápido

| Acción | Comando |
|--------|---------|
| Ver PRs abiertos | `gh pr list --author "app/dependabot"` |
| Cerrar todos los PRs | `./scripts/close-dependabot-prs.sh` |
| Aprobar un PR | `gh pr review <NUM> --approve` |
| Mergear un PR | `gh pr merge <NUM> --squash` |
| Ver configuración | `cat .github/dependabot.yml` |

---

**Última actualización:** 2025-11-30
