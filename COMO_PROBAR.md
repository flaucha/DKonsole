# Cómo Probar DKonsole - Guía Rápida

## 🚀 Inicio Rápido

### Opción 1: Script Automático (Recomendado)

```bash
# Desde la raíz del proyecto
./test-all.sh
```

Este script ejecuta todos los tests automáticamente.

### Opción 2: Manual

## 📋 Prerrequisitos

### Verificar versiones instaladas

```bash
# Verificar Go
go version
# Debe ser Go 1.24 o superior

# Verificar Node.js
node --version
# Debe ser Node 20 o superior

# Verificar npm
npm --version
```

## 🧪 Tests del Frontend

### 1. Instalar dependencias

```bash
cd frontend
npm install
```

**Nota:** Si no tienes npm instalado:
```bash
# En Ubuntu/Debian
sudo apt install npm

# O usar nvm (recomendado)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 20
nvm use 20
```

### 2. Ejecutar tests

```bash
# Modo watch (se ejecuta automáticamente al cambiar archivos)
npm run test

# Ejecutar una vez
npm run test -- --run

# Con interfaz gráfica (muy útil)
npm run test:ui

# Con cobertura de código
npm run test:coverage
```

### 3. Ver resultados

Los tests deberían mostrar algo como:
```
✓ src/utils/__tests__/dateUtils.test.js (5)
✓ src/utils/__tests__/resourceParser.test.js (4)
✓ src/utils/__tests__/statusBadge.test.js (1)
✓ src/utils/__tests__/expandableRow.test.js (4)
✓ src/api/__tests__/k8sApi.test.js (9)

Test Files  5 passed (5)
     Tests  23 passed (23)
```

## 🔧 Tests del Backend

### 1. Verificar versión de Go

```bash
go version
```

**Si tienes una versión antigua de Go (< 1.24):**

```bash
# Opción 1: Actualizar Go manualmente
# Descargar desde https://go.dev/dl/
# O usar g (gestor de versiones de Go)
go install github.com/voidint/g@latest
g install 1.24.0
g use 1.24.0

# Opción 2: Usar Docker (ver sección Docker más abajo)
```

### 2. Instalar dependencias

```bash
cd backend
go mod download
```

### 3. Ejecutar tests

```bash
# Todos los tests
go test ./...

# Con más detalles
go test -v ./...

# Tests de un módulo específico
go test ./internal/utils/... -v
go test ./internal/models/... -v

# Con cobertura
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. Verificar código

```bash
# Verificar que no haya problemas de estilo/errores
go vet ./...

# Formatear código
go fmt ./...
```

## 🐳 Usar Docker (Alternativa si no tienes Go/Node instalados)

### Crear Dockerfile para testing

```dockerfile
# Dockerfile.test
FROM golang:1.24-alpine AS backend-test
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go test ./...

FROM node:20-alpine AS frontend-test
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run test -- --run
```

### Ejecutar tests con Docker

```bash
# Backend
docker run --rm -v $(pwd)/backend:/app -w /app golang:1.24-alpine sh -c "go mod download && go test ./..."

# Frontend
docker run --rm -v $(pwd)/frontend:/app -w /app node:20-alpine sh -c "npm ci && npm run test -- --run"
```

## 📊 Verificar GitHub Actions Localmente

### Usar Act (opcional)

```bash
# Instalar act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Ejecutar el workflow localmente
act push

# O solo un job específico
act -j test-backend
act -j test-frontend
```

## ✅ Checklist de Verificación

Antes de hacer commit, verifica:

- [ ] `go test ./...` pasa sin errores
- [ ] `go vet ./...` no muestra problemas
- [ ] `npm run test -- --run` pasa sin errores
- [ ] `npm run lint` no muestra errores críticos
- [ ] Los tests nuevos están documentados

## 🔍 Troubleshooting

### Error: "Cannot find package"

**Backend:**
```bash
cd backend
go mod tidy
go mod download
```

**Frontend:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Error: "Go version too old"

Actualiza Go o usa Docker (ver sección Docker arriba).

### Error: "npm not found"

Instala Node.js y npm:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install nodejs npm

# O usa nvm (recomendado)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 20
```

### Tests fallan pero el código funciona

Algunos tests pueden requerir mocks. Revisa los tests individuales y ajusta según sea necesario.

## 📝 Ejemplos de Salida Esperada

### Frontend (éxito)
```
✓ src/utils/__tests__/dateUtils.test.js (5) 234ms
✓ src/utils/__tests__/resourceParser.test.js (4) 123ms
✓ src/utils/__tests__/statusBadge.test.js (1) 45ms
✓ src/utils/__tests__/expandableRow.test.js (4) 67ms
✓ src/api/__tests__/k8sApi.test.js (9) 456ms

Test Files  5 passed (5)
     Tests  23 passed (23)
      Time  925ms
```

### Backend (éxito)
```
=== RUN   TestIsSystemNamespace
--- PASS: TestIsSystemNamespace (0.00s)
=== RUN   TestValidateK8sName
--- PASS: TestValidateK8sName (0.00s)
...
PASS
ok      github.com/example/k8s-view/internal/utils    0.123s
```

## 🎯 Próximos Pasos

Una vez que los tests básicos funcionen:

1. Agregar más tests para componentes React
2. Agregar tests para hooks y contextos
3. Agregar tests de integración
4. Configurar coverage thresholds
5. Agregar tests E2E con Playwright o Cypress

## 📚 Recursos

- [Documentación de Vitest](https://vitest.dev/)
- [Testing Library para React](https://testing-library.com/react)
- [Go Testing](https://go.dev/doc/effective_go#testing)
- [GitHub Actions](https://docs.github.com/en/actions)

