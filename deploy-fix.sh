#!/bin/bash
set -e

echo "=========================================="
echo "🔍 Verificando HPA y desplegando fix"
echo "=========================================="
echo ""

# 1. Verificar que el HPA existe
echo "1️⃣ Verificando que el HPA existe..."
if kubectl get hpa dkonsole-hpa -n dkonsole &>/dev/null; then
    echo "✅ HPA encontrado"
    echo ""
    echo "📋 Información del HPA:"
    kubectl get hpa dkonsole-hpa -n dkonsole -o yaml | grep -E "apiVersion|kind|name:" | head -5
    echo ""
else
    echo "❌ HPA no encontrado. ¿Quieres crearlo desde dkonsole-hpa.yaml?"
    read -p "Crear HPA? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl apply -f dkonsole-hpa.yaml
        echo "✅ HPA creado"
    else
        echo "⚠️  Continuando sin crear HPA..."
    fi
    echo ""
fi

# 2. Verificar imagen actual
echo "2️⃣ Verificando imagen actual del deployment..."
CURRENT_IMAGE=$(kubectl get deployment dkonsole -n dkonsole -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "not found")
echo "   Imagen actual: $CURRENT_IMAGE"
echo ""

# 3. Verificar si la nueva imagen está disponible localmente
echo "3️⃣ Verificando imagen local..."
if docker images | grep -q "dkonsole/dkonsole.*1.1.0"; then
    echo "✅ Imagen local encontrada"
    echo ""
    echo "4️⃣ ¿Quieres hacer push a Docker Hub? (requiere docker login)"
    read -p "Hacer push? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📤 Haciendo push..."
        docker push dkonsole/dkonsole:1.1.0
        docker push dkonsole/dkonsole:latest
        echo "✅ Push completado"
    fi
    echo ""
else
    echo "❌ Imagen local no encontrada. Ejecuta ./build.sh primero"
    exit 1
fi

# 5. Actualizar deployment
echo "5️⃣ Actualizando deployment..."
cd /home/flaucha/repos/gitops/apps/dkonsole 2>/dev/null || cd /home/flaucha/repos/DKonsole/helm/dkonsole

if [ -f "values.yaml" ]; then
    echo "   Usando Helm chart en $(pwd)"
    helm upgrade dkonsole . -n dkonsole --set image.tag=1.1.0 --set image.repository=dkonsole/dkonsole
    echo "✅ Deployment actualizado"
else
    echo "   Helm chart no encontrado, forzando redeploy..."
    kubectl set image deployment/dkonsole dkonsole=dkonsole/dkonsole:1.1.0 -n dkonsole
    kubectl rollout restart deployment/dkonsole -n dkonsole
    echo "✅ Redeploy iniciado"
fi
echo ""

# 6. Esperar rollout
echo "6️⃣ Esperando rollout..."
kubectl rollout status deployment/dkonsole -n dkonsole --timeout=120s
echo ""

# 7. Verificar nueva imagen
echo "7️⃣ Verificando nueva imagen..."
NEW_IMAGE=$(kubectl get deployment dkonsole -n dkonsole -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "not found")
echo "   Nueva imagen: $NEW_IMAGE"
echo ""

# 8. Mostrar logs
echo "8️⃣ Mostrando últimos logs (Ctrl+C para salir)..."
echo "   Intenta ver el YAML del HPA ahora y observa los logs:"
echo ""
kubectl logs -n dkonsole deployment/dkonsole --tail=20 -f


