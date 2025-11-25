# ADR 0002: Autenticación Basada en JWT

**Estado**: Aceptado
**Fecha**: 2024-11-24
**Decisor**: Equipo de desarrollo

## Contexto

DKonsole necesita un sistema de autenticación seguro para proteger el acceso a recursos de Kubernetes. Las opciones consideradas fueron:

1. Autenticación basada en sesiones (cookies)
2. Autenticación basada en JWT (JSON Web Tokens)
3. Autenticación basada en tokens opacos

## Decisión

Implementamos autenticación basada en JWT con las siguientes características:

- **Algoritmo**: HS256 (HMAC-SHA256)
- **Expiración**: 24 horas
- **Almacenamiento**: HTTP-only cookie + opcional en header Authorization
- **Hash de contraseñas**: Argon2id
- **Secreto JWT**: Configurado via variable de entorno `JWT_SECRET`

## Consecuencias

### Positivas

- **Stateless**: No requiere almacenamiento de sesiones en el servidor
- **Escalabilidad**: Funciona bien en arquitecturas distribuidas
- **Seguridad**: Argon2id es resistente a ataques de fuerza bruta
- **Flexibilidad**: El token puede ser enviado en header o cookie
- **Estándar**: JWT es un estándar ampliamente adoptado

### Negativas

- **Revocación limitada**: No se puede revocar un token antes de su expiración sin mantener una blacklist
- **Tamaño**: Los tokens JWT pueden ser más grandes que tokens opacos
- **Exposición**: Si el token es comprometido, es válido hasta su expiración

## Implementación

- El token se genera en `/api/login` después de validar credenciales
- Se almacena en cookie HTTP-only para protección XSS
- Se valida en cada request mediante `AuthMiddleware`
- Los claims incluyen `username` y `role`

## Alternativas Consideradas

- **Sesiones**: Rechazado por requerir almacenamiento de estado en el servidor
- **OAuth2**: Considerado para futuras versiones para integración con proveedores externos
