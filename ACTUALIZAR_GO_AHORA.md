# ⚡ Actualizar Go Ahora - Instrucciones Rápidas

## 🚀 Opción Rápida (Recomendada)

Ejecuta este comando en tu terminal:

```bash
sudo ./scripts/install-go.sh
```

## 📋 O Manualmente

Copia y pega estos comandos en tu terminal:

```bash
# 1. Eliminar instalación anterior
sudo rm -rf /usr/local/go

# 2. Extraer Go 1.24.4
sudo tar -C /usr/local -xzf /tmp/go1.24.4.linux-amd64.tar.gz

# 3. Verificar
/usr/local/go/bin/go version

# 4. Agregar al PATH para esta sesión
export PATH=$PATH:/usr/local/go/bin

# 5. Verificar que funciona
go version
```

## ✅ Verificación

Después de instalar, verifica:

```bash
# Debe mostrar: go version go1.24.4 linux/amd64
go version

# Probar con el proyecto
cd /home/flaucha/repos/DKonsole/backend
go mod download
go vet ./...
```

## 🔄 Para Nuevas Terminales

El PATH ya está configurado en `~/.bashrc`. Solo necesitas:

- **Opción 1:** Abrir una nueva terminal
- **Opción 2:** Ejecutar `source ~/.bashrc`

## 🐳 Alternativa: Docker

Si no puedes usar sudo, usa Docker:

```bash
./scripts/test-backend-docker.sh
```

