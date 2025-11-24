# Guía de Testing para DKonsole

Esta guía explica cómo ejecutar los tests unitarios tanto del frontend como del backend.

## Prerrequisitos

- Node.js 20+ (para frontend)
- Go 1.24+ (para backend)
- npm o yarn

## Frontend

### 1. Instalar dependencias

```bash
cd frontend
npm install
```

### 2. Ejecutar tests

**Modo watch (recomendado para desarrollo):**
```bash
npm run test
```

**Ejecutar una vez y salir:**
```bash
npm run test -- --run
```

**Con interfaz gráfica:**
```bash
npm run test:ui
```

**Con cobertura de código:**
```bash
npm run test:coverage
```

### 3. Ejecutar linter

```bash
npm run lint
```

## Backend

### 1. Instalar dependencias

```bash
cd backend
go mod download
```

### 2. Ejecutar tests

**Todos los tests:**
```bash
go test ./...
```

**Tests con verbosidad:**
```bash
go test -v ./...
```

**Tests con cobertura:**
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Tests de un módulo específico:**
```bash
go test ./internal/utils/...
go test ./internal/models/...
```

### 3. Verificar código con go vet

```bash
go vet ./...
```

## Ejecutar todo

### Script para ejecutar todos los tests

```bash
#!/bin/bash
# test-all.sh

echo "🧪 Ejecutando tests del backend..."
cd backend
go test -v ./...
if [ $? -ne 0 ]; then
    echo "❌ Tests del backend fallaron"
    exit 1
fi

echo ""
echo "🧪 Ejecutando tests del frontend..."
cd ../frontend
npm run test -- --run
if [ $? -ne 0 ]; then
    echo "❌ Tests del frontend fallaron"
    exit 1
fi

echo ""
echo "✅ Todos los tests pasaron!"
```

## GitHub Actions

El workflow de CI/CD se ejecuta automáticamente cuando:
- Se hace push a la rama `main`
- Se crea un Pull Request hacia `main`

### Verificar el workflow localmente

Puedes usar [act](https://github.com/nektos/act) para ejecutar GitHub Actions localmente:

```bash
# Instalar act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Ejecutar el workflow
act push
```

## Estructura de Tests

### Frontend

```
frontend/src/
├── utils/
│   └── __tests__/
│       ├── dateUtils.test.js
│       ├── resourceParser.test.js
│       ├── statusBadge.test.js
│       └── expandableRow.test.js
├── api/
│   └── __tests__/
│       └── k8sApi.test.js
└── test/
    └── setup.js
```

### Backend

```
backend/internal/
├── utils/
│   └── utils_test.go
├── models/
│   └── models_test.go
├── auth/
│   └── auth_test.go (pendiente)
├── cluster/
│   └── cluster_test.go (pendiente)
└── ...
```

## Troubleshooting

### Frontend: Error "Cannot find module"

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Backend: Error de dependencias

```bash
cd backend
go mod tidy
go mod download
```

### Frontend: Tests no encuentran módulos

Verifica que `vite.config.js` tenga la configuración de test correcta.

### Backend: Tests fallan por falta de mocks

Algunos tests pueden requerir mocks de clientes de Kubernetes. Revisa la documentación de testing de Go para crear mocks.

## Próximos Pasos

- [ ] Agregar más tests para componentes React
- [ ] Agregar tests para hooks personalizados
- [ ] Agregar tests para contextos
- [ ] Agregar tests para módulos del backend restantes
- [ ] Configurar coverage thresholds
- [ ] Agregar tests de integración

