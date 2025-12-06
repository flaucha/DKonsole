package server

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/flaucha/DKonsole/backend/internal/api"
	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/health"
	"github.com/flaucha/DKonsole/backend/internal/helm"
	"github.com/flaucha/DKonsole/backend/internal/k8s"
	"github.com/flaucha/DKonsole/backend/internal/ldap"
	"github.com/flaucha/DKonsole/backend/internal/logo"
	"github.com/flaucha/DKonsole/backend/internal/middleware"
	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/pod"
	"github.com/flaucha/DKonsole/backend/internal/prometheus"
	"github.com/flaucha/DKonsole/backend/internal/settings"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Dependencies bundles the services and configuration required to build the HTTP router.
type Dependencies struct {
	AuthService       *auth.Service
	LDAPService       *ldap.Service
	ClusterService    *cluster.Service
	K8sService        *k8s.Service
	APIService        *api.Service
	HelmService       *helm.Service
	PodService        *pod.Service
	PrometheusService *prometheus.HTTPHandler
	SettingsService   *settings.Service
	LogoService       *logo.Service
	HandlersModel     *models.Handlers
	StaticDir         string
}

// NewRouter builds and returns the HTTP mux with all routes and middleware wired.
func NewRouter(deps Dependencies) *http.ServeMux {
	staticDir := deps.StaticDir
	if staticDir == "" {
		staticDir = "static"
	}

	mux := http.NewServeMux()

	authService := deps.AuthService
	ldapService := deps.LDAPService
	settingsService := deps.SettingsService
	handlersModel := deps.HandlersModel

	secure := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.SecurityHeadersMiddleware(enableCors(middleware.RateLimitMiddleware(middleware.CSRFMiddleware(middleware.AuditMiddleware(authService.AuthMiddleware(h))))))
	}

	secureHandler := func(h http.Handler) http.Handler {
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
		return middleware.SecurityHeadersHandler(http.HandlerFunc(
			enableCors(middleware.RateLimitMiddleware(middleware.CSRFMiddleware(middleware.AuditMiddleware(authService.AuthMiddleware(handlerFunc)))))),
		)
	}

	public := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.SecurityHeadersMiddleware(enableCors(middleware.RateLimitMiddleware(middleware.AuditMiddleware(h))))
	}

	// Setup endpoints
	setupCompleteHandler := func(w http.ResponseWriter, r *http.Request) {
		utils.LogInfo("Request received for /api/setup/complete", map[string]interface{}{
			"method":  r.Method,
			"ip":      r.RemoteAddr,
			"origin":  r.Header.Get("Origin"),
			"referer": r.Header.Get("Referer"),
		})
		authService.SetupCompleteHandler(w, r)
	}
	mux.HandleFunc("/api/setup/status", public(authService.SetupStatusHandler))
	mux.HandleFunc("/api/setup/complete", public(setupCompleteHandler))

	// Auth endpoints
	mux.HandleFunc("/api/login", middleware.SecurityHeadersMiddleware(enableCors(middleware.LoginRateLimitMiddleware(middleware.AuditMiddleware(authService.LoginHandler)))))
	mux.HandleFunc("/api/logout", public(authService.LogoutHandler))
	mux.HandleFunc("/api/me", secure(authService.MeHandler))

	// Health endpoints
	mux.HandleFunc("/healthz", middleware.SecurityHeadersMiddleware(health.HealthHandler))
	mux.HandleFunc("/health", middleware.SecurityHeadersMiddleware(health.HealthHandler))
	mux.HandleFunc("/readyz", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		setupRequired := false
		if authService != nil && authService.IsSetupMode() {
			setupRequired = true
		}

		if handlersModel == nil {
			utils.ErrorResponse(w, http.StatusServiceUnavailable, "handlers model not initialized")
			return
		}

		handlersModel.RLock()
		clientset := handlersModel.Clients["default"]
		handlersModel.RUnlock()
		if clientset == nil {
			utils.ErrorResponse(w, http.StatusServiceUnavailable, "kubernetes client not initialized")
			return
		}

		discoveryClient := clientset.Discovery()
		if discoveryClient == nil {
			utils.ErrorResponse(w, http.StatusServiceUnavailable, "kubernetes discovery client not initialized")
			return
		}

		if _, err := discoveryClient.ServerVersion(); err != nil {
			utils.ErrorResponse(w, http.StatusServiceUnavailable, fmt.Sprintf("kubernetes api unreachable: %v", err))
			return
		}

		prometheusStatus := map[string]interface{}{
			"enabled": false,
			"healthy": true,
		}
		if deps.PrometheusService != nil {
			prometheusStatus["enabled"] = deps.PrometheusService.IsConfigured()
			if deps.PrometheusService.IsConfigured() {
				if err := deps.PrometheusService.HealthCheck(ctx); err != nil {
					utils.ErrorResponse(w, http.StatusServiceUnavailable, fmt.Sprintf("prometheus unhealthy: %v", err))
					return
				}
			} else {
				prometheusStatus["healthy"] = false
			}
		}

		utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"status":        "ready",
			"setupRequired": setupRequired,
			"prometheus":    prometheusStatus,
		})
	}))

	// Swagger documentation - protected with authentication
	mux.Handle("/swagger/", secureHandler(httpSwagger.WrapHandler))

	// WebSocket endpoint - secure but without CORS/CSRF (handled by Upgrader)
	secureWS := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.SecurityHeadersMiddleware(middleware.RateLimitMiddleware(middleware.AuditMiddleware(authService.AuthMiddleware(h))))
	}

	// K8s handlers - using services
	mux.HandleFunc("/api/namespaces", secure(deps.K8sService.GetNamespaces))
	mux.HandleFunc("/api/resources", secure(deps.K8sService.GetResources))
	mux.HandleFunc("/api/resources/watch", secure(deps.K8sService.WatchResources))
	mux.HandleFunc("/api/resource/yaml", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			deps.K8sService.GetResourceYAML(w, r)
		} else if r.Method == http.MethodPut {
			deps.K8sService.UpdateResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/resource/import", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			deps.K8sService.ImportResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// API handlers - using services
	mux.HandleFunc("/api/apis", secure(deps.APIService.ListAPIResources))
	mux.HandleFunc("/api/apis/resources", secure(deps.APIService.ListAPIResourceObjects))
	mux.HandleFunc("/api/apis/yaml", secure(deps.APIService.GetAPIResourceYAML))
	mux.HandleFunc("/api/crds", secure(deps.APIService.GetCRDs))
	mux.HandleFunc("/api/crds/resources", secure(deps.APIService.GetCRDResources))
	mux.HandleFunc("/api/crds/yaml", secure(deps.APIService.GetCRDYaml))

	mux.HandleFunc("/api/scale", secure(deps.K8sService.ScaleResource))
	mux.HandleFunc("/api/deployments/rollout", secure(deps.K8sService.RolloutDeployment))
	mux.HandleFunc("/api/overview", secure(deps.K8sService.GetClusterStats))

	// Helm handlers - using services
	mux.HandleFunc("/api/helm/releases", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			deps.HelmService.GetHelmReleases(w, r)
		} else if r.Method == http.MethodDelete {
			deps.HelmService.DeleteHelmRelease(w, r)
		} else if r.Method == http.MethodPost {
			deps.HelmService.UpgradeHelmRelease(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/helm/releases/install", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			deps.HelmService.InstallHelmRelease(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Pod handlers - using services
	mux.HandleFunc("/api/pods/logs", secure(middleware.WebSocketLimitMiddleware(deps.PodService.StreamPodLogs)))
	mux.HandleFunc("/api/pods/events", secure(deps.PodService.GetPodEvents))
	mux.HandleFunc("/api/pods/exec", secureWS(middleware.WebSocketLimitMiddleware(deps.PodService.ExecIntoPod)))

	mux.HandleFunc("/api/resource", secure(deps.K8sService.DeleteResource))
	mux.HandleFunc("/api/cronjobs/trigger", secure(deps.K8sService.TriggerCronJob))

	// Logo handlers - using service
	mux.HandleFunc("/api/logo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			public(func(w http.ResponseWriter, r *http.Request) {
				deps.LogoService.GetLogo(w, r)
			})(w, r)
		} else if r.Method == http.MethodPost {
			secure(func(w http.ResponseWriter, r *http.Request) {
				deps.LogoService.UploadLogo(w, r)
			})(w, r)
		} else if r.Method == http.MethodDelete {
			secure(func(w http.ResponseWriter, r *http.Request) {
				deps.LogoService.DeleteLogo(w, r)
			})(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Prometheus handlers - using services
	mux.HandleFunc("/api/prometheus/status", secure(deps.PrometheusService.GetStatus))
	mux.HandleFunc("/api/prometheus/metrics", secure(deps.PrometheusService.GetMetrics))
	mux.HandleFunc("/api/prometheus/pod-metrics", secure(deps.PrometheusService.GetPodMetrics))
	mux.HandleFunc("/api/prometheus/cluster-overview", secure(deps.PrometheusService.GetClusterOverview))

	// Settings handlers - protected with admin check
	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			isAdmin, err := permissions.IsAdmin(ctx, ldapService)
			if err != nil {
				utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check admin status: %v", err))
				return
			}
			if !isAdmin {
				utils.ErrorResponse(w, http.StatusForbidden, "Admin access required")
				return
			}
			h(w, r)
		}
	}

	mux.HandleFunc("/api/settings/prometheus/url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			secure(adminOnly(settingsService.GetPrometheusURLHandler))(w, r)
		} else if r.Method == http.MethodPut {
			secure(adminOnly(settingsService.UpdatePrometheusURLHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Auth handlers - password change
	mux.HandleFunc("/api/auth/change-password", secure(authService.ChangePasswordHandler))

	// LDAP handlers - using services (protected with admin check)
	mux.HandleFunc("/api/ldap/status", public(ldapService.GetLDAPStatusHandler))
	mux.HandleFunc("/api/ldap/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			secure(adminOnly(ldapService.GetConfigHandler))(w, r)
		} else if r.Method == http.MethodPut {
			secure(adminOnly(ldapService.UpdateConfigHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/ldap/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			secure(adminOnly(ldapService.GetGroupsHandler))(w, r)
		} else if r.Method == http.MethodPut {
			secure(adminOnly(ldapService.UpdateGroupsHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/ldap/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			secure(adminOnly(ldapService.GetCredentialsHandler))(w, r)
		} else if r.Method == http.MethodPut {
			secure(adminOnly(ldapService.UpdateCredentialsHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/ldap/test", secure(adminOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ldapService.TestConnectionHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Serve static files from frontend build
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		utils.LogWarn("Static directory not found, frontend will not be served", map[string]interface{}{
			"static_dir": staticDir,
		})
	} else {
		secureStaticFiles := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fixer := &contentTypeFixer{
					ResponseWriter: w,
					path:           r.URL.Path,
				}
				middleware.SecurityHeadersHandler(next).ServeHTTP(fixer, r)
			})
		}

		assetsDir := filepath.Join(staticDir, "assets")
		fileServer := http.FileServer(http.Dir(assetsDir))
		mux.Handle("/assets/", secureStaticFiles(http.StripPrefix("/assets/", fileServer)))

		mux.HandleFunc("/favicon.ico", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "favicon.ico"))
		}))
		mux.HandleFunc("/favicon.png", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			faviconPath := filepath.Join(staticDir, "favicon.png")
			if _, err := os.Stat(faviconPath); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeFile(w, r, faviconPath)
		}))
		mux.HandleFunc("/logo-full-dark.png", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			logoPath := filepath.Join(staticDir, "logo-full-dark.png")
			if _, err := os.Stat(logoPath); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeFile(w, r, logoPath)
		}))
		mux.HandleFunc("/logo-full-light.png", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			logoPath := filepath.Join(staticDir, "logo-full-light.png")
			if _, err := os.Stat(logoPath); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeFile(w, r, logoPath)
		}))
		mux.HandleFunc("/robots.txt", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "robots.txt"))
		}))

		mux.HandleFunc("/", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				http.NotFound(w, r)
				return
			}
			if r.URL.Path == "/favicon.ico" || r.URL.Path == "/favicon.png" || r.URL.Path == "/logo-full-dark.png" || r.URL.Path == "/logo-full-light.png" || r.URL.Path == "/robots.txt" {
				http.NotFound(w, r)
				return
			}
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
			} else {
				http.NotFound(w, r)
			}
		}))
	}

	return mux
}

// ResponseWriter wrapper to fix Content-Type after FileServer sets it
type contentTypeFixer struct {
	http.ResponseWriter
	path string
}

func (c *contentTypeFixer) WriteHeader(code int) {
	ext := filepath.Ext(c.path)
	switch ext {
	case ".js":
		c.ResponseWriter.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".css":
		c.ResponseWriter.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".json":
		c.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		c.ResponseWriter.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		c.ResponseWriter.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		c.ResponseWriter.Header().Set("Content-Type", "image/svg+xml")
	case ".woff", ".woff2":
		c.ResponseWriter.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		c.ResponseWriter.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		c.ResponseWriter.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	default:
		if ct := mime.TypeByExtension(ext); ct != "" {
			c.ResponseWriter.Header().Set("Content-Type", ct)
		}
	}
	c.ResponseWriter.WriteHeader(code)
}

// enableCors applies CORS headers allowing configured origins or host match
func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

		if origin == "" && r.Method != "OPTIONS" {
			next(w, r)
			return
		}

		allowed := false
		if origin != "" {
			// Exact matches from allowlist
			if allowedOrigins != "" {
				origins := strings.Split(allowedOrigins, ",")
				for _, o := range origins {
					o = strings.TrimSpace(o)
					if o != "" && strings.EqualFold(o, origin) {
						allowed = true
						break
					}
				}
			}

			// Always allow same-origin requests (host match)
			if !allowed {
				originURL, err := url.Parse(origin)
				if err == nil {
					host := r.Host
					if strings.Contains(host, ":") {
						host, _, _ = strings.Cut(host, ":")
					}
					originHost := originURL.Host
					if strings.Contains(originHost, ":") {
						originHost, _, _ = strings.Cut(originHost, ":")
					}
					if strings.EqualFold(strings.TrimSpace(originHost), strings.TrimSpace(host)) &&
						(originURL.Scheme == "http" || originURL.Scheme == "https") {
						allowed = true
					}
				}
			}
		}

		if !allowed && origin != "" {
			utils.LogWarn("CORS: Origin not allowed", map[string]interface{}{
				"origin": origin,
				"host":   r.Host,
				"path":   r.URL.Path,
			})
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
