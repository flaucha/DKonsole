# Guía de GitHub Actions para DKonsole

## 📋 ¿Qué es GitHub Actions?

GitHub Actions es un sistema de CI/CD (Integración Continua / Despliegue Continuo) integrado en GitHub que ejecuta automáticamente tareas cuando ocurren eventos en tu repositorio (como push, pull request, etc.).

## 🔧 Cómo Funciona en DKonsole

### Ubicación del Workflow

El archivo de configuración está en:
```
.github/workflows/ci.yaml
```

### ¿Cuándo se Ejecuta?

El workflow se ejecuta automáticamente cuando:

1. **Push a `main`**: Cada vez que haces push a la rama principal
2. **Pull Request a `main`**: Cuando alguien crea un PR hacia `main`

### Estructura del Workflow

El workflow tiene 3 jobs que se ejecutan:

#### 1. `test-backend` - Tests del Backend
- ✅ Verifica el código
- ✅ Instala Go 1.24
- ✅ Descarga dependencias
- ✅ Ejecuta `go vet` (verificación de código)
- ✅ Ejecuta `go test` (tests unitarios)
- ✅ Genera reporte de cobertura

#### 2. `test-frontend` - Tests del Frontend
- ✅ Instala Node.js 20
- ✅ Instala dependencias npm
- ✅ Ejecuta linter
- ✅ Ejecuta tests con Vitest
- ✅ Genera reporte de cobertura

#### 3. `build` - Compilación
- ✅ Solo se ejecuta si los tests pasan
- ✅ Compila el backend
- ✅ Compila el frontend
- ✅ Verifica que todo se puede construir

## 📊 Cómo Ver los Resultados

### Opción 1: En GitHub (Interfaz Web)

1. **Ve a tu repositorio en GitHub**
   ```
   https://github.com/tu-usuario/DKonsole
   ```

2. **Pestaña "Actions"**
   - Haz clic en la pestaña **"Actions"** en la parte superior del repositorio
   - Verás una lista de todas las ejecuciones del workflow

3. **Ver una ejecución específica**
   - Haz clic en cualquier ejecución para ver detalles
   - Verás el estado de cada job (✅ éxito, ❌ fallo, ⏸️ en progreso)

4. **Ver logs detallados**
   - Haz clic en un job específico (ej: "Test Backend")
   - Expande los pasos individuales para ver logs detallados

### Opción 2: Badge de Estado (Opcional)

Puedes agregar un badge a tu README para mostrar el estado:

```markdown
![CI](https://github.com/tu-usuario/DKonsole/workflows/CI/badge.svg)
```

### Opción 3: Notificaciones

GitHub te enviará notificaciones si:
- Un workflow falla
- Un workflow se completa exitosamente (opcional, configurable)

## 🎯 Interpretando los Resultados

### ✅ Éxito (Verde)
```
✅ Todos los tests pasaron
✅ Build completado exitosamente
```
**Significa:** Tu código está listo para merge/deploy

### ❌ Fallo (Rojo)
```
❌ Tests fallaron
❌ Build falló
```
**Significa:** Hay problemas que necesitas corregir antes de hacer merge

### ⚠️ Advertencias (Amarillo)
```
⚠️ Linter encontró problemas
⚠️ Cobertura de código baja
```
**Significa:** No bloquea, pero deberías revisar

## 🔍 Ejemplo de Flujo Completo

### 1. Haces cambios y haces commit:
```bash
git add .
git commit -m "feat: agregar nueva funcionalidad"
git push origin main
```

### 2. GitHub Actions se activa automáticamente:
- Ve a la pestaña "Actions" en GitHub
- Verás un nuevo workflow ejecutándose (icono amarillo ⏸️)

### 3. Mientras se ejecuta:
- Puedes ver el progreso en tiempo real
- Cada job muestra su estado

### 4. Resultado:
- ✅ **Verde**: Todo pasó, puedes continuar
- ❌ **Rojo**: Revisa los logs para ver qué falló

## 📝 Ver Logs Detallados

Cuando un test falla, puedes ver:

1. **Qué test falló:**
   ```
   FAIL: TestValidateK8sName
   ```

2. **Por qué falló:**
   ```
   expected: "Pod"
   got: "pod"
   ```

3. **Dónde falló:**
   ```
   internal/models/models_test.go:45
   ```

## 🛠️ Configuración Avanzada

### Ver solo tests que fallaron

En los logs, busca:
```
FAIL
```

### Ver cobertura de código

Los reportes de cobertura se generan pero necesitas configurar Codecov o similar para verlos en la UI.

### Ejecutar manualmente (si tienes permisos)

1. Ve a la pestaña "Actions"
2. Selecciona el workflow "CI"
3. Haz clic en "Run workflow"
4. Selecciona la rama y ejecuta

## 🔗 Enlaces Útiles

- **Documentación de GitHub Actions**: https://docs.github.com/en/actions
- **Marketplace de Actions**: https://github.com/marketplace?type=actions
- **Tus workflows**: `https://github.com/tu-usuario/DKonsole/actions`

## 💡 Tips

1. **Revisa los logs antes de hacer merge**: Aunque los tests pasen, revisa warnings
2. **Usa PRs para verificar**: Los workflows también se ejecutan en PRs
3. **Configura notificaciones**: Para saber inmediatamente si algo falla
4. **Revisa cobertura**: Asegúrate de que los nuevos cambios tienen tests

## 🐛 Troubleshooting

### El workflow no se ejecuta

**Causas comunes:**
- El archivo no está en `.github/workflows/ci.yaml`
- El archivo tiene errores de sintaxis YAML
- No tienes permisos para ejecutar workflows

**Solución:**
```bash
# Verificar que el archivo existe
ls -la .github/workflows/ci.yaml

# Verificar sintaxis YAML (puedes usar un validador online)
```

### Los tests pasan localmente pero fallan en GitHub

**Causas comunes:**
- Diferencias de versión (Go, Node.js)
- Variables de entorno no configuradas
- Dependencias no actualizadas

**Solución:**
- Verifica las versiones en el workflow
- Asegúrate de que `go.mod` y `package.json` están actualizados
- Revisa los logs en GitHub para ver el error exacto

### El workflow tarda mucho

**Optimizaciones:**
- Usa cache para dependencias (ya configurado)
- Ejecuta jobs en paralelo (ya configurado)
- Considera usar matrix builds solo si es necesario

## 📈 Próximos Pasos

1. **Haz un push de prueba** para ver el workflow en acción
2. **Revisa los resultados** en la pestaña Actions
3. **Configura notificaciones** si quieres recibir emails
4. **Agrega más tests** para aumentar la cobertura

