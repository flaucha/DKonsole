# ADR 0001: Arquitectura Modular con Capas Separadas

**Estado**: Aceptado  
**Fecha**: 2024-11-24  
**Decisor**: Equipo de desarrollo

## Contexto

DKonsole necesita manejar operaciones complejas de Kubernetes, autenticación, gestión de Helm releases, y streaming de logs. Inicialmente, el código estaba organizado de manera monolítica con handlers que mezclaban lógica de negocio, acceso a datos y presentación HTTP.

## Decisión

Adoptamos una arquitectura modular con capas claramente separadas:

1. **Handler (HTTP Layer)**: Maneja requests/responses HTTP, validación de entrada, serialización JSON
2. **Service (Business Logic Layer)**: Contiene la lógica de negocio, orquestación, y transformación de datos
3. **Repository (Data Access Layer)**: Abstrae el acceso a datos (Kubernetes API, bases de datos, etc.)

Cada paquete (`auth`, `k8s`, `pod`, `helm`, `prometheus`) sigue esta estructura.

## Consecuencias

### Positivas

- **Separación de responsabilidades**: Cada capa tiene una responsabilidad clara
- **Testabilidad**: Las capas pueden ser testeadas independientemente usando mocks
- **Mantenibilidad**: Cambios en una capa no afectan directamente a las otras
- **Reutilización**: Los servicios pueden ser reutilizados por diferentes handlers
- **Dependency Injection**: Facilita el uso de factories y mocks en tests

### Negativas

- **Más archivos**: La estructura es más verbosa que un enfoque monolítico
- **Curva de aprendizaje**: Los nuevos desarrolladores necesitan entender la arquitectura
- **Overhead inicial**: Requiere más código boilerplate al inicio

## Implementación

Ejemplo de estructura:

```
internal/
  auth/
    auth.go          # HTTP handlers
    service.go       # Business logic
    repository.go    # Data access
  k8s/
    k8s.go          # HTTP handlers
    resource_service.go  # Business logic
    resource_repository.go  # Data access
```

