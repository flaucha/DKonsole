# 🤖 Cómo Funciona Dependabot

## ¿Qué es Dependabot?

Dependabot es un bot automatizado de GitHub que:
- **Detecta dependencias desactualizadas** en tu proyecto
- **Crea Pull Requests automáticamente** para actualizar las dependencias
- **Mantiene tu proyecto seguro** actualizando dependencias vulnerables
- **Sigue un schedule configurado** (semanal, mensual, etc.)

## 📋 Configuración Actual

El archivo `.github/dependabot.yml` define cómo funciona Dependabot en este proyecto:

### Backend (Go modules)
- **Frecuencia**: Semanal (cada lunes)
- **Límite de PRs**: Máximo 5 PRs abiertos simultáneamente
- **Labels**: `dependencies`, `go`
- **Reviewer**: `flaucha`

### Frontend (npm)
- **Frecuencia**: Semanal (cada lunes)
- **Límite de PRs**: Máximo 5 PRs abiertos simultáneamente
- **Labels**: `dependencies`, `javascript`
- **Reviewer**: `flaucha`
- **Ignora**: Actualizaciones patch de devDependencies

### GitHub Actions
- **Frecuencia**: Mensual
- **Labels**: `dependencies`, `ci`

## 🔄 Flujo de Trabajo

1. **Dependabot escanea el proyecto** según el schedule configurado
2. **Detecta dependencias desactualizadas** comparando con las versiones más recientes
3. **Crea un Pull Request** para cada actualización
4. **Ejecuta los workflows de CI** (pr-checks.yml) para validar los cambios
5. **Espera aprobación** del reviewer antes de mergear

## ✅ Validación Automática

Cuando Dependabot crea un PR, automáticamente se ejecuta:
- ✅ Tests del backend (Go)
- ✅ Tests del frontend (Node.js)
- ✅ Linting y validaciones de código
- ✅ Escaneo de vulnerabilidades

Si los checks pasan, el PR está listo para revisión.

## 🛠️ Comandos Útiles

### Ver PRs abiertos de Dependabot
```bash
gh pr list --author "app/dependabot" --state open
```

### Cerrar todos los PRs de Dependabot
```bash
./scripts/close-dependabot-prs.sh
```

### Aprobar y mergear un PR de Dependabot
```bash
gh pr review <PR_NUMBER> --approve
gh pr merge <PR_NUMBER> --squash
```

### Ver configuración de Dependabot
```bash
cat .github/dependabot.yml
```

## ⚙️ Personalización

### Cambiar frecuencia de actualizaciones
Edita `.github/dependabot.yml`:
```yaml
schedule:
  interval: "daily"  # daily, weekly, monthly
  day: "monday"      # Solo para weekly
```

### Cambiar límite de PRs
```yaml
open-pull-requests-limit: 10  # Aumentar o disminuir
```

### Ignorar dependencias específicas
```yaml
ignore:
  - dependency-name: "nombre-del-paquete"
    update-types: ["version-update:semver-major"]
```

### Agrupar actualizaciones
```yaml
groups:
  production-dependencies:
    patterns:
      - "express"
      - "react"
```

## 🚨 Solución de Problemas

### PRs fallando constantemente
- Verifica que el workflow `pr-checks.yml` esté funcionando
- Revisa los logs del workflow en GitHub Actions
- Asegúrate de que los tests pasen localmente

### Demasiados PRs
- Reduce `open-pull-requests-limit`
- Aumenta el `interval` del schedule
- Agrega más dependencias a `ignore`

### PRs no se crean
- Verifica que Dependabot esté habilitado en Settings > Security > Dependabot
- Revisa que el archivo `.github/dependabot.yml` esté en la rama `main`
- Espera al próximo schedule (puede tardar hasta 24 horas)

## 📚 Recursos

- [Documentación oficial de Dependabot](https://docs.github.com/en/code-security/dependabot)
- [Configuración de dependabot.yml](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
