# ADR 0003: Uso de React Query para Gestión de Estado del Servidor

**Estado**: Aceptado  
**Fecha**: 2024-11-24  
**Decisor**: Equipo de desarrollo

## Contexto

El frontend de DKonsole necesita gestionar datos del servidor (recursos Kubernetes, métricas, logs) con características como:

- Caching automático
- Refetching en background
- Sincronización de estado
- Manejo de errores y loading states
- Optimistic updates

Las opciones consideradas fueron:

1. Redux con RTK Query
2. React Query (TanStack Query)
3. SWR
4. Estado local con hooks personalizados

## Decisión

Adoptamos React Query (TanStack Query) v5 para la gestión de estado del servidor.

## Consecuencias

### Positivas

- **Caching inteligente**: Cache automático con invalidación configurable
- **Refetching automático**: Revalidación en background cuando la ventana recupera foco
- **Deduplicación**: Múltiples componentes pueden usar la misma query sin requests duplicados
- **Optimistic updates**: Soporte nativo para actualizaciones optimistas
- **DevTools**: Herramientas de desarrollo para debugging
- **TypeScript**: Excelente soporte para TypeScript
- **Menos código**: Reduce significativamente el código boilerplate comparado con Redux

### Negativas

- **Dependencia externa**: Agrega una dependencia al proyecto
- **Curva de aprendizaje**: Requiere entender conceptos como queries, mutations, invalidation
- **Bundle size**: Aumenta ligeramente el tamaño del bundle (mitigado con code splitting)

## Implementación

- React Query se usa para todas las operaciones de lectura (GET)
- Mutations se usan para operaciones de escritura (POST, PUT, DELETE)
- Queries se invalidan después de mutations exitosas
- Configuración global en `App.jsx` con `QueryClient`

## Alternativas Consideradas

- **Redux + RTK Query**: Rechazado por ser más verboso y requerir más configuración
- **SWR**: Considerado pero React Query ofrece más características (mutations, devtools)
- **Estado local**: Rechazado por requerir implementar manualmente caching, refetching, etc.

