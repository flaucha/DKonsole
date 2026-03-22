#!/bin/bash

# Script alternativo para ejecutar tests del backend usando Docker
# Útil cuando la versión de Go instalada es incompatible
# Uso: ./scripts/test-backend-docker.sh [opciones]

set -e

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Verificar que Docker está instalado
if ! command -v docker &> /dev/null; then
    error "Docker no está instalado"
    echo "Instala Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Verificar que estamos en la raíz del proyecto
if [ ! -d "backend" ]; then
    error "Este script debe ejecutarse desde la raíz del proyecto DKonsole"
    exit 1
fi

info "Usando Docker para ejecutar tests del backend"
info "Imagen: golang:1.25.8-alpine3.22"
echo ""

# Variables
GO_VERSION="1.25.8"
IMAGE="golang:${GO_VERSION}-alpine3.22"
COVERAGE=false
MODULE=""
VERBOSE=true

# Procesar argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage)
            COVERAGE=true
            shift
            ;;
        --module)
            if [ -z "$2" ]; then
                error "--module requiere un nombre de módulo"
                exit 1
            fi
            MODULE="$2"
            shift 2
            ;;
        --quiet)
            VERBOSE=false
            shift
            ;;
        *)
            error "Opción desconocida: $1"
            echo "Opciones: --coverage, --module <nombre>, --quiet"
            exit 1
            ;;
    esac
done

# Construir comando
TEST_CMD="go test"

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
fi

if [ -n "$MODULE" ]; then
    TEST_CMD="$TEST_CMD ./internal/$MODULE/..."
else
    TEST_CMD="$TEST_CMD ./..."
fi

# Ejecutar en Docker
echo "========================================"
info "Ejecutando tests en contenedor Docker..."
echo "========================================"

docker run --rm \
    -v "$(pwd)/backend:/app" \
    -w /app \
    "$IMAGE" \
    sh -c "
        echo '📦 Descargando dependencias...' &&
        go mod download &&
        echo '' &&
        echo '🔍 Ejecutando go vet...' &&
        go vet ./... || echo '⚠️  go vet encontró algunos problemas' &&
        echo '' &&
        echo '🧪 Ejecutando tests...' &&
        $TEST_CMD
    "

if [ $? -eq 0 ]; then
    success "Tests completados exitosamente"

    if [ "$COVERAGE" = true ]; then
        info "Reporte de cobertura generado en: backend/coverage.out"
    fi
else
    error "Tests fallaron"
    exit 1
fi
