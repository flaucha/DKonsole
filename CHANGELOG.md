# Changelog

All notable changes to DKonsole will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.10] - 2025-01-27

### Security

- **Dependency Updates**: Updated dependencies to address security vulnerabilities
  - Updated Kubernetes client libraries from v0.29.0 to v0.34.2 (k8s.io/api, k8s.io/apimachinery, k8s.io/client-go, k8s.io/metrics)
  - Updated JWT library from v5.2.1 to v5.3.0 (github.com/golang-jwt/jwt/v5)
  - Updated WebSocket library to latest version (github.com/gorilla/websocket)
  - Updated multiple transitive dependencies to latest secure versions
  - These updates address known security vulnerabilities in dependencies

## [1.1.9] - 2025-01-27

### Security

- **Fixed Critical RCE Vulnerability**: Fixed unauthenticated pod execution endpoint
  - The `/api/pods/exec` endpoint was not protected with authentication middleware
  - Now properly wrapped with `secure()` middleware requiring authentication
  - Prevents unauthenticated attackers from executing arbitrary commands in pods
  - This was a critical security vulnerability that allowed Remote Code Execution (RCE)

### Fixed

- **Resource Loading Issue**: Fixed problem where resources wouldn't load when switching between sections
  - Added `currentCluster` to React Query `queryKey` to properly invalidate cache
  - Improved state management when switching between resource types (Pods, ConfigMaps, etc.)
  - Added forced refetch when `kind`, `namespace`, or `cluster` changes
  - Configured React Query with `staleTime: 0` and `refetchOnMount: true` for fresh data
  - Resources now load correctly without requiring manual page refresh

## [1.1.8] - 2025-11-24

### Added

- **Selector de color para logs con persistencia**: Nueva funcionalidad para personalizar el color del texto en los logs
  - Selector visual mejorado con cuadraditos de color seleccionables
  - Opciones disponibles: gris, verde, celeste, amarillo, naranja, blanco
  - Persistencia de selección usando localStorage (la selección se mantiene entre sesiones)
  - Diseño más armónico con fondo oscuro, mejor espaciado y efectos visuales mejorados
  - El color seleccionado se aplica a todos los logs en tiempo real
  - Disponible en todos los visores de logs (LogViewerInline)

## [1.1.7] - 2025-01-26

### Fixed

- **Settings UI Improvements**: Moved About section to Settings tab
  - About is now integrated as a tab within Settings instead of a separate page
  - Added General tab with Languages placeholder ("Coming soon... Languages")
  - Improved Settings navigation consistency

## [1.1.6] - 2025-01-26

### 🎯 In-Place Editing & Performance Improvements

This release introduces in-place editing for Secrets and ConfigMaps, removes pagination limitations, and adds an About section.

### Added

- **In-Place Editing for Secrets and ConfigMaps**: Direct editing of data fields without YAML
  - "Edit in place" button for Secrets and ConfigMaps
  - Unmask secret values for editing
  - Direct field editing and saving
  - Automatic base64 encoding/decoding for secrets
- **About Section**: New About page with version information
  - Version display
  - GitHub, email, and BuyMeACoffee links
  - Features list
  - Accessible from Settings navigation

### Changed

- **Removed Pagination Limitations**: All resources now load completely without pagination
  - Removed pagination limits from backend API
  - Removed items per page configuration from Settings
  - Removed display limit selector from resource lists
  - All resources load automatically without manual pagination
- **Improved Resource Loading**: Simplified resource fetching
  - Direct array responses from API (no paginated structure)
  - Automatic loading of all resources
  - Better performance for large clusters

### Fixed

- **Namespace "all" Support**: Fixed issue where ConfigMaps, Secrets, and Deployments returned empty when selecting "all" namespaces
- **Dependency Updates**: Updated npm dependencies to resolve memory leak warnings
  - Updated `inflight` and `glob` dependencies
  - Fixed deprecated package warnings

### Technical

- **Backend Changes**:
  - Removed pagination parameters (`limit`, `continue`) from resource listing
  - Simplified `GetResources` endpoint to return direct array
  - Removed `PaginatedResources` structure usage
- **Frontend Changes**:
  - Removed pagination UI components
  - Removed `itemsPerPage` from SettingsContext
  - Simplified `useWorkloads` hook
  - Added `DataEditor` component for in-place editing
  - Added `About` component with automatic version reading

## [1.1.5] - 2025-01-25

### 🧪 Testing Infrastructure & CI/CD

This release introduces comprehensive testing infrastructure, unit tests, and automated CI/CD pipeline.

### Added

- **Testing Framework Setup**: Complete testing infrastructure for both frontend and backend
  - Vitest configured with React Testing Library for frontend
  - Go testing framework configured for backend
  - Test setup files and utilities created
  - Test configuration in `vite.config.js` and Go test files
- **Unit Tests**: Comprehensive test suite added
  - **Frontend Tests** (23 tests across 5 files):
    - `dateUtils.test.js`: Tests for date formatting utilities
    - `resourceParser.test.js`: Tests for resource parsing functions
    - `statusBadge.test.js`: Tests for status badge utilities
    - `expandableRow.test.js`: Tests for expandable row styling
    - `k8sApi.test.js`: Tests for Kubernetes API client
  - **Backend Tests**:
    - `utils_test.go`: Tests for utility functions
    - `models_test.go`: Tests for shared models and types
- **CI/CD Pipeline**: GitHub Actions workflow configured
  - Automated testing on push to `main` branch
  - Automated testing on Pull Requests to `main`
  - Coverage reports generated for both frontend and backend
  - Build verification step to ensure code compiles
  - Workflow excludes documentation (`.md`) and script changes to reduce unnecessary runs
- **Testing Scripts**: Automation scripts for easy testing
  - `test-all.sh`: Run all tests with a single command
  - `scripts/test-frontend.sh`: Frontend testing script with watch mode support
  - `scripts/test-backend.sh`: Backend testing script with verbose output
  - `scripts/test-backend-docker.sh`: Docker-based backend testing alternative
  - `scripts/update-go.sh`: Script to update Go version
  - `scripts/install-go.sh`: Script to install Go
- **Documentation**: Comprehensive testing and CI/CD documentation
  - `TESTING.md`: Complete testing guide with examples
  - `COMO_PROBAR.md`: Quick testing guide in Spanish
  - `GITHUB_ACTIONS_GUIA.md`: GitHub Actions guide
  - `COMO_VER_RESULTADOS_GITHUB.md`: Visual guide for viewing CI results
  - `scripts/README_TESTS.md`: Scripts documentation
  - `ACTUALIZAR_GO_AHORA.md`: Go update instructions
  - `INSTRUCCIONES_ACTUALIZAR_GO.md`: Detailed Go update instructions

### Changed

- **Go Version**: Backend updated to Go 1.25.4
- **CI Optimization**: GitHub Actions workflow now ignores changes to `.md` files and `scripts/` directory
- **Project Structure**: Added test directories and configuration files

### Technical Improvements

- **Frontend Testing**:
  - Vitest configured with React Testing Library
  - Test setup file with proper mocks and utilities
  - Coverage reporting configured
- **Backend Testing**:
  - Go test framework properly configured
  - Test utilities for common testing scenarios
  - Coverage reporting with `go tool cover`
- **CI/CD**:
  - Parallel job execution for faster feedback
  - Caching of dependencies (Go modules, npm packages)
  - Conditional job execution based on file changes
  - Coverage upload to codecov

## [1.1.4] - 2025-01-24

### 🎨 UI Refactor & Bug Fixes

This release includes a major frontend refactor with improved consistency, bug fixes, and enhanced user experience.

### Added

- **Consistent List Style**: Unified list styling across all resource managers
  - WorkloadList, NamespaceManager, and HelmChartManager now share the same modern design
  - Consistent toolbar with search, item count, and refresh button
  - Unified table headers with sorting indicators
  - Expandable rows with smooth animations and consistent styling
- **Resource Delete Menu**: Restored three-dot menu for resource management
  - Delete and Force Delete options for all resources
  - Confirmation modal with clear warnings for force delete
  - Consistent menu placement and styling across all resource types

### Changed

- **Frontend API Integration**: Updated to use correct backend endpoint
  - Changed from `/api/workloads/${kind}` to `/api/resources?kind=${kind}`
  - Improved error handling and validation
  - Better error messages for debugging
- **Expanded Details Styling**: Enhanced visual design for resource details
  - Improved background colors and borders for better contrast
  - Better spacing and padding in detail panels
  - Scrollable content when details exceed viewport height
  - Consistent styling across all resource types (Workloads, Namespaces, Helm Charts)
- **Log Viewer**: Improved scroll behavior
  - Smooth scrolling when opening logs tab
  - Consistent behavior with terminal tab
  - Better viewport management

### Fixed

- **API Endpoint Issues**: Fixed 404 errors when viewing resource lists
  - Added validation to prevent API calls with undefined `kind` parameter
  - Fixed endpoint mismatch between frontend and backend
  - Improved error handling for missing resource types
- **Empty State Display**: Fixed blank screen when no resources exist
  - Shows informative message with icon when no resources found
  - Displays refresh button for manual reload
  - Handles null/undefined resource arrays gracefully
- **Backend Error Handling**: Improved error propagation for Jobs and CronJobs
  - Errors are now properly returned instead of silently ignored
  - Better error messages for debugging
- **Edit YAML Button**: Fixed non-functional Edit YAML button in WorkloadList
  - Restored YamlEditor modal integration
  - Automatic data refresh after saving changes
- **UI Alignment**: Fixed tab positioning in Pod details
  - Added padding to tabs container for better visual spacing
  - Improved alignment with detail panel content

### Technical Improvements

- **Frontend Refactoring**:
  - Centralized resource fetching logic
  - Improved state management for expanded rows
  - Better error boundary handling
  - Consistent component structure across resource managers
- **Backend Error Handling**:
  - Proper error propagation for Batch API resources (Jobs, CronJobs)
  - Better logging for debugging
  - Improved error messages in API responses

## [1.1.3] - 2025-01-23

### 🚀 Helm Charts Manager & UI Enhancements

This release introduces a complete Helm Charts Manager and significant UI improvements across the application.

### Added

- **Helm Charts Manager**: Complete interface for managing Helm releases
  - View all installed Helm releases across all namespaces
  - Install new Helm charts with automatic repository detection
  - Upgrade existing releases with Monaco YAML editor for values
  - Uninstall releases with confirmation modal
  - Expandable rows showing detailed release information (Chart, Version, App Version, Revision, Status, Description, Last Updated)
  - Automatic repository URL and version detection from existing releases
  - Dynamic repository discovery using `helm search repo` when needed
  - Safe hint suggesting to specify repo URL and version for best results
- **Pod Events & Timeline Tab**: New tab in pod details view
  - Kubernetes events with type (Warning/Normal), reason, message, source, count, and timestamps
  - Container status timeline showing all state transitions (Running, Waiting, Terminated)
  - Detailed restart information with exit codes
  - Image information and timestamps for all container states
  - All text in English for consistency
- **UI State Persistence**: Browser-level persistence using localStorage
  - Remembers selected namespace across page refreshes
  - Remembers current view/page across refreshes
  - Improves user experience by maintaining context
- **Responsive Tables**: Horizontal scroll when window is narrow
  - All tables now support horizontal scrolling when content exceeds viewport width
  - Applied to WorkloadList, HelmChartManager, NamespaceManager, ApiExplorer, CRDExplorer
- **Pod Table Enhancements**:
  - Added "Ready" column showing ready containers vs total (e.g., "2/2")
  - Added "Restarts" column with numeric count
  - Removed "Kind" column as requested
  - Improved date formatting using standardized format
  - Sorting support for Ready and Restarts columns
- **PVC Table Enhancements**:
  - Added "Size" column displaying capacity or requested storage
  - Sorting support for Size column (parses Gi, Mi, Ki units)
  - Removed usage progress bar from details view
- **Date Formatting Standardization**:
  - Created `dateUtils.js` utility with consistent date formatting functions
  - `formatDateTime`, `formatDateTimeShort`, `formatDate`, `formatRelativeTime`, `formatAge`
  - Applied across all components for consistency
- **Expandable Row Standardization**:
  - Created `expandableRow.js` utility with consistent styling functions
  - Smooth animations and consistent appearance across all tables
  - Applied to WorkloadList, HelmChartManager, NamespaceManager

### Changed

- **Namespace Manager**: Fixed line break issues and added smooth slide animations
  - Applied `min-w-0` and `truncate` to prevent line breaks
  - Added smooth transition animations for expandable details
  - Improved table border rendering
- **WorkloadList**: Enhanced pod details with container status information
  - Added `containerStatuses` to pod details API response
  - Includes state, ready status, restart count, timestamps, reasons, exit codes, images
- **Helm Upgrade/Install**: Improved repository detection logic
  - Extracts repository URL from existing Helm release metadata
  - Falls back to dynamic `helm search repo` when needed
  - Adds common Helm repositories automatically for search
  - Better handling of various repository URL patterns
- **PVC Details**: Removed usage progress bar, added Requested and Capacity as detail rows

### Fixed

- **Helm Charts Manager**: Fixed missing "Install Chart" button visibility
- **Namespace Manager**: Fixed line break in table header/first row
- **Date Formats**: Standardized all date displays across the application
- **Expandable Rows**: Consistent styling and animations across all components

### Technical Improvements

- **Backend Helm Support**:
  - New endpoints: `/api/helm/releases` (GET, POST for upgrade, DELETE for uninstall)
  - New endpoint: `/api/helm/releases/install` (POST)
  - Helm release parsing from Kubernetes Secrets with base64 and gzip decompression
  - Kubernetes Job creation for Helm operations using `dkonsole` ServiceAccount
  - ConfigMap creation for values YAML in `dkonsole` namespace
  - Automatic chart name detection from existing releases
  - Robust repository URL extraction and normalization
- **Backend Pod Events**:
  - New endpoint: `/api/pods/events` for fetching Kubernetes events
  - Returns events with type, reason, message, source, count, timestamps
- **Backend PVC Metrics**:
  - Enhanced PVC details with `requested` and `capacity` fields
  - Prometheus integration for PVC usage metrics (if configured)
- **Frontend Utilities**:
  - `dateUtils.js`: Centralized date formatting functions
  - `expandableRow.js`: Centralized expandable row styling functions
- **State Management**: Improved localStorage integration for UI state persistence

## [1.1.2] - 2025-01-23

### ✨ Resource Quota Manager Improvements

This release introduces significant improvements to the Resource Quota Manager interface and functionality.

### Added

- **Namespace Selector**: Added namespace filter selector similar to API Explorer
  - Toggle between "All Namespaces" and selected namespace from header
  - Automatically syncs with namespace selected in the main header
  - Clean, consistent UI without dropdown clutter
- **Automatic Refresh**: Resources now automatically refresh after create, edit, or delete operations
  - No manual refresh button needed
  - Immediate visual feedback on all operations

### Changed

- **UI Consistency**: Reorganized Resource Quota Manager layout to match other pages
  - Removed centered max-width container
  - Standardized header with icon and title size
  - Consistent button styles and spacing
- **Menu Simplification**: Streamlined card menu options
  - Removed redundant "Edit" button
  - Only "Edit YAML" and "Delete" options remain
  - Cleaner, more focused interface
- **YAML Editor**: Fixed YAML editor to use `kubectl apply` (Server-Side Apply)
  - Changed from PUT to `/api/resource/import` endpoint
  - Properly handles both create and update operations
  - Equivalent to `kubectl apply -f` behavior
- **Create New Menu**: Changed from hover to click interaction
  - More predictable user experience
  - Menu closes on outside click
  - Closes automatically after selection
- **Template Namespace**: New resource templates now use the selected namespace
  - ResourceQuota and LimitRange templates automatically use correct namespace
  - Respects namespace filter selection

### Fixed

- **Color Scheme**: Removed colorful elements for consistent gray-scale design
  - Progress bars now use neutral gray colors
  - Removed blue/purple accent colors from cards and tabs
  - Consistent with rest of application aesthetic
- **Delete Refresh**: Fixed issue where deleted resources didn't refresh automatically
  - Modal now closes properly after deletion
  - List refreshes immediately after successful delete

### Technical Improvements

- Improved state management for namespace filtering
- Better synchronization between header namespace and resource filter
- Enhanced error handling in delete operations
- Cleaner code organization

## [1.1.1] - 2024-12-19

### 🐛 Bug Fixes and Improvements

This release fixes critical issues with resource import and YAML viewing functionality.

### Fixed

- **Resource Import**: Fixed import functionality to accept ANY Kubernetes resource (including CRDs and custom resources)
  - Removed restrictive whitelist that limited importable resources
  - Added dynamic GVR resolution using Kubernetes Discovery API
  - Improved support for cluster-scoped resources (ClusterRole, ClusterRoleBinding, etc.)
- **HPA YAML Viewing**: Fixed issue where HorizontalPodAutoscaler YAML could not be viewed
  - Added alias normalization (HPA → HorizontalPodAutoscaler)
  - Implemented automatic fallback between autoscaling/v1 and autoscaling/v2 API versions
  - Added comprehensive error handling and logging
- **Resource Kind Aliases**: Added support for common resource aliases
  - HPA → HorizontalPodAutoscaler
  - PVC → PersistentVolumeClaim
  - PV → PersistentVolume
  - SC → StorageClass
  - SA → ServiceAccount
  - CR → ClusterRole
  - CRB → ClusterRoleBinding
  - RB → RoleBinding

### Changed

- **RBAC Permissions**: Updated to provide full access to all namespaced and cluster-scoped resources
  - Enables importing and managing any Kubernetes resource via YAML
  - Maintains security by keeping cluster resources read-only where appropriate
- **Dynamic Resource Discovery**: Enhanced resource resolution using Kubernetes Discovery API
  - Automatically finds correct GroupVersionResource for any resource type
  - Works with Custom Resource Definitions (CRDs) and unknown resource types
  - Falls back to static mapping if discovery fails

### Technical Improvements

- Added extensive logging for debugging resource resolution issues
- Improved error messages with detailed context
- Better handling of API version mismatches
- Enhanced validation and error recovery

---

## [1.1.0] - 2024-12-19

### 🏗️ Unified Architecture Release

This release introduces a major architectural improvement with enhanced security and simplified deployment.

### Architecture Changes

- **Unified Container**: Backend and Frontend integrated in a single Docker image
  - Eliminated inter-container communication
  - Reduced attack surface
  - Single process model for easier auditing
- **Simplified Deployment**: Single service, single deployment, single port (8080)
- **Easier Management**: One image to build, version, and deploy

### Technical Improvements

- Backend now serves frontend static files directly
- Simplified Helm chart with unified deployment
- Reduced resource overhead (single container instead of two)
- Single ingress path configuration (`/` instead of separate `/api` and `/` paths)

### Security Improvements

- **Reduced Attack Surface**: Unified architecture eliminates network communication between containers
- **Security Score**: 88/100 (comprehensive security audit completed)
- **Security Features**:
  - Argon2 password hashing with JWT-based authentication
  - HttpOnly cookies with Secure and SameSite=Strict flags
  - Comprehensive input validation (RFC 1123 for Kubernetes names)
  - Rate limiting (50 req/sec general, 5 req/min for login)
  - Audit logging for all API requests
  - Content Security Policy (CSP) headers
  - CORS validation with origin checking
  - WebSocket origin validation

### Changed

- Docker image changed from `dkonsole/dkonsole-backend` and `dkonsole/dkonsole-frontend` to `dkonsole/dkonsole`
- Helm values updated: `image.backend` and `image.frontend` replaced with single `image` configuration
- Ingress paths simplified: single path `/` instead of separate `/api` and `/` paths

### Migration Notes

**Breaking Changes:**
- If upgrading from v1.0.x, update your Helm values to use the new unified `image` configuration
- Update ingress configuration to use single path `/`
- Docker image repository changed to `dkonsole/dkonsole`

**Upgrade Path:**
```yaml
# Old configuration (v1.0.x)
image:
  backend:
    repository: dkonsole/dkonsole-backend
    tag: "1.0.7"
  frontend:
    repository: dkonsole/dkonsole-frontend
    tag: "1.0.7"

# New configuration (v1.1.0+)
image:
  repository: dkonsole/dkonsole
  tag: "1.1.0"
```

---

## [1.0.7] - 2025-11-23

### 🔒 Security Hardening Release

This release focuses on addressing critical security vulnerabilities and implementing enterprise-grade security measures.

### Critical Security Fixes

- **PromQL Injection Prevention**: Added strict input validation for all Prometheus queries using regex patterns to prevent injection attacks
- **Cookie-Based Authentication**: Migrated from localStorage to HttpOnly cookies for JWT tokens, eliminating XSS token theft risks
  - Cookies now use `HttpOnly`, `Secure`, and `SameSite=Strict` flags
  - Removed token from response body in LoginHandler
  - Added `/api/logout` and `/api/me` endpoints for proper session management
- **Kubernetes Name Validation**: Implemented RFC 1123 validation for all resource names and namespaces to prevent command injection
  - Applied to: GetResourceYAML, DeleteResource, ScaleResource, StreamPodLogs, ExecIntoPod
  - Uses native regex validation for better portability
- **Input Sanitization**: Added comprehensive validation across all API endpoints

### Security Enhancements

- **Audit Logging**: Implemented comprehensive audit middleware logging all API requests
  - Logs include: user, method, path, status code, duration
  - Enables tracking of "who did what and when" for compliance
- **Rate Limiting**: Added intelligent rate limiting (300 req/min per IP)
  - Prevents DoS attacks and brute force attempts
  - Automatically skips WebSocket upgrade requests
- **CSP Updates**: Enhanced Content Security Policy to allow Monaco Editor while maintaining security
  - Added `https://cdn.jsdelivr.net` to allowed sources
  - Added `worker-src 'self' blob:` for Monaco workers
- **JWT Secret Enforcement**: Strict validation requiring JWT_SECRET to be set and minimum 32 characters

### Bug Fixes

- Fixed table sorting functionality in WorkloadList component (was using filteredResources instead of sortedResources)
- Fixed YAML editor infinite loading issue caused by CSP restrictions
- Fixed Secret data display (now properly decodes byte arrays to strings)
- Removed duplicate mux initialization in main.go

### Architecture Improvements

- Unified middleware chain for consistent security policy application
- Separated public and secure route handlers with dedicated wrapper functions
- Enhanced error handling and validation across all handlers
- Created dedicated `middleware.go` file for better code organization

### Developer Experience

- Improved code organization with dedicated middleware.go
- Removed external validation dependency (`k8s.io/apimachinery/pkg/util/validation`)
- Better separation of concerns between authentication, auditing, and rate limiting
- Cleaner route registration with `secure()` and `public()` helpers

### Changed

- Frontend now exclusively uses cookie-based authentication
- AuthContext refactored to use `/api/me` for session validation
- All API calls now use `authFetch` wrapper for consistent auth handling

---

## [1.0.6] - 2025-11-22

### Added

- Enhanced Pod metrics with Network RX/TX and PVC usage
- Improved Cluster Overview with Prometheus integration
- Added real-time node metrics table with CPU, Memory, Disk, and Network stats
- Added cluster-wide statistics (avg CPU/Memory, network traffic, trends)
- Added fadeIn animation for smooth tab transitions

### Fixed

- Fixed visual issues in Pod metrics tabs (removed extra border, smooth transitions)
- Conditional rendering of metrics based on data availability

---

## [1.0.5] - 2025-11-22

### Added

- Added build and release scripts in `scripts/` directory
- Scripts documentation available in `scripts/SCRIPTS.md`

### Changed

- Improved build and release scripts

### Deprecated

- Deprecated `deploy.sh` in favor of new scripts

---

## [1.0.4] - 2025-11-22

### Added

- Prometheus integration for Pod metrics
- Historical metrics with time range selector (1h, 6h, 12h, 1d, 7d, 15d)
- Metrics tab to Pod details

### Fixed

- Fixed namespace display for cluster-scoped resources (Nodes, ClusterRoles, etc.)

### Removed

- Removed Metrics tab from Deployment details (kept for future use)

---

## [1.0.3] - 2025-11-21

### Added

- Initial stable release
- Resource management for Deployments, Pods, Services, ConfigMaps, Secrets
- Live log streaming
- Terminal access to pods
- YAML editor
- Argon2 password hashing
- JWT-based authentication
- Multi-cluster support

---

[1.1.5]: https://github.com/flaucha/DKonsole/compare/v1.1.4...v1.1.5
[1.1.4]: https://github.com/flaucha/DKonsole/compare/v1.1.3...v1.1.4
[1.1.3]: https://github.com/flaucha/DKonsole/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/flaucha/DKonsole/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/flaucha/DKonsole/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/flaucha/DKonsole/compare/v1.0.7...v1.1.0
[1.0.7]: https://github.com/flaucha/DKonsole/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/flaucha/DKonsole/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/flaucha/DKonsole/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/flaucha/DKonsole/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/flaucha/DKonsole/releases/tag/v1.0.3
