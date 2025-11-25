# Análisis de Código: Fat Handler y SRP en DKonsole

## Introducción
Este documento detalla el análisis del código fuente de "DKonsole" con el objetivo de identificar "Fat Handlers" y aplicar el Principio de Responsabilidad Única (SRP). Se ha realizado una búsqueda exhaustiva en el backend.

## Hallazgo Principal: Fat Handler en `UploadLogo`
Se ha identificado que la función `UploadLogo` en el archivo `backend/handlers.go` es el caso más claro y crítico de un "Fat Handler".

### Problemas Identificados
La función `UploadLogo` tiene múltiples responsabilidades mezcladas:
1.  **Manejo HTTP**: Parsea el formulario multipart y escribe respuestas HTTP.
2.  **Validación**: Verifica el tamaño del archivo, la extensión y el tipo de contenido (MIME).
3.  **Seguridad**: Realiza una inspección profunda de archivos SVG para prevenir XSS.
4.  **Lógica de Negocio/Persistencia**: Crea directorios, elimina archivos antiguos y guarda el nuevo archivo en el disco.

### Violación de SRP
El Principio de Responsabilidad Única establece que una clase o módulo debe tener una sola razón para cambiar. Actualmente, `UploadLogo` cambiaría si:
-   Cambia la lógica de validación de archivos.
-   Cambia la forma en que se almacenan los archivos (ej. S3 en lugar de disco local).
-   Cambian los requisitos de seguridad para SVG.
-   Cambia la estructura de la petición HTTP.

## Hallazgos Secundarios: Instanciación de Servicios en Handlers
Aunque no son "Fat Handlers" monolíticos (el código está dividido en funciones), se observó un patrón repetitivo en `backend/internal/k8s/resource_operations.go` y `backend/internal/helm/helm_operations.go` que viola parcialmente SRP y la Inversión de Dependencias.

### El Patrón Observado
Los handlers HTTP (ej. `UpdateResourceYAML`, `UpgradeHelmRelease`) son responsables de **construir** todo el grafo de dependencias (Repositorios, Servicios) dentro de la propia función del handler.

```go
// Ejemplo del patrón actual
func (s *Service) UpdateResourceYAML(w http.ResponseWriter, r *http.Request) {
    // ... parsing ...
    resourceRepo := NewK8sResourceRepository(dynamicClient) // El handler sabe qué repo usar
    resourceService := NewResourceService(resourceRepo, ...) // El handler sabe cómo crear el servicio
    // ... llamada al servicio ...
}
```

### Por qué es un problema (menor)
-   **Acoplamiento**: El handler está fuertemente acoplado a las implementaciones concretas de los servicios y repositorios.
-   **Testabilidad**: Es difícil hacer mocks de `resourceService` para probar solo la capa HTTP, porque se crea *dentro* de la función.

*Nota: Para este ejercicio, nos centraremos en refactorizar `UploadLogo` como el objetivo principal, ya que representa una violación más directa de la lógica de negocio mezclada con lógica HTTP.*

## Propuesta de Refactorización (UploadLogo)
Para aplicar SRP, se propone separar la lógica en capas:
1.  **Handler (Controlador)**: Solo debe encargarse de recibir la petición, llamar al servicio y devolver la respuesta.
2.  **Servicio**: Debe contener la lógica de negocio (validación, procesamiento).
3.  **Repositorio/Storage**: Debe encargarse de la persistencia física del archivo.

### Beneficios
-   **Testabilidad**: Se puede probar la lógica de validación y almacenamiento sin necesidad de un servidor HTTP.
-   **Mantenibilidad**: El código es más fácil de leer y modificar.
-   **Reusabilidad**: La lógica de carga de logos podría ser reutilizada en otros contextos.
