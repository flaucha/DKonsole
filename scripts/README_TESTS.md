# Scripts de Testing para DKonsole

Esta carpeta contiene scripts para ejecutar tests del frontend y backend de DKonsole.

## Scripts Disponibles

### `test-frontend.sh`

Ejecuta los tests del frontend (React + Vitest).

**Uso básico:**
```bash
./scripts/test-frontend.sh
```

**Opciones:**
- `--watch` - Modo watch (por defecto, se ejecuta automáticamente al cambiar archivos)
- `--run` - Ejecutar una vez y salir
- `--ui` - Ejecutar con interfaz gráfica (muy útil para debugging)
- `--coverage` - Ejecutar con cobertura de código
- `--lint-only` - Solo ejecutar el linter
- `--no-lint` - No ejecutar el linter antes de los tests

**Ejemplos:**
```bash
# Modo watch (recomendado para desarrollo)
./scripts/test-frontend.sh --watch

# Ejecutar una vez (útil para CI)
./scripts/test-frontend.sh --run

# Con interfaz gráfica
./scripts/test-frontend.sh --ui

# Con cobertura
./scripts/test-frontend.sh --coverage

# Solo linter
./scripts/test-frontend.sh --lint-only
```

### `test-backend.sh`

Ejecuta los tests del backend (Go).

**Uso básico:**
```bash
./scripts/test-backend.sh
```

**Opciones:**
- `--verbose` - Mostrar salida detallada (por defecto)
- `--quiet` - Salida mínima
- `--coverage` - Generar reporte de cobertura
- `--module <nombre>` - Ejecutar tests de un módulo específico (ej: `utils`, `models`, `auth`)
- `--vet-only` - Solo ejecutar `go vet` (no tests)
- `--no-vet` - No ejecutar `go vet` antes de los tests

**Ejemplos:**
```bash
# Todos los tests con salida detallada
./scripts/test-backend.sh --verbose

# Tests de un módulo específico
./scripts/test-backend.sh --module utils
./scripts/test-backend.sh --module models

# Con cobertura
./scripts/test-backend.sh --coverage

# Solo verificar código (go vet)
./scripts/test-backend.sh --vet-only

# Salida mínima
./scripts/test-backend.sh --quiet
```

## Uso Combinado

### Ejecutar ambos

```bash
# Desde la raíz del proyecto
./test-all.sh
```

Este script ejecuta ambos scripts automáticamente.

### Ejecutar manualmente

```bash
# Backend primero
./scripts/test-backend.sh --verbose

# Luego frontend
./scripts/test-frontend.sh --run
```

## Características

### test-frontend.sh
- ✅ Verifica que npm y node estén instalados
- ✅ Instala dependencias automáticamente si faltan
- ✅ Ejecuta linter antes de los tests
- ✅ Soporta múltiples modos de ejecución
- ✅ Mensajes de error claros y coloridos
- ✅ Manejo de errores robusto

### test-backend.sh
- ✅ Verifica que Go esté instalado
- ✅ Descarga dependencias automáticamente
- ✅ Ejecuta `go vet` antes de los tests
- ✅ Soporta tests de módulos específicos
- ✅ Genera reportes de cobertura
- ✅ Lista módulos disponibles al finalizar
- ✅ Mensajes de error claros y coloridos

### `test-backend-docker.sh`

Ejecuta los tests del backend usando Docker (útil cuando la versión de Go instalada es incompatible).

**Uso:**
```bash
./scripts/test-backend-docker.sh
```

**Opciones:**
- `--coverage` - Generar reporte de cobertura
- `--module <nombre>` - Ejecutar tests de un módulo específico
- `--quiet` - Salida mínima

**Ejemplo:**
```bash
# Todos los tests
./scripts/test-backend-docker.sh

# Con cobertura
./scripts/test-backend-docker.sh --coverage

# Módulo específico
./scripts/test-backend-docker.sh --module utils
```

## Troubleshooting

### Error: "malformed module path" o "cannot load io/fs"

Este error indica que tu versión de Go es demasiado antigua (probablemente < 1.16).

**Solución 1: Actualizar Go**
```bash
# Opción A: Descargar desde https://go.dev/dl/
# Opción B: Usar g (gestor de versiones)
go install github.com/voidint/g@latest
g install 1.24.0
g use 1.24.0
```

**Solución 2: Usar Docker (recomendado si no puedes actualizar Go)**
```bash
./scripts/test-backend-docker.sh
```

### Error: "Permission denied"

```bash
chmod +x scripts/test-frontend.sh scripts/test-backend.sh scripts/test-backend-docker.sh
```

### Error: "npm not found"

Instala Node.js y npm:
```bash
sudo apt install npm
# O usa nvm (recomendado)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
```

### Error: "go not found"

Instala Go desde https://go.dev/dl/

### Error: "Cannot find module"

**Frontend:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**Backend:**
```bash
cd backend
go mod tidy
go mod download
```

## Integración con CI/CD

Estos scripts están diseñados para funcionar tanto localmente como en CI/CD.

**Para GitHub Actions:**
```yaml
- name: Test Backend
  run: ./scripts/test-backend.sh --verbose

- name: Test Frontend
  run: ./scripts/test-frontend.sh --run
```

## Próximos Pasos

- [ ] Agregar opción para ejecutar tests en paralelo
- [ ] Agregar opción para filtrar tests por nombre
- [ ] Agregar opción para generar reportes HTML
- [ ] Agregar opción para comparar cobertura entre commits

