# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)
![Version](https://img.shields.io/badge/version-1.1.0-green.svg)

**DKonsole** es un dashboard moderno y ligero para Kubernetes, construido enteramente con **Inteligencia Artificial**. Proporciona una interfaz intuitiva para gestionar los recursos de tu cluster, ver logs, ejecutar comandos en pods y monitorear métricas históricas con integración de Prometheus.

## 🤖 Creado con IA

Todo este proyecto, desde el backend hasta el frontend y el código de infraestructura, fue generado utilizando agentes de IA avanzados. Demuestra el poder de la IA en el desarrollo de software moderno.

## ✨ Características

- 🎯 **Gestión de Recursos**: Ver y gestionar Deployments, Pods, Services, ConfigMaps, Secrets y más
- 📊 **Integración con Prometheus**: Métricas históricas para Pods con rangos de tiempo personalizables (1h, 6h, 12h, 1d, 7d, 15d)
- 📝 **Logs en Tiempo Real**: Transmitir logs de contenedores en tiempo real
- 💻 **Acceso a Terminal**: Ejecutar comandos directamente en contenedores de pods
- ✏️ **Editor YAML**: Editar recursos con un editor YAML integrado
- 🔐 **Autenticación Segura**: Hash de contraseñas con Argon2 y sesiones basadas en JWT
- 🌐 **Soporte Multi-Cluster**: Gestionar múltiples clusters de Kubernetes desde una sola interfaz

## 🚀 Inicio Rápido

### 1. Desplegar con Helm

```bash
# Clonar el repositorio
git clone https://github.com/flaucha/DKonsole.git
cd DKonsole

# Checkout de la última versión estable
git checkout v1.1.0

# Instalar
helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
```

## ⚙️ Configuración

El archivo `values.yaml` está diseñado para ser simple. Solo necesitas configurar lo esencial:

### 1. Autenticación (Requerido)
Debes proporcionar un usuario `admin` y un `passwordHash` (Argon2). También necesitas un `jwtSecret` para la seguridad de la sesión.

```yaml
admin:
  username: admin
  passwordHash: "$argon2id$..." # Generar con herramienta argon2
jwtSecret: "..." # Generar con openssl rand -base64 32
```

### 2. Ingress (Requerido para acceso externo)
Configura tu dominio y ajustes TLS para acceder al dashboard.

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: dkonsole.ejemplo.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: dkonsole-tls
      hosts:
        - dkonsole.ejemplo.com

# Opcional: Restringir orígenes de WebSocket (CORS)
allowedOrigins: "https://dkonsole.ejemplo.com"
```

### 3. Integración con Prometheus (Opcional)
Habilita métricas históricas configurando tu endpoint de Prometheus.

```yaml
prometheusUrl: "http://prometheus-server.monitoring.svc.cluster.local:9090"
```

**Características habilitadas con Prometheus:**
- Métricas históricas de CPU y memoria para Pods
- Selector de rango de tiempo (1 hora, 6 horas, 12 horas, 1 día, 7 días, 15 días)
- Pestaña de métricas en la vista de detalles del Pod

**Nota:** Si `prometheusUrl` no está configurado, la pestaña de Métricas no se mostrará.

### 4. Imagen Docker (Opcional)
Por defecto usa la imagen oficial. Puedes cambiar el tag o repositorio si es necesario.

```yaml
image:
  repository: dkonsole/dkonsole
  tag: "1.1.0"
```

## 🐳 Imagen Docker

La imagen oficial está disponible en:

- **Unificada**: `dkonsole/dkonsole:1.1.0`

**Nota:** A partir de v1.1.0, DKonsole usa una arquitectura de contenedor unificada donde el backend sirve los archivos estáticos del frontend. Esto mejora la seguridad al reducir la superficie de ataque y eliminar la comunicación entre contenedores.

## 📊 Métricas de Prometheus

DKonsole se integra con Prometheus para proporcionar visualización de métricas históricas. Se utilizan las siguientes consultas PromQL:

**Uso de CPU (millicores):**
```promql
sum(rate(container_cpu_usage_seconds_total{namespace="<namespace>",pod="<pod-name>",container!=""}[5m])) * 1000
```

**Uso de Memoria (MiB):**
```promql
sum(container_memory_working_set_bytes{namespace="<namespace>",pod="<pod-name>",container!=""}) / 1024 / 1024
```

## 💰 Apoya el Proyecto

Si encuentras útil este proyecto, considera hacer una donación para apoyar el desarrollo.

**Billetera BSC (Binance Smart Chain):**
`0x9baf648fa316030e12b15cbc85278fdbd82a7d20`

**Buy me a coffee:**
https://buymeacoffee.com/flaucha

## 📧 Contacto

Para preguntas o comentarios, por favor contacta a: **flaucha@gmail.com**

## 🛠️ Desarrollo

Para ejecutar localmente:

```bash
# Backend
cd backend && go run main.go

# Frontend
cd frontend && npm run dev
```

## 📝 Changelog

### v1.1.0 (2024-12-19)
**🏗️ Lanzamiento de Arquitectura Unificada**

Este lanzamiento introduce una mejora arquitectónica importante con seguridad mejorada:

**Cambios de Arquitectura:**
- 🎯 **Contenedor Unificado**: Backend y Frontend integrados en una sola imagen Docker
- 🔒 **Seguridad Mejorada**: Superficie de ataque reducida al eliminar la comunicación entre contenedores
- 🚀 **Despliegue Simplificado**: Un solo servicio, un solo deployment, un solo puerto (8080)
- 📦 **Gestión Más Fácil**: Una imagen para construir, versionar y desplegar

**Mejoras Técnicas:**
- El backend ahora sirve los archivos estáticos del frontend directamente
- Chart de Helm simplificado con deployment unificado
- Reducción de overhead de recursos (un solo contenedor en lugar de dos)
- Configuración de ruta de ingress única

**Notas de Migración:**
- La imagen Docker cambió de `dkonsole/dkonsole-backend` y `dkonsole/dkonsole-frontend` a `dkonsole/dkonsole`
- Valores de Helm actualizados: `image.backend` e `image.frontend` reemplazados con configuración única `image`
- Rutas de ingress simplificadas: ruta única `/` en lugar de rutas separadas `/api` y `/`

### v1.0.7 (2025-11-23)
**🔒 Lanzamiento de Endurecimiento de Seguridad**

Este lanzamiento se enfoca en abordar vulnerabilidades críticas de seguridad e implementar medidas de seguridad de nivel empresarial:

**Correcciones Críticas de Seguridad:**
- 🛡️ **Prevención de Inyección PromQL**: Validación estricta de entrada para todas las consultas de Prometheus
- 🔐 **Autenticación Basada en Cookies**: Migración de localStorage a cookies HttpOnly para tokens JWT
- ✅ **Validación de Nombres de Kubernetes**: Validación RFC 1123 para todos los nombres de recursos y namespaces
- 🚫 **Sanitización de Entrada**: Validación exhaustiva en todos los endpoints de API

**Mejoras de Seguridad:**
- 📝 **Auditoría de Logs**: Middleware de auditoría completo registrando todas las solicitudes de API
- ⏱️ **Rate Limiting**: Rate limiting inteligente (300 req/min por IP)
- 🔒 **Actualizaciones CSP**: Content Security Policy mejorado
- 🔑 **Aplicación de JWT Secret**: Validación estricta requiriendo JWT_SECRET de mínimo 32 caracteres

## Licencia

Licencia MIT
