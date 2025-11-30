# Command Framework (Slash Commands)

This document defines the efficient "Slash Command" system for interacting with the AI.
Use these commands for rapid and precise interaction.

## 1. Core Commands

### `/audit`
**Description**: Performs a comprehensive analysis of the project or a specific component.
**Usage**:
- `/audit`: Analyze the entire repository.
- `/audit [file/path]`: Analyze a specific file or directory.
**Workflow**: Follows `docs/ANALYSIS_GUIDELINES.md`.

### `/feat [name]`
**Description**: Initiates the workflow to add a new feature.
**Usage**:
- `/feat Dark Mode`: Plans and implements "Dark Mode".
**Workflow**:
1.  Plan (Create `implementation_plan.md`).
2.  Execute (Backend + Frontend).
3.  Verify (Tests).

### `/fix [description]`
**Description**: Investigates and fixes a bug.
**Usage**:
- `/fix Login 500 error`: Fixes the login issue.
**Workflow**: Reproduce -> Analyze -> Fix -> Verify.

### `/refactor [target]`
**Description**: Refactors code to improve quality without changing behavior.
**Usage**:
- `/refactor auth service`: Refactors the auth module.
**Workflow**: Analyze -> Plan -> Refactor -> Verify.

### `/release [version]`
**Description**: Automates the release process.
**Usage**:
- `/release 1.2.0`: Releases version 1.2.0.
**Workflow**: Strictly follows `docs/RELEASE.md`.

### `/check`
**Description**: Runs the project sanity check.
**Usage**:
- `/check`: Runs `./scripts/ai-check.sh` (or equivalent).

---

## 2. Legacy Commands (Natural Language)
The AI still understands natural language, but maps it to the above commands:
- "Analyze the project" -> `/audit`
- "Add a new button" -> `/feat`
- "Fix the bug" -> `/fix`
