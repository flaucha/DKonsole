# AI Framework Template

This directory contains the templates for the "AI Development Framework".
This framework allows an AI agent to autonomously manage a project by following strict guidelines and workflows.

## How to Use

1.  **Copy Files**: Copy the contents of `ai_framework/` to the root (or `docs/`) of your new project.
    - `IA_GUIDELINES.md` -> Project Root
    - `docs/SECURITY_GUIDELINES.md` -> `docs/`
    - `docs/TESTING_GUIDELINES.md` -> `docs/`
    - `docs/ANALYSIS_GUIDELINES.md` -> `docs/`
    - `docs/COMMANDS.md` -> `docs/`

2.  **Customize**:
    - Open `IA_GUIDELINES.md` and replace `[Project Name]` with your project name.
    - Update `docs/CODING_GUIDELINES.md` (you need to create this based on your project's language and style).
    - Update `docs/RELEASE.md` (create this based on your release process).

3.  **Instruct the AI**:
    - Tell the AI: "Read IA_GUIDELINES.md and start working."

## Components

- **IA_GUIDELINES.md**: The entry point.
- **COMMANDS.md**: The "language" you use to talk to the AI.
- **ANALYSIS_GUIDELINES.md**: How the AI should audit the code.
- **SECURITY_GUIDELINES.md**: Security standards (OWASP).
- **TESTING_GUIDELINES.md**: Testing strategy.
