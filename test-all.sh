#!/bin/bash

# Script para ejecutar todos los tests de DKonsole
# Uso: ./test-all.sh

set -e  # Salir si algún comando falla

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

# Tests del Backend
echo "🔧 Ejecutando tests del backend..."
echo "-----------------------------------"
if ./scripts/test-backend.sh --verbose; then
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
    if ./scripts/test-frontend.sh --run; then
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

