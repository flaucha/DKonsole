package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/health"
	"github.com/flaucha/DKonsole/backend/internal/middleware"

	"github.com/flaucha/DKonsole/backend/internal/utils"

	httpSwagger "github.com/swaggo/http-swagger"
)

type RouterConfig struct {
	Deps Dependencies
	Mux  *http.ServeMux
	Secure func(http.HandlerFunc) http.HandlerFunc
	Public func(http.HandlerFunc) http.HandlerFunc
	SecureHandler func(http.Handler) http.Handler
	SecureWS func(http.HandlerFunc) http.HandlerFunc
	AdminOnly func(http.HandlerFunc) http.HandlerFunc
}

func registerRoutes(c RouterConfig) {
	// Setup endpoints
	setupCompleteHandler := func(w http.ResponseWriter, r *http.Request) {
		utils.LogInfo("Request received for /api/setup/complete", map[string]interface{}{
			"method":  r.Method,
			"ip":      r.RemoteAddr,
			"origin":  r.Header.Get("Origin"),
			"referer": r.Header.Get("Referer"),
		})
		c.Deps.AuthService.SetupCompleteHandler(w, r)
	}
	c.Mux.HandleFunc("/api/setup/status", c.Public(c.Deps.AuthService.SetupStatusHandler))
	c.Mux.HandleFunc("/api/setup/complete", c.Public(setupCompleteHandler))

	// Auth endpoints
	c.Mux.HandleFunc("/api/login", middleware.SecurityHeadersMiddleware(enableCors(middleware.LoginRateLimitMiddleware(middleware.AuditMiddleware(c.Deps.AuthService.LoginHandler)))))
	c.Mux.HandleFunc("/api/logout", c.Public(c.Deps.AuthService.LogoutHandler))
	c.Mux.HandleFunc("/api/me", c.Secure(c.Deps.AuthService.MeHandler))
	c.Mux.HandleFunc("/api/auth/change-password", c.Secure(c.Deps.AuthService.ChangePasswordHandler))

	// Swagger documentation - protected with authentication
	c.Mux.Handle("/swagger/", c.SecureHandler(httpSwagger.WrapHandler))
}

func registerHealthRoutes(c RouterConfig) {
	c.Mux.HandleFunc("/healthz", middleware.SecurityHeadersMiddleware(health.HealthHandler))
	c.Mux.HandleFunc("/health", middleware.SecurityHeadersMiddleware(health.HealthHandler))
	c.Mux.HandleFunc("/readyz", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		setupRequired := false
		if c.Deps.AuthService != nil && c.Deps.AuthService.IsSetupMode() {
			setupRequired = true
		}

		handlersModel := c.Deps.HandlersModel
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
		if c.Deps.PrometheusService != nil {
			prometheusStatus["enabled"] = c.Deps.PrometheusService.IsConfigured()
			if c.Deps.PrometheusService.IsConfigured() {
				if err := c.Deps.PrometheusService.HealthCheck(ctx); err != nil {
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
}

func registerK8sRoutes(c RouterConfig) {
	c.Mux.HandleFunc("/api/namespaces", c.Secure(c.Deps.K8sService.GetNamespaces))
	c.Mux.HandleFunc("/api/resources", c.Secure(c.Deps.K8sService.GetResources))
	c.Mux.HandleFunc("/api/resources/watch", c.Secure(c.Deps.K8sService.WatchResources))
	c.Mux.HandleFunc("/api/resource/yaml", c.Secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Deps.K8sService.GetResourceYAML(w, r)
		} else if r.Method == http.MethodPut {
			c.Deps.K8sService.UpdateResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	c.Mux.HandleFunc("/api/resource/import", c.Secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			c.Deps.K8sService.ImportResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	c.Mux.HandleFunc("/api/scale", c.Secure(c.Deps.K8sService.ScaleResource))
	c.Mux.HandleFunc("/api/deployments/rollout", c.Secure(c.Deps.K8sService.RolloutDeployment))
	c.Mux.HandleFunc("/api/overview", c.Secure(c.Deps.K8sService.GetClusterStats))
	c.Mux.HandleFunc("/api/resource", c.Secure(c.Deps.K8sService.DeleteResource))
	c.Mux.HandleFunc("/api/cronjobs/trigger", c.Secure(c.Deps.K8sService.TriggerCronJob))
}

func registerAPIRoutes(c RouterConfig) {
	c.Mux.HandleFunc("/api/apis", c.Secure(c.Deps.APIService.ListAPIResources))
	c.Mux.HandleFunc("/api/apis/resources", c.Secure(c.Deps.APIService.ListAPIResourceObjects))
	c.Mux.HandleFunc("/api/apis/yaml", c.Secure(c.Deps.APIService.GetAPIResourceYAML))
	c.Mux.HandleFunc("/api/crds", c.Secure(c.Deps.APIService.GetCRDs))
	c.Mux.HandleFunc("/api/crds/resources", c.Secure(c.Deps.APIService.GetCRDResources))
	c.Mux.HandleFunc("/api/crds/yaml", c.Secure(c.Deps.APIService.GetCRDYaml))
}

func registerHelmRoutes(c RouterConfig) {
	c.Mux.HandleFunc("/api/helm/releases", c.Secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Deps.HelmService.GetHelmReleases(w, r)
		} else if r.Method == http.MethodDelete {
			c.Deps.HelmService.DeleteHelmRelease(w, r)
		} else if r.Method == http.MethodPost {
			c.Deps.HelmService.UpgradeHelmRelease(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	c.Mux.HandleFunc("/api/helm/releases/install", c.Secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			c.Deps.HelmService.InstallHelmRelease(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

func registerPodRoutes(c RouterConfig) {
	c.Mux.HandleFunc("/api/pods/logs", c.Secure(middleware.WebSocketLimitMiddleware(c.Deps.PodService.StreamPodLogs)))
	c.Mux.HandleFunc("/api/pods/events", c.Secure(c.Deps.PodService.GetPodEvents))
	c.Mux.HandleFunc("/api/pods/exec", c.SecureWS(middleware.WebSocketLimitMiddleware(c.Deps.PodService.ExecIntoPod)))
}

func registerOtherRoutes(c RouterConfig) {
	// Logo handlers
	c.Mux.HandleFunc("/api/logo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Public(func(w http.ResponseWriter, r *http.Request) {
				c.Deps.LogoService.GetLogo(w, r)
			})(w, r)
		} else if r.Method == http.MethodPost {
			c.Secure(func(w http.ResponseWriter, r *http.Request) {
				c.Deps.LogoService.UploadLogo(w, r)
			})(w, r)
		} else if r.Method == http.MethodDelete {
			c.Secure(func(w http.ResponseWriter, r *http.Request) {
				c.Deps.LogoService.DeleteLogo(w, r)
			})(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Prometheus handlers
	c.Mux.HandleFunc("/api/prometheus/status", c.Secure(c.Deps.PrometheusService.GetStatus))
	c.Mux.HandleFunc("/api/prometheus/metrics", c.Secure(c.Deps.PrometheusService.GetMetrics))
	c.Mux.HandleFunc("/api/prometheus/pod-metrics", c.Secure(c.Deps.PrometheusService.GetPodMetrics))
	c.Mux.HandleFunc("/api/prometheus/cluster-overview", c.Secure(c.Deps.PrometheusService.GetClusterOverview))

	// Settings handlers
	c.Mux.HandleFunc("/api/settings/prometheus/url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Secure(c.AdminOnly(c.Deps.SettingsService.GetPrometheusURLHandler))(w, r)
		} else if r.Method == http.MethodPut {
			c.Secure(c.AdminOnly(c.Deps.SettingsService.UpdatePrometheusURLHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// LDAP handlers
	c.Mux.HandleFunc("/api/ldap/status", c.Public(c.Deps.LDAPService.GetLDAPStatusHandler))
	c.Mux.HandleFunc("/api/ldap/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Secure(c.AdminOnly(c.Deps.LDAPService.GetConfigHandler))(w, r)
		} else if r.Method == http.MethodPut {
			c.Secure(c.AdminOnly(c.Deps.LDAPService.UpdateConfigHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	c.Mux.HandleFunc("/api/ldap/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Secure(c.AdminOnly(c.Deps.LDAPService.GetGroupsHandler))(w, r)
		} else if r.Method == http.MethodPut {
			c.Secure(c.AdminOnly(c.Deps.LDAPService.UpdateGroupsHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	c.Mux.HandleFunc("/api/ldap/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			c.Secure(c.AdminOnly(c.Deps.LDAPService.GetCredentialsHandler))(w, r)
		} else if r.Method == http.MethodPut {
			c.Secure(c.AdminOnly(c.Deps.LDAPService.UpdateCredentialsHandler))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	c.Mux.HandleFunc("/api/ldap/test", c.Secure(c.AdminOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			c.Deps.LDAPService.TestConnectionHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
}
