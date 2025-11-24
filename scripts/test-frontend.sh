#!/bin/bash

# Script para ejecutar tests del frontend de DKonsole
# Uso: ./scripts/test-frontend.sh [opciones]
# Opciones:
#   --watch    Ejecutar en modo watch (por defecto)
#   --run      Ejecutar una vez y salir
#   --ui       Ejecutar con interfaz gráfica
#   --coverage Ejecutar con cobertura de código
#   --lint     Solo ejecutar linter

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
if [ ! -d "frontend" ]; then
    error "Este script debe ejecutarse desde la raíz del proyecto DKonsole"
    exit 1
fi

# Verificar que npm está instalado
if ! command -v npm &> /dev/null; then
    error "npm no está instalado. Por favor instálalo primero:"
    echo "  sudo apt install npm"
    echo "  o usa nvm: curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
    exit 1
fi

# Verificar que node está instalado
if ! command -v node &> /dev/null; then
    error "node no está instalado"
    exit 1
fi

# Verificar versión de Node.js
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    warning "Node.js versión $NODE_VERSION detectada. Se recomienda Node.js 20 o superior"
fi

info "Node.js versión: $(node -v)"
info "npm versión: $(npm -v)"
echo ""

# Cambiar al directorio frontend
cd frontend

# Verificar si node_modules existe
if [ ! -d "node_modules" ]; then
    info "Instalando dependencias..."
    npm install
    success "Dependencias instaladas"
    echo ""
fi

# Procesar argumentos
MODE="watch"
RUN_LINT=true
RUN_TESTS=true

while [[ $# -gt 0 ]]; do
    case $1 in
        --watch)
            MODE="watch"
            shift
            ;;
        --run)
            MODE="run"
            shift
            ;;
        --ui)
            MODE="ui"
            shift
            ;;
        --coverage)
            MODE="coverage"
            shift
            ;;
        --lint-only)
            RUN_TESTS=false
            shift
            ;;
        --no-lint)
            RUN_LINT=false
            shift
            ;;
        *)
            error "Opción desconocida: $1"
            echo "Opciones disponibles: --watch, --run, --ui, --coverage, --lint-only, --no-lint"
            exit 1
            ;;
    esac
done

# Ejecutar linter
if [ "$RUN_LINT" = true ]; then
    echo "========================================"
    info "Ejecutando linter..."
    echo "========================================"
    if npm run lint; then
        success "Linter pasó sin errores"
    else
        warning "El linter encontró algunos problemas (esto no detiene la ejecución)"
    fi
    echo ""
fi

# Ejecutar tests
if [ "$RUN_TESTS" = true ]; then
    echo "========================================"
    info "Ejecutando tests del frontend..."
    echo "========================================"
    
    case $MODE in
        watch)
            info "Modo: Watch (presiona 'q' para salir)"
            npm run test
            ;;
        run)
            info "Modo: Ejecutar una vez"
            if npm run test -- --run; then
                success "Todos los tests pasaron"
            else
                error "Algunos tests fallaron"
                exit 1
            fi
            ;;
        ui)
            info "Modo: Interfaz gráfica"
            info "Abre tu navegador en la URL que se mostrará"
            npm run test:ui
            ;;
        coverage)
            info "Modo: Con cobertura de código"
            if npm run test:coverage; then
                success "Tests ejecutados con cobertura"
                info "Revisa el reporte de cobertura en frontend/coverage/"
            else
                error "Algunos tests fallaron"
                exit 1
            fi
            ;;
    esac
fi

echo ""
echo "========================================"
success "Tests del frontend completados"
echo "========================================"


