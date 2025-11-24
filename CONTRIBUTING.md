# Guía de Contribución a DKonsole

¡Gracias por tu interés en contribuir a DKonsole! Esta guía te ayudará a entender cómo puedes participar en el proyecto.

## 📋 Tabla de Contenidos

- [Código de Conducta](#código-de-conducta)
- [Configuración del Entorno de Desarrollo](#configuración-del-entorno-de-desarrollo)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Estándares de Código](#estándares-de-código)
- [Proceso de Contribución](#proceso-de-contribución)
- [Testing](#testing)
- [Arquitectura del Proyecto](#arquitectura-del-proyecto)

## 🤝 Código de Conducta

Al participar en este proyecto, te comprometes a mantener un ambiente respetuoso y acogedor para todos los colaboradores.

## 🛠️ Configuración del Entorno de Desarrollo

### Prerrequisitos

- **Go 1.24+**: Para el backend
- **Node.js 20+**: Para el frontend
- **Docker** (opcional): Para pruebas con contenedores
- **kubectl** (opcional): Para pruebas con Kubernetes

### Configuración Inicial

1. **Fork y clona el repositorio:**
   ```bash
   git clone https://github.com/tu-usuario/DKonsole.git
   cd DKonsole
   ```

2. **Configura el backend:**
   ```bash
   cd backend
   go mod download
   ```

3. **Configura el frontend:**
   ```bash
   cd frontend
   npm install
   ```

4. **Variables de entorno (backend):**
   ```bash
   export JWT_SECRET="tu-secret-key-de-al-menos-32-caracteres"
   export GO_ENV="development"
   ```

## 📁 Estructura del Proyecto

```
DKonsole/
├── backend/              # Backend en Go
│   ├── internal/        # Módulos internos (arquitectura orientada al dominio)
│   │   ├── models/      # Tipos compartidos
│   │   ├── api/         # Handlers de API genéricos
│   │   ├── k8s/         # Handlers de recursos Kubernetes
│   │   ├── helm/        # Handlers de Helm
│   │   ├── auth/        # Handlers de autenticación
│   │   ├── cluster/     # Gestión de clusters
│   │   ├── pod/         # Operaciones de pods
│   │   └── utils/       # Utilidades compartidas
│   └── main.go          # Punto de entrada
├── frontend/            # Frontend en React
│   └── src/
│       └── components/  # Componentes React
├── scripts/             # Scripts de utilidad
├── helm/                # Charts de Helm
└── .github/             # Configuración de GitHub Actions
```

## 📝 Estándares de Código

### Backend (Go)

- **Formato**: Usa `gofmt` o `goimports` para formatear el código
- **Linting**: El código debe pasar `go vet ./...`
- **Nombres**: Usa nombres descriptivos y sigue las convenciones de Go
- **Comentarios**: Documenta funciones públicas con comentarios
- **Errores**: Siempre maneja errores explícitamente, nunca los ignores
- **Contextos**: Usa `r.Context()` en handlers HTTP en lugar de `context.TODO()`

**Ejemplo:**
```go
// GetNamespaces obtiene la lista de namespaces del cluster especificado
func (s *Service) GetNamespaces(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // Usar el contexto de la petición
    // ...
}
```

### Frontend (React)

- **Formato**: Usa Prettier (configurado en el proyecto)
- **Linting**: El código debe pasar `npm run lint`
- **Componentes**: Usa componentes funcionales con hooks
- **Nombres**: Usa PascalCase para componentes, camelCase para funciones
- **Props**: Define tipos para las props cuando sea posible

**Ejemplo:**
```jsx
// PodTable.jsx
const PodTable = ({ pods, onSelect }) => {
  // ...
};
```

### Commits

- Usa mensajes de commit descriptivos
- Prefiere commits pequeños y frecuentes
- Formato sugerido:
  ```
  tipo(área): descripción breve
  
  Descripción detallada si es necesario
  ```

  Tipos comunes:
  - `feat`: Nueva funcionalidad
  - `fix`: Corrección de bug
  - `refactor`: Refactorización
  - `docs`: Documentación
  - `test`: Tests
  - `chore`: Tareas de mantenimiento

**Ejemplo:**
```
feat(k8s): agregar paginación a GetResources

Implementa paginación usando limit y continue para evitar
problemas de memoria en clusters grandes.
```

## 🔄 Proceso de Contribución

### 1. Crear una Rama

```bash
git checkout -b feature/mi-nueva-funcionalidad
# o
git checkout -b fix/correccion-de-bug
```

### 2. Hacer Cambios

- Realiza tus cambios siguiendo los estándares de código
- Asegúrate de que los tests pasen localmente
- Actualiza la documentación si es necesario

### 3. Ejecutar Tests

**Backend:**
```bash
cd backend
go vet ./...
go test -v ./...
```

**Frontend:**
```bash
cd frontend
npm run lint
npm run test -- --run
```

### 4. Commit y Push

```bash
git add .
git commit -m "feat(área): descripción"
git push origin feature/mi-nueva-funcionalidad
```

### 5. Crear Pull Request

1. Ve a GitHub y crea un Pull Request
2. Describe claramente los cambios realizados
3. Menciona cualquier issue relacionado
4. Espera la revisión del código

### 6. Revisión de Código

- Responde a los comentarios de los revisores
- Realiza los cambios solicitados
- Mantén la conversación respetuosa y constructiva

## 🧪 Testing

### Backend

Ejecuta los tests antes de hacer commit:

```bash
cd backend
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Ver cobertura
```

**Escribir nuevos tests:**
- Crea archivos `*_test.go` en el mismo paquete
- Usa la convención de nombres de Go
- Prueba casos de éxito y error

**Ejemplo:**
```go
func TestGetNamespaces(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

### Frontend

```bash
cd frontend
npm run test -- --run
npm run test:coverage  # Con cobertura
```

**Escribir nuevos tests:**
- Usa Vitest (ya configurado)
- Prueba componentes y funciones
- Mockea dependencias externas

## 🏗️ Arquitectura del Proyecto

### Backend - Arquitectura Orientada al Dominio

El backend está organizado en módulos dentro de `backend/internal/`:

- **`models/`**: Tipos compartidos y estructuras de datos
- **`api/`**: Handlers para recursos de API genéricos y CRDs
- **`k8s/`**: Handlers para recursos estándar de Kubernetes
- **`helm/`**: Handlers para operaciones de Helm
- **`auth/`**: Lógica de autenticación y autorización
- **`cluster/`**: Gestión de múltiples clusters
- **`pod/`**: Operaciones específicas de pods (logs, exec, events)
- **`utils/`**: Funciones auxiliares compartidas

**Principios:**
- Separación de responsabilidades
- Evitar dependencias circulares (usar `models/` para tipos compartidos)
- Cada módulo puede ser testeado independientemente
- Los handlers HTTP deben usar el contexto de la petición

### Frontend - React

- Componentes funcionales con hooks
- Separación de lógica y presentación
- Reutilización de componentes

## 📚 Recursos Adicionales

- [Guía de Testing](./TESTING.md)
- [Arquitectura del Backend](./backend/internal/README.md)
- [GitHub Actions](./GITHUB_ACTIONS_GUIA.md)

## ❓ Preguntas

Si tienes preguntas o necesitas ayuda:
- Abre un issue en GitHub
- Contacta a los mantenedores: **flaucha@gmail.com**

## 🙏 Agradecimientos

¡Gracias por contribuir a DKonsole! Tu ayuda hace que este proyecto sea mejor para todos.

