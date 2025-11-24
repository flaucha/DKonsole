# Instrucciones para Actualizar Go

## Problema
Tu sistema tiene Go 1.13.8, pero DKonsole requiere Go 1.24+.

## Solución Rápida

### Opción 1: Usar el script (requiere sudo)

```bash
# Ejecutar el script de actualización
sudo ./scripts/update-go.sh
```

### Opción 2: Instalación Manual

```bash
# 1. Descargar Go 1.24.4 (ya descargado en /tmp)
cd /tmp
ls -lh go1.24.4.linux-amd64.tar.gz

# 2. Eliminar instalación anterior
sudo rm -rf /usr/local/go

# 3. Extraer Go
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# 4. Verificar instalación
/usr/local/go/bin/go version
# Debe mostrar: go version go1.24.4 linux/amd64

# 5. Agregar al PATH (ya está en .bashrc, pero para esta sesión):
export PATH=$PATH:/usr/local/go/bin

# 6. Verificar que funciona
go version
# Debe mostrar: go version go1.24.4 linux/amd64
```

### Opción 3: Usar Docker (sin actualizar Go del sistema)

```bash
# Ejecutar tests con Docker
./scripts/test-backend-docker.sh
```

## Verificación

Después de actualizar, verifica:

```bash
# Verificar versión
go version
# Debe mostrar go1.24.x

# Probar que funciona
cd backend
go mod download
go vet ./...
```

## Nota Importante

Si actualizas Go, **cierra y abre una nueva terminal** para que el PATH se actualice correctamente, o ejecuta:

```bash
source ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
```

