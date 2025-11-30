# AI Interaction Guide

This project is managed with the help of an AI Agent. This guide explains how to effectively collaborate with the AI.

## 🚀 Getting Started

To start any task, always instruct the AI to read the guidelines first:

> **"Read IA_GUIDELINES.md and [Your Command]"**

This ensures the AI loads the correct context, architecture, and constraints.

## ⚡ Fast Track (Slash Commands)

Use these commands for rapid interaction (defined in `docs/COMMANDS.md`):

### 1. Analysis & Audits
- **`/audit`**: Full project analysis.
- **`/audit [file]`**: Analyze a specific file.

### 2. Development
- **`/feat [Name]`**: Add a new feature.
  - *Example*: `/feat Dark Mode`
- **`/refactor [Target]`**: Clean up code.
  - *Example*: `/refactor auth service`
- **`/fix [Bug]`**: Fix a bug.
  - *Example*: `/fix Login 500 error`

### 3. Release & Verify
- **`/release [Version]`**: Trigger release workflow.
  - *Example*: `/release 1.2.0`
- **`/check`**: Run sanity checks (Lint + Test).

## 🛠️ Verification

You can ask the AI to verify the project state at any time:

> **"Run the sanity check"**

The AI will execute `./scripts/ai-check.sh` to run linting and tests.

## 📂 Documentation Structure

- **`IA_GUIDELINES.md`**: The brain. The AI reads this to know what to do.
- **`docs/CODING_GUIDELINES.md`**: Rules for Go and React code.
- **`docs/SECURITY_GUIDELINES.md`**: Security standards.
- **`docs/COMMANDS.md`**: Detailed workflows for the commands above.
