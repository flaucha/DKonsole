#!/bin/bash

# Script para ejecutar tests del backend de DKonsole
# Uso: ./scripts/test-backend.sh [opciones]
# Opciones:
#   --verbose    Mostrar salida detallada (por defecto)
#   --quiet      Salida mínima
#   --coverage   Generar reporte de cobertura
#   --module     Ejecutar tests de un módulo específico (ej: --module utils)
#   --vet-only   Solo ejecutar go vet
#   --no-vet     No ejecutar go vet

set -e  # Salir si algún comando falla

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para mostrar mensajes
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

# Verificar que estamos en la raíz del proyecto
if [ ! -d "backend" ]; then
    error "Este script debe ejecutarse desde la raíz del proyecto DKonsole"
    exit 1
fi

# Verificar que Go está instalado
if ! command -v go &> /dev/null; then
    error "Go no está instalado. Por favor instálalo primero:"
    echo "  Visita https://go.dev/dl/ para descargar Go"
    exit 1
fi

# Verificar versión de Go
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d'.' -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d'.' -f2)

# Verificar versión mínima requerida (Go 1.16+ para io/fs, pero el proyecto requiere 1.24)
if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 16 ]); then
    error "Go versión $GO_VERSION detectada. Se requiere Go 1.16 o superior (el proyecto requiere Go 1.24)"
    echo ""
    echo "Opciones para actualizar Go:"
    echo "  1. Descargar desde https://go.dev/dl/"
    echo "  2. Usar g (gestor de versiones): go install github.com/voidint/g@latest"
    echo "  3. Usar Docker: docker run --rm -v \$(pwd)/backend:/app -w /app golang:1.24-alpine go test ./..."
    exit 1
elif [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 24 ]; then
    warning "Go versión $GO_VERSION detectada. Se recomienda Go 1.24 o superior"
    warning "Algunas características pueden no funcionar correctamente"
fi

info "Go versión: $(go version)"
echo ""

# Cambiar al directorio backend
cd backend

# Variables por defecto
VERBOSE=true
RUN_VET=true
RUN_TESTS=true
GENERATE_COVERAGE=false
MODULE=""

# Procesar argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE=true
            shift
            ;;
        --quiet)
            VERBOSE=false
            shift
            ;;
        --coverage)
            GENERATE_COVERAGE=true
            shift
            ;;
        --module)
            if [ -z "$2" ]; then
                error "--module requiere un nombre de módulo (ej: utils, models, auth)"
                exit 1
            fi
            MODULE="$2"
            shift 2
            ;;
        --vet-only)
            RUN_TESTS=false
            shift
            ;;
        --no-vet)
            RUN_VET=false
            shift
            ;;
        *)
            error "Opción desconocida: $1"
            echo "Opciones disponibles: --verbose, --quiet, --coverage, --module <nombre>, --vet-only, --no-vet"
            exit 1
            ;;
    esac
done

# Actualizar go.mod si es necesario
echo "========================================"
info "Actualizando go.mod..."
echo "========================================"
if go mod tidy 2>&1 | tee /tmp/gomod_tidy.txt; then
    success "go.mod actualizado"
else
    warning "go mod tidy encontró algunos problemas (continuando...)"
fi
echo ""

# Descargar dependencias
echo "========================================"
info "Descargando dependencias de Go..."
echo "========================================"
if go mod download 2>&1 | tee /tmp/gomod_output.txt; then
    success "Dependencias descargadas"
else
    # Verificar si el error es por versión de Go
    if grep -q "malformed module path\|cannot load\|missing dot" /tmp/gomod_output.txt; then
        error "Error al descargar dependencias debido a versión incompatible de Go"
        error "Tu versión de Go ($GO_VERSION) es demasiado antigua"
        error "Se requiere Go 1.16+ (recomendado: Go 1.24+)"
        echo ""
        echo "Soluciones:"
        echo "  1. Actualizar Go: https://go.dev/dl/"
        echo "  2. Usar g: go install github.com/voidint/g@latest && g install 1.24.0"
        echo "  3. Usar Docker (ver scripts/README_TESTS.md)"
        rm -f /tmp/gomod_output.txt
        exit 1
    else
        error "Error al descargar dependencias"
        rm -f /tmp/gomod_output.txt
        exit 1
    fi
fi
rm -f /tmp/gomod_output.txt
echo ""

# Ejecutar go vet
if [ "$RUN_VET" = true ]; then
    echo "========================================"
    info "Ejecutando go vet..."
    echo "========================================"
    
    if [ -n "$MODULE" ]; then
        VET_PATH="./internal/$MODULE/..."
    else
        VET_PATH="./..."
    fi
    
    # Intentar ejecutar go vet, pero no fallar si hay problemas de versión
    if go vet $VET_PATH 2>&1 | tee /tmp/govet_output.txt; then
        success "go vet pasó sin errores"
    else
        # Verificar si el error es por versión de Go
        if grep -q "malformed module path\|cannot load\|missing dot" /tmp/govet_output.txt; then
            error "go vet falló debido a versión incompatible de Go"
            warning "Tu versión de Go ($GO_VERSION) es demasiado antigua para este proyecto"
            warning "Se requiere Go 1.16+ (recomendado: Go 1.24+)"
            echo ""
            info "Omitiendo go vet y continuando con los tests..."
            warning "Los tests pueden fallar si la versión de Go es incompatible"
        else
            warning "go vet encontró algunos problemas (esto no detiene la ejecución)"
        fi
    fi
    rm -f /tmp/govet_output.txt
    echo ""
fi

# Ejecutar tests
if [ "$RUN_TESTS" = true ]; then
    echo "========================================"
    if [ -n "$MODULE" ]; then
        info "Ejecutando tests del módulo: $MODULE"
    else
        info "Ejecutando todos los tests del backend..."
    fi
    echo "========================================"
    
    # Construir comando de test
    TEST_CMD="go test"
    
    if [ "$VERBOSE" = true ]; then
        TEST_CMD="$TEST_CMD -v"
    fi
    
    if [ "$GENERATE_COVERAGE" = true ]; then
        TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
        info "Generando reporte de cobertura..."
    fi
    
    if [ -n "$MODULE" ]; then
        TEST_PATH="./internal/$MODULE/..."
    else
        TEST_PATH="./..."
    fi
    
    TEST_CMD="$TEST_CMD $TEST_PATH"
    
    if eval $TEST_CMD; then
        success "Todos los tests pasaron"
        
        # Mostrar cobertura si se generó
        if [ "$GENERATE_COVERAGE" = true ]; then
            echo ""
            info "Cobertura de código generada en: backend/coverage.out"
            info "Para ver el reporte HTML, ejecuta: go tool cover -html=coverage.out"
            
            # Mostrar resumen de cobertura
            if command -v go tool cover &> /dev/null; then
                echo ""
                info "Resumen de cobertura:"
                go tool cover -func=coverage.out | tail -1
            fi
        fi
    else
        error "Algunos tests fallaron"
        exit 1
    fi
fi

echo ""
echo "========================================"
success "Tests del backend completados"
echo "========================================"

# Mostrar módulos disponibles si se ejecutó sin especificar módulo
if [ -z "$MODULE" ] && [ "$RUN_TESTS" = true ]; then
    echo ""
    info "Módulos con tests disponibles:"
    find ./internal -name "*_test.go" -type f | sed 's|.*/\([^/]*\)/[^/]*_test.go|\1|' | sort -u | while read mod; do
        echo "  - $mod (ejecutar: ./scripts/test-backend.sh --module $mod)"
    done
fi

