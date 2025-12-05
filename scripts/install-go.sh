#!/bin/bash

# Script simple para instalar Go 1.25.5
# Ejecutar: sudo ./scripts/install-go.sh

set -e

echo "🔧 Instalando Go 1.25.5..."
echo ""

# Verificar que el archivo existe
if [ ! -f "/tmp/go1.25.5.linux-amd64.tar.gz" ]; then
    echo "❌ Error: Archivo go1.25.5.linux-amd64.tar.gz no encontrado en /tmp"
    echo "Descargando..."
    cd /tmp
    wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
fi

# Eliminar instalación anterior
echo "🗑️  Eliminando instalación anterior..."
rm -rf /usr/local/go

# Extraer Go
echo "📦 Extrayendo Go 1.25.5..."
tar -C /usr/local -xzf /tmp/go1.25.5.linux-amd64.tar.gz

# Verificar
echo ""
echo "✅ Go instalado exitosamente!"
/usr/local/go/bin/go version

echo ""
echo "📝 Para usar Go en esta sesión, ejecuta:"
echo "   export PATH=\$PATH:/usr/local/go/bin"
echo ""
echo "📝 O simplemente abre una nueva terminal (el PATH ya está en .bashrc)"
