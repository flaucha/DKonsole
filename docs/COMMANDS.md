# Command Framework (Slash Commands)

This document defines the efficient "Slash Command" system for interacting with the AI.
Use these commands for rapid and precise interaction.

## 1. Core Commands

### `/audit`
**Description**: Performs a comprehensive analysis and **generates a report in the root directory**.
**Usage**:
- `/audit`: Analyze the entire repository. Generates `AUDIT_REPORT.md` (or similar) in the root.
- `/audit [file/path]`: Analyze a specific file.
**Workflow**:
1.  Follow `docs/ANALYSIS_GUIDELINES.md`.
2.  Save the report as `[NAME]_ANALYSIS.md` in the project root.

### `/feat [name]`
**Description**: Initiates the workflow to add a new feature.
**Usage**:
- `/feat Dark Mode`: Plans and implements "Dark Mode".
**Workflow**:
1.  Plan (Create `implementation_plan.md`).
2.  Execute (Backend + Frontend).
3.  Verify (Tests).

### `/fix [target]`
**Description**: Investigates and fixes a bug or an audit point.
**Usage**:
- `/fix point [X] [analysis_file.md]`: Fixes specific point X from an analysis file.
- `/fix [description]`: Fixes a bug described by the user.
**Workflow**:
1.  **Analyze**: Understand the issue (or read the analysis point).
2.  **Plan**: Create/Update `implementation_plan.md`.
3.  **Execute**: Apply the fix following `docs/CODING_GUIDELINES.md`.
4.  **Verify**: Ensure tests pass.

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
- `/check`: Runs `./scripts/ai-check.sh`.

---

## 2. Legacy Commands (Natural Language)
The AI still understands natural language, but maps it to the above commands:
- "Analyze the project" -> `/audit`
- "Add a new button" -> `/feat`
- "Fix the bug" -> `/fix`
