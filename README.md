# DKonsole

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![AI Generated](https://img.shields.io/badge/AI-Generated-100000?style=flat&logo=openai&logoColor=white)

**DKonsole** is a modern, lightweight Kubernetes dashboard built entirely with **Artificial Intelligence**. It provides an intuitive interface to manage your cluster resources, view logs, and execute commands in pods.

## 🤖 Built with AI

This entire project, from backend to frontend and infrastructure code, was generated using advanced AI agents. It demonstrates the power of AI in modern software development.

## 🚀 Quick Start

### 1. Deploy with Helm

```bash
# Add the repo (if applicable) or clone
git clone https://github.com/flaucha/DKonsole.git
cd DKonsole

# Install
helm install dkonsole ./helm/dkonsole -n dkonsole --create-namespace
```

## ⚙️ Configuration

The `values.yaml` file is designed to be simple. You only need to configure the essentials:

### 1. Authentication (Required)
You must provide an `admin` username and an Argon2 `passwordHash`. You also need a `jwtSecret` for session security.

```yaml
admin:
  username: admin
  passwordHash: "$argon2id$..." # Generate with argon2 tool
jwtSecret: "..." # Generate with openssl rand -base64 32
```

### 2. Ingress (Required for external access)
Configure your domain and TLS settings to access the dashboard.

```yaml
ingress:
  enabled: true
  hosts:
    - host: dkonsole.example.com

# Optional: Restrict WebSocket origins (CORS)
allowedOrigins: "https://dkonsole.example.com"
```

### 3. Docker Images (Optional)
By default, it uses the official images. You can change tags or repositories if needed.

```yaml
image:
  backend:
    tag: "0.0.5"
```

### 2. Docker Images

The official images are available at:

- **Backend**: `dkonsole/dkonsole-backend`
- **Frontend**: `dkonsole/dkonsole-frontend`

## 💰 Support the Project

If you find this project useful, consider donating to support development.

**BSC (Binance Smart Chain) Wallet:**
`0x9baf648fa316030e12b15cbc85278fdbd82a7d20`

## 📧 Contact

For questions or feedback, please contact: **flaucha@gmail.com**

## 🛠️ Development

To run locally:

```bash
# Backend
cd backend && go run main.go

# Frontend
cd frontend && npm run dev
```

## License

MIT License
