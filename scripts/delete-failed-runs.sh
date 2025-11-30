#!/bin/bash
# Script para eliminar todos los workflow runs fallidos o cancelados
# Requiere: GitHub CLI (gh) instalado y autenticado
#
# Uso:
#   ./delete-failed-runs.sh          # Modo interactivo (pide confirmación)
#   ./delete-failed-runs.sh --yes    # Modo no interactivo (elimina sin confirmar)

set -e

REPO="flaucha/DKonsole"
SKIP_CONFIRM=false

# Verificar si se pasó --yes
if [ "$1" == "--yes" ] || [ "$1" == "-y" ]; then
    SKIP_CONFIRM=true
fi

echo "=========================================="
echo "🔍 Buscando workflow runs fallidos o cancelados..."
echo "=========================================="

# Obtener lista de runs fallidos y cancelados
FAILED_RUNS=$(gh run list --repo "$REPO" --limit 100 --json databaseId,status,conclusion,number,workflowName,createdAt --jq '.[] | select(.conclusion == "failure" or .conclusion == "cancelled") | "\(.databaseId)|\(.number)|\(.workflowName)|\(.conclusion)|\(.createdAt)"')

if [ -z "$FAILED_RUNS" ]; then
    echo "✅ No hay runs fallidos o cancelados para eliminar"
    exit 0
fi

# Contar runs
TOTAL=$(echo "$FAILED_RUNS" | wc -l)
FAILED_COUNT=$(echo "$FAILED_RUNS" | grep -c "failure" || echo "0")
CANCELLED_COUNT=$(echo "$FAILED_RUNS" | grep -c "cancelled" || echo "0")

echo ""
echo "📋 Encontrados $TOTAL run(s):"
echo "   - Fallidos: $FAILED_COUNT"
echo "   - Cancelados: $CANCELLED_COUNT"
echo ""
echo "Detalles:"
echo "$FAILED_RUNS" | while IFS='|' read -r dbId number workflow conclusion createdAt; do
    echo "  - Run #$number ($workflow): $conclusion (ID: $dbId, Fecha: $createdAt)"
done

if [ "$SKIP_CONFIRM" = false ]; then
    echo ""
    read -p "¿Eliminar todos estos runs? (s/N): " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        echo "❌ Operación cancelada"
        exit 1
    fi
else
    echo ""
    echo "⚠️  Modo no interactivo: se eliminarán todos los runs sin confirmación"
fi

echo ""
echo "🗑️  Eliminando runs..."

SUCCESS=0
FAILED=0

while IFS='|' read -r dbId number workflow conclusion createdAt; do
    if gh run delete "$dbId" --repo "$REPO" 2>/dev/null; then
        echo "  ✅ Eliminado run #$number ($workflow)"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "  ⚠️  Error al eliminar run #$number ($workflow)"
        FAILED=$((FAILED + 1))
    fi
done <<< "$FAILED_RUNS"

echo ""
echo "✅ Proceso completado"
echo "   - Eliminados: $SUCCESS"
echo "   - Errores: $FAILED"
