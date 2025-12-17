package server

import (
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/api"
	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
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
	mux := http.NewServeMux()

	authService := deps.AuthService
	ldapService := deps.LDAPService

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

	secureWS := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.SecurityHeadersMiddleware(middleware.RateLimitMiddleware(middleware.AuditMiddleware(authService.AuthMiddleware(h))))
	}

	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			isAdmin, err := permissions.IsAdmin(ctx, ldapService)
			if err != nil {
				utils.HandleErrorJSON(w, err, "Failed to check admin status", http.StatusInternalServerError, nil)
				return
			}
			if !isAdmin {
				utils.ErrorResponse(w, http.StatusForbidden, "Admin access required")
				return
			}
			h(w, r)
		}
	}

	config := RouterConfig{
		Deps:          deps,
		Mux:           mux,
		Secure:        secure,
		Public:        public,
		SecureHandler: secureHandler,
		SecureWS:      secureWS,
		AdminOnly:     adminOnly,
	}

	registerRoutes(config)
	registerHealthRoutes(config)
	registerK8sRoutes(config)
	registerAPIRoutes(config)
	registerHelmRoutes(config)
	registerPodRoutes(config)
	registerOtherRoutes(config)
	registerStaticRoutes(config)

	return mux
}
