# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)

**DKonsole** es un dashboard moderno y ligero para Kubernetes, construido enteramente con **Inteligencia Artificial**. Proporciona una interfaz intuitiva para gestionar los recursos de tu cluster, ver logs y ejecutar comandos en pods.

## 🤖 Creado con IA

Todo este proyecto, desde el backend hasta el frontend y el código de infraestructura, fue generado utilizando agentes de IA avanzados. Demuestra el poder de la IA en el desarrollo de software moderno.

## 🚀 Inicio Rápido

### 1. Desplegar con Helm

```bash
# Clonar el repositorio
git clone https://github.com/flaucha/DKonsole.git
cd DKonsole

# Checkout de la última versión estable
git checkout v1.0.0

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
  hosts:
    - host: dkonsole.ejemplo.com

# Opcional: Restringir orígenes de WebSocket (CORS)
allowedOrigins: "https://dkonsole.ejemplo.com"
```

### 3. Imágenes Docker (Opcional)
Por defecto usa las imágenes oficiales. Puedes cambiar tags o repositorios si es necesario.

```yaml
image:
  backend:
    tag: "1.0.0"
```

### 2. Imágenes Docker

Las imágenes oficiales están disponibles en:

- **Backend**: `dkonsole/dkonsole-backend`
- **Frontend**: `dkonsole/dkonsole-frontend`

## 💰 Apoya el Proyecto

Si encuentras útil este proyecto, considera hacer una donación para apoyar el desarrollo.

**Billetera BSC (Binance Smart Chain):**
`0x9baf648fa316030e12b15cbc85278fdbd82a7d20`

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

## Licencia

Licencia MIT
