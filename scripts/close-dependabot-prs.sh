#!/bin/bash
# Script para cerrar todos los PRs abiertos de Dependabot
# Requiere: GitHub CLI (gh) instalado y autenticado

set -e

REPO="flaucha/DKonsole"

echo "=========================================="
echo "🔍 Verificando autenticación de GitHub..."
echo "=========================================="

# Verificar si gh está instalado
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) no está instalado"
    echo "   Instala con: sudo apt-get install gh"
    exit 1
fi

# Verificar autenticación
if ! gh auth status &> /dev/null; then
    echo "⚠️  No estás autenticado en GitHub CLI"
    echo ""
    echo "Para autenticarte, ejecuta:"
    echo "  gh auth login"
    echo ""
    echo "O establece un token:"
    echo "  export GH_TOKEN=tu_token_aqui"
    echo ""
    read -p "¿Quieres autenticarte ahora? (s/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        gh auth login
    else
        echo "❌ Operación cancelada. Autentícate primero."
        exit 1
    fi
fi

echo "✅ Autenticación verificada"
echo ""

echo "=========================================="
echo "🔍 Buscando PRs abiertos de Dependabot..."
echo "=========================================="

# Obtener lista de PRs abiertos de dependabot
PRS=$(gh pr list --repo "$REPO" --author "app/dependabot" --state open --json number,title --jq '.[] | "\(.number)|\(.title)"' 2>/dev/null || echo "")

if [ -z "$PRS" ]; then
    echo "✅ No hay PRs abiertos de Dependabot"
    exit 0
fi

PR_COUNT=$(echo "$PRS" | wc -l)
echo "📋 Encontrados $PR_COUNT PR(s) de Dependabot:"
echo ""
echo "$PRS" | while IFS='|' read -r number title; do
    echo "  - #$number: $title"
done

echo ""
read -p "¿Cerrar todos estos PRs? (s/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Ss]$ ]]; then
    echo "❌ Operación cancelada"
    exit 1
fi

echo ""
echo "🗑️  Cerrando PRs..."

SUCCESS=0
FAILED=0

echo "$PRS" | while IFS='|' read -r number title; do
    echo "  Cerrando PR #$number: $title"
    if gh pr close "$number" --repo "$REPO" --comment "Cerrado automáticamente. Los PRs de dependabot se recrearán según el schedule configurado." 2>/dev/null; then
        echo "    ✅ Cerrado"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "    ⚠️  Error al cerrar PR #$number"
        FAILED=$((FAILED + 1))
    fi
done

echo ""
echo "=========================================="
echo "✅ Proceso completado"
echo "   Exitosos: $SUCCESS"
if [ $FAILED -gt 0 ]; then
    echo "   Fallidos: $FAILED"
fi
echo "=========================================="
