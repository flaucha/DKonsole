# IA Configuration Framework (IACF)

Framework liviano para que un agente de IA pueda arrancar, detectar el stack y ejecutar tareas de forma consistente.

## Punto de entrada
- **[IA_GUIDELINES.md](./IA_GUIDELINES.md)**: léelo primero; explica cómo detectar alcance/stack y qué guías aplicar.

## Guías incluidas
- [IA_README.md](./IA_README.md): introducción y atajos de uso.
- [CODING_GUIDELINES.md](./CODING_GUIDELINES.md): principios y perfiles de ejemplo.
- [SECURITY_GUIDELINES.md](./SECURITY_GUIDELINES.md): controles mínimos y OWASP agnóstico.
- [TESTING_GUIDELINES.md](./TESTING_GUIDELINES.md): estrategia de pruebas.
- [ANALYSIS_GUIDELINES.md](./ANALYSIS_GUIDELINES.md): cómo auditar y puntuar.

## Cómo adoptarlo en tu repo
1. Copia estos archivos (raíz o `docs/`; mantén los enlaces coherentes).
2. Ajusta `CODING_GUIDELINES.md` con los perfiles de stack que uses.
3. Ajusta controles de `SECURITY_GUIDELINES.md` y criterios de pruebas en `TESTING_GUIDELINES.md`.
4. Indica al agente: “Lee IA_GUIDELINES.md y atiende esta petición: <tu solicitud>”.

## Ejemplo de inicialización
- **Repo nuevo**: copia las guías a la raíz, revisa y adapta `CODING_GUIDELINES.md`/`TESTING_GUIDELINES.md`/`SECURITY_GUIDELINES.md` a tu stack, y añade cualquier script de verificación. Primer mensaje al agente: “Lee IA_GUIDELINES.md y prepara la verificación del proyecto”.
- **Proyecto existente**: coloca las guías en la raíz o `docs/`, actualiza referencias internas si cambian rutas, añade perfiles de stack reales a `CODING_GUIDELINES.md` y comandos de lint/test actuales en `TESTING_GUIDELINES.md`. Primer mensaje al agente: “Lee IA_GUIDELINES.md, detecta el stack del proyecto y resume qué perfiles y scripts usar”.

## Alcance
Puede usarse en repos fullstack o de solo backend/frontend; el agente detecta el alcance con las reglas de `IA_GUIDELINES.md` y aplica el perfil correspondiente. Originado en DKonsole y generalizado para uso amplio.
