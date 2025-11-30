# AI Interaction Guide

This project is managed with the help of an AI Agent. This guide explains how to effectively collaborate with the AI.

## 🚀 Getting Started

To start any task, always instruct the AI to read the guidelines first:

> **"Read IA_GUIDELINES.md and [Your Command]"**

This ensures the AI loads the correct context, architecture, and constraints.

## 🗣️ Common Commands

Here are the standard commands you can use (defined in `docs/COMMANDS.md`):

### 1. Analysis & Audits
- **"Analyze the repo"**: Performs a full code quality and security audit.
- **"Check this file"**: Analyzes a specific file for issues.
- **"Do a security audit"**: Checks for OWASP vulnerabilities.

### 2. Development
- **"Add Feature: [Name]"**: Starts the workflow to add a new feature (Backend + Frontend).
  - *Example*: "Add Feature: Dark Mode toggle"
- **"Refactor: [Target]"**: Cleans up code or improves structure.
  - *Example*: "Refactor: The auth service to be cleaner"
- **"Fix Bug: [Description]"**: Investigates and fixes a bug.
  - *Example*: "Fix Bug: Login fails with 500 error"

### 3. Release
- **"Release version X.Y.Z"**: Automates the release process (Version bump, Changelog, Tags).
  - *Example*: "Release version 1.2.0"

## 🛠️ Verification

You can ask the AI to verify the project state at any time:

> **"Run the sanity check"**

The AI will execute `./scripts/ai-check.sh` to run linting and tests.

## 📂 Documentation Structure

- **`IA_GUIDELINES.md`**: The brain. The AI reads this to know what to do.
- **`docs/CODING_GUIDELINES.md`**: Rules for Go and React code.
- **`docs/SECURITY_GUIDELINES.md`**: Security standards.
- **`docs/COMMANDS.md`**: Detailed workflows for the commands above.
