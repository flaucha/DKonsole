#!/bin/bash

# Script para ejecutar todos los tests de DKonsole
# Uso: ./scripts/test-all.sh

set -e  # Salir si algún comando falla

# Obtener el directorio del script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🧪 DKonsole - Ejecutando todos los tests"
echo "========================================"
echo ""

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para verificar si un comando existe
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Verificar prerrequisitos
echo "📋 Verificando prerrequisitos..."
if ! command_exists go; then
    echo -e "${RED}❌ Go no está instalado${NC}"
    exit 1
fi

if ! command_exists npm; then
    echo -e "${YELLOW}⚠️  npm no está instalado. Los tests del frontend se omitirán.${NC}"
    SKIP_FRONTEND=true
else
    SKIP_FRONTEND=false
fi

echo -e "${GREEN}✅ Prerrequisitos verificados${NC}"
echo ""

# Cambiar al directorio raíz del proyecto
cd "$PROJECT_ROOT"

# Tests del Backend
echo "🔧 Ejecutando tests del backend..."
echo "-----------------------------------"
if "$SCRIPT_DIR/test-backend.sh" --verbose; then
    echo -e "${GREEN}✅ Tests del backend pasaron${NC}"
else
    echo -e "${RED}❌ Tests del backend fallaron${NC}"
    exit 1
fi
echo ""

# Tests del Frontend
if [ "$SKIP_FRONTEND" = false ]; then
    echo "⚛️  Ejecutando tests del frontend..."
    echo "-----------------------------------"
    if "$SCRIPT_DIR/test-frontend.sh" --run; then
        echo -e "${GREEN}✅ Tests del frontend pasaron${NC}"
    else
        echo -e "${RED}❌ Tests del frontend fallaron${NC}"
        exit 1
    fi
    echo ""
else
    echo -e "${YELLOW}⏭️  Omitiendo tests del frontend (npm no disponible)${NC}"
    echo ""
fi

# Resumen
echo "========================================"
echo -e "${GREEN}✅ Todos los tests completados${NC}"
echo ""
echo "Para más información, consulta TESTING.md"




