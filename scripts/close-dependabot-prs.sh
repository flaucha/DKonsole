#!/bin/bash
# Script para cerrar todos los PRs abiertos de Dependabot
# Requiere: GitHub CLI (gh) instalado y autenticado

set -e

REPO="flaucha/DKonsole"

echo "=========================================="
echo "🔍 Buscando PRs abiertos de Dependabot..."
echo "=========================================="

# Obtener lista de PRs abiertos de dependabot
PRS=$(gh pr list --repo "$REPO" --author "app/dependabot" --state open --json number,title --jq '.[] | "\(.number)|\(.title)"')

if [ -z "$PRS" ]; then
    echo "✅ No hay PRs abiertos de Dependabot"
    exit 0
fi

echo ""
echo "📋 PRs encontrados:"
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

echo "$PRS" | while IFS='|' read -r number title; do
    echo "  Cerrando PR #$number: $title"
    gh pr close "$number" --repo "$REPO" --comment "Cerrado automáticamente. Los PRs de dependabot se recrearán según el schedule configurado." || echo "  ⚠️  Error al cerrar PR #$number"
done

echo ""
echo "✅ Proceso completado"
