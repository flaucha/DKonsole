#!/bin/bash

# Script para actualizar Go a la versión requerida por DKonsole
# Uso: ./scripts/update-go.sh [versión]
# Si no se especifica versión, usa 1.24.4 (última 1.24.x)

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

# Versión por defecto
GO_VERSION="${1:-1.24.4}"
ARCH="linux-amd64"
GO_URL="https://go.dev/dl/go${GO_VERSION}.${ARCH}.tar.gz"
INSTALL_DIR="/usr/local"

info "Actualizando Go a versión ${GO_VERSION}"
echo ""

# Verificar si ya está instalada la versión correcta
if command -v go &> /dev/null; then
    CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [ "$CURRENT_VERSION" = "go${GO_VERSION}" ]; then
        success "Go ${GO_VERSION} ya está instalado"
        go version
        exit 0
    else
        info "Versión actual: ${CURRENT_VERSION}"
        info "Actualizando a: go${GO_VERSION}"
    fi
fi

# Verificar permisos de sudo
if ! sudo -n true 2>/dev/null; then
    info "Se necesitarán permisos de sudo para instalar Go"
fi

# Crear directorio temporal
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Descargar Go
info "Descargando Go ${GO_VERSION}..."
if wget --progress=bar:force:noscroll "$GO_URL" -O "go${GO_VERSION}.${ARCH}.tar.gz" 2>&1 | grep -E "(saved|100%)"; then
    success "Descarga completada"
else
    error "Error al descargar Go"
    rm -rf "$TMP_DIR"
    exit 1
fi

# Verificar que el archivo se descargó correctamente
if [ ! -f "go${GO_VERSION}.${ARCH}.tar.gz" ]; then
    error "El archivo descargado no existe"
    rm -rf "$TMP_DIR"
    exit 1
fi

# Eliminar instalación anterior
info "Eliminando instalación anterior de Go..."
sudo rm -rf "${INSTALL_DIR}/go"

# Extraer Go
info "Extrayendo Go a ${INSTALL_DIR}..."
sudo tar -C "$INSTALL_DIR" -xzf "go${GO_VERSION}.${ARCH}.tar.gz"

# Limpiar
rm -rf "$TMP_DIR"

# Verificar instalación
if [ -f "${INSTALL_DIR}/go/bin/go" ]; then
    NEW_VERSION=$("${INSTALL_DIR}/go/bin/go" version | awk '{print $3}')
    success "Go instalado: ${NEW_VERSION}"
else
    error "La instalación de Go falló"
    exit 1
fi

# Configurar PATH en .bashrc si no está
if ! grep -q '/usr/local/go/bin' ~/.bashrc 2>/dev/null; then
    info "Agregando Go al PATH en ~/.bashrc..."
    echo "" >> ~/.bashrc
    echo "# Go" >> ~/.bashrc
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    success "PATH configurado en ~/.bashrc"
    warning "Ejecuta 'source ~/.bashrc' o abre una nueva terminal para usar Go"
fi

# Configurar PATH para esta sesión
export PATH=$PATH:/usr/local/go/bin

# Verificar que funciona
echo ""
info "Verificando instalación..."
if go version; then
    echo ""
    success "Go ${GO_VERSION} instalado y funcionando correctamente"
    echo ""
    info "Para usar Go en esta sesión, ejecuta:"
    echo "  export PATH=\$PATH:/usr/local/go/bin"
    echo ""
    info "O simplemente abre una nueva terminal"
else
    error "Go no está disponible en el PATH"
    info "Ejecuta: export PATH=\$PATH:/usr/local/go/bin"
fi





