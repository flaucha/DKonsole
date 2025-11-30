# Command Framework

This document defines the "Language" for interacting with the AI agent.
The AI must recognize these patterns and execute the corresponding workflows.

## 1. "Analyze Project" / "Analyze [File/Component]"

**Trigger**: "Analyze the repo", "Check this file", "Do a security audit".

**Workflow**:
1.  Read `docs/ANALYSIS_GUIDELINES.md`.
2.  Perform the analysis (Code Quality, Security, SWOT).
3.  Calculate the Score.
4.  Generate a report (Markdown).

## 2. "Add Feature: [Name]"

**Trigger**: "I need a new button that does X", "Add support for Y".

**Workflow**:
1.  **Plan**:
    - Read `docs/CODING_GUIDELINES.md`.
    - Create `implementation_plan.md`.
    - Identify necessary changes in Backend and Frontend.
2.  **Execute**:
    - Implement changes (TDD preferred).
    - Update `CHANGELOG.md`.
3.  **Verify**:
    - Run tests.
    - Verify against `SECURITY_GUIDELINES.md`.

## 3. "Refactor: [Target]"

**Trigger**: "Clean up this code", "Refactor the module".

**Workflow**:
1.  **Analyze**: Identify the issue.
2.  **Plan**: Propose the refactoring in `implementation_plan.md`.
3.  **Execute**: Apply changes.
4.  **Verify**: Ensure tests pass.

## 4. "Release Version [X.Y.Z]"

**Trigger**: "Release version 1.2.0", "Prepare a release".

**Workflow**:
1.  **Strictly** follow `docs/RELEASE.md`.
2.  Update version files and changelogs.
3.  Commit and Tag.

## 5. "Fix Bug: [Description]"

**Trigger**: "Fix the error", "Something is broken".

**Workflow**:
1.  **Reproduce**: Create a test case that fails.
2.  **Analyze**: Find the root cause.
3.  **Fix**: Apply the fix.
4.  **Verify**: Ensure the test passes.

---
**Note to AI**: When you receive a command, explicitly state which workflow you are triggering.
