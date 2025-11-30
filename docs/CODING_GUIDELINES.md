# DKonsole Coding Guidelines

This document establishes the coding standards and architectural patterns for the DKonsole project. **All code contributions, whether by humans or AI models, must adhere to these guidelines.**

## 1. Core Philosophy

- **AI-First & Clean Code**: The codebase is maintained by AI agents. Code must be clear, self-documenting, and follow strict patterns to minimize hallucinations and context errors.
- **Domain-Driven Design (DDD)**: Logic is organized by domain (e.g., `k8s`, `auth`, `helm`), not by technical layer (e.g., controllers, services).
- **"Shared-Nothing" Architecture**: The backend serves as a unified binary. Dependencies are injected explicitly.

---

## 2. Backend Guidelines (Go)

### 2.1 Directory Structure
All backend logic resides in `backend/internal/`.

- `backend/internal/models/`: Shared types and structs only. **No logic.** Used to prevent circular dependencies.
- `backend/internal/utils/`: Generic helpers (Logger, HTTP responses).
- `backend/internal/<domain>/`: Feature modules (e.g., `k8s`, `auth`).

### 2.2 Layered Architecture Pattern
Each domain module must follow this structure:

1.  **Factory (`factories.go`)**:
    - Defines interfaces for Services and Repositories.
    - Implements a `ServiceFactory` struct to inject dependencies (e.g., K8s clients).
    - **Rule**: Never instantiate services directly in handlers; use the factory.

2.  **Service (`service.go`)**:
    - Contains business logic.
    - Methods receive `context.Context` as the first argument.
    - **Rule**: Services should not deal with HTTP specific objects (`http.ResponseWriter`) directly if possible; return data and errors, let the Handler layer manage HTTP. (Note: Current implementation sometimes mixes this, but moving towards separation is preferred).

3.  **Repository (`repository.go`)**:
    - Handles data access (Kubernetes API calls, Database, external APIs).
    - **Rule**: All Kubernetes client calls go here.

### 2.3 Dependency Injection
- Use constructor functions (`NewService`, `NewRepository`).
- Pass dependencies via interfaces to facilitate testing (mocking).

### 2.4 Error Handling
- **Wrap Errors**: Use `fmt.Errorf("failed to operation: %w", err)` to provide context.
- **Specific Errors**: Define domain-specific errors in the module or `models` if reused.
- **HTTP Errors**: Use `utils.ErrorResponse(w, code, message)` for consistent JSON error responses.

### 2.5 Logging
- **Strict Rule**: NEVER use `fmt.Println` or `log.Println` in production code.
- **Use**: `utils.Logger` (built on `slog`).
- **Format**: Structured JSON logging.
    ```go
    utils.LogInfo("Operation successful", map[string]interface{}{
        "user": username,
        "resource": resourceName,
    })
    ```

---

## 3. Frontend Guidelines (React)

### 3.1 Component Structure
- **Functional Components**: Use React Functional Components with Hooks.
- **Location**:
    - `src/components/`: Shared/atomic components.
    - `src/components/details/`: Resource-specific detail views.
    - `src/pages/`: Top-level page views.

### 3.2 State Management
- Use **React Context** for global state (e.g., `AuthContext`, `ClusterContext`).
- Use local `useState` / `useReducer` for component-specific state.

### 3.3 Styling
- Use **Tailwind CSS** utility classes.
- Avoid inline styles unless dynamic values are required.
- Maintain Dark Mode compatibility (use `dark:` prefix).

### 3.4 API Interaction
- Use the configured `api` client (axios/fetch wrapper) which handles JWT tokens automatically.
- Handle loading and error states explicitly in UI.

---

## 4. Git & Workflow

### 4.1 Commit Messages
Follow **Conventional Commits** spec:
- `feat: add new feature`
- `fix: resolve bug`
- `docs: update documentation`
- `chore: maintenance tasks`
- `refactor: code restructuring`

### 4.2 Versioning
- Follow **Semantic Versioning** (MAJOR.MINOR.PATCH).
- Version is tracked in the `VERSION` file.
- Release process is automated via `RELEASE.md`.

---

## 5. New Feature Checklist

When adding a new feature (e.g., "NetworkPolicies"):

1.  [ ] Define shared types in `backend/internal/models/`.
2.  [ ] Create directory `backend/internal/networkpolicy/`.
3.  [ ] Implement `repository.go` (K8s calls).
4.  [ ] Implement `service.go` (Business logic).
5.  [ ] Implement `factories.go` (Dependency injection).
6.  [ ] Register handlers in `backend/main.go` using the factory.
7.  [ ] Create frontend component in `frontend/src/components/details/`.
8.  [ ] Add tests.
9.  [ ] Update `CHANGELOG.md`.
