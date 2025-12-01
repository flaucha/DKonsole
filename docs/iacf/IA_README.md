# AI Interaction Guide

Guía rápida para colaborar con la IA usando el IACF.

## 🚀 Inicio rápido
Siempre indica a la IA que lea las guías antes de actuar:

> **"Lee IA_GUIDELINES.md y atiende esta petición: [tu solicitud en lenguaje natural]"**

Así detecta el stack (fullstack/backend/frontend) y carga el contexto correcto.

## ⚡ Cómo pedir (lenguaje natural)
- Análisis: “Analiza el proyecto completo y entrega un reporte.” / “Analiza `ruta/archivo` y dame hallazgos.”
- Desarrollo: “Agrega la funcionalidad X en backend/frontend.” / “Refactoriza <módulo/archivo> para mejorar legibilidad.”
- Bugs: “Corrige el bug Y” o “Arregla el punto N del análisis Z.”
- Verificación/Release: “Ejecuta la verificación del proyecto (lint/tests).” / “Prepara la release <versión>” (si el proyecto tiene flujo).

## 🛠️ Verificación
Pide en cualquier momento:

> **"Ejecuta la verificación del proyecto"**

La IA seguirá `TESTING_GUIDELINES.md` según el stack detectado.

## 📂 Estructura de guías
- `IA_GUIDELINES.md`: cerebro y flujo base.
- `CODING_GUIDELINES.md`: principios y perfiles de ejemplo.
- `SECURITY_GUIDELINES.md`: controles y OWASP agnóstico.
- `TESTING_GUIDELINES.md`: estrategia y cobertura mínima.
- `ANALYSIS_GUIDELINES.md`: formato de análisis y reportes.
