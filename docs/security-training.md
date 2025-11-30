# Security Training - LDAP Injection Prevention

## ¿Qué es LDAP Injection?

LDAP injection es similar a SQL injection, pero en filtros LDAP.

**Ejemplo vulnerable**:
```go
// ❌ VULNERABLE
filter := fmt.Sprintf("(cn=%s)", username)
// username = "admin)(cn=*" → filter = "(cn=admin)(cn=*)"
// Esto hace bypass de autenticación
```

**Ejemplo seguro**:
```go
// ✅ SEGURO
filter := fmt.Sprintf("(cn=%s)", ldap.EscapeFilter(username))
// username = "admin)(cn=*" → filter = "(cn=admin\\29\\28cn=\\2a)"
// Caracteres especiales escapados
```

## Reglas de Seguridad LDAP

1. ✅ **SIEMPRE** usar `ldap.EscapeFilter()` para inputs de usuario
2. ✅ Validar formato de username ANTES de usar
3. ✅ Loggear intentos de injection
4. ✅ NO confiar en input del frontend
5. ✅ Usar DN completos cuando sea posible

## Checklist de Code Review

- [ ] ¿Se usa `fmt.Sprintf` con input de usuario en filtros LDAP?
- [ ] ¿Se llama `ldap.EscapeFilter()` antes de usar el input?
- [ ] ¿Hay validación de formato de username?
- [ ] ¿Hay tests unitarios para injection payloads?
- [ ] ¿Se loggea el filtro LDAP sanitizado?
