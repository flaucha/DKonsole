package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/auth"
	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/k8s"
	"github.com/example/k8s-view/internal/api"
	"github.com/example/k8s-view/internal/helm"
	"github.com/example/k8s-view/internal/pod"
)

// ResponseWriter wrapper to fix Content-Type after FileServer sets it
type contentTypeFixer struct {
	http.ResponseWriter
	path string
}

func (c *contentTypeFixer) WriteHeader(code int) {
	// Fix Content-Type based on file extension before writing headers
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
		// Try to detect MIME type from extension
		if ct := mime.TypeByExtension(ext); ct != "" {
			c.ResponseWriter.Header().Set("Content-Type", ct)
		}
	}
	c.ResponseWriter.WriteHeader(code)
}

// setupHandlerDelegates connects service methods to old handler implementations
// This is temporary during migration
func setupHandlerDelegates(h *Handlers, k8sSvc *k8s.Service, apiSvc *api.Service, helmSvc *helm.Service, podSvc *pod.Service) {
	// Services will delegate to handlers.go methods through h
	// This maintains functionality while we migrate
	_ = h // Used by services
	_ = k8sSvc
	_ = apiSvc
	_ = helmSvc
	_ = podSvc
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Error building kubeconfig from flags: %v", err)
		log.Println("Falling back to in-cluster config...")
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	metricsClient, _ := metricsv.NewForConfig(config)

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	handlersModel := &models.Handlers{
		Clients:       make(map[string]*kubernetes.Clientset),
		Dynamics:      make(map[string]dynamic.Interface),
		Metrics:       make(map[string]*metricsv.Clientset),
		RESTConfigs:   make(map[string]*rest.Config),
		PrometheusURL: os.Getenv("PROMETHEUS_URL"),
	}
	handlersModel.Clients["default"] = clientset
	handlersModel.Dynamics["default"] = dynamicClient
	handlersModel.RESTConfigs["default"] = config
	if metricsClient != nil {
		handlersModel.Metrics["default"] = metricsClient
	}

	// Create wrapper for backward compatibility with handlers not yet migrated
	h := &Handlers{Handlers: handlersModel}

	// Initialize services
	authService := auth.NewService(handlersModel)
	clusterService := cluster.NewService(handlersModel)
	k8sService := k8s.NewService(handlersModel, clusterService)
	apiService := api.NewService(handlersModel, clusterService)
	helmService := helm.NewService(handlersModel, clusterService)
	podService := pod.NewService(handlersModel, clusterService)
	
	// authenticateRequest is now handled by authService.AuthenticateRequest
	// No need for a global wrapper variable
	
	// Set up handler delegates for services that need access to old handlers
	setupHandlerDelegates(h, k8sService, apiService, helmService, podService)

	mux := http.NewServeMux()
	// Helper for authenticated routes
	secure := func(h http.HandlerFunc) http.HandlerFunc {
		return enableCors(RateLimitMiddleware(CSRFMiddleware(AuditMiddleware(authService.AuthMiddleware(h)))))
	}

	// Helper for public routes
	public := func(h http.HandlerFunc) http.HandlerFunc {
		return enableCors(RateLimitMiddleware(AuditMiddleware(h)))
	}

	mux.HandleFunc("/api/login", enableCors(LoginRateLimitMiddleware(AuditMiddleware(authService.LoginHandler))))
	mux.HandleFunc("/api/logout", public(authService.LogoutHandler))
	mux.HandleFunc("/api/me", secure(authService.MeHandler))
	mux.HandleFunc("/healthz", h.HealthHandler)
	
	// K8s handlers - using services
	mux.HandleFunc("/api/namespaces", secure(k8sService.GetNamespaces))
	mux.HandleFunc("/api/resources", secure(k8sService.GetResources))
	mux.HandleFunc("/api/resources/watch", secure(k8sService.WatchResources))
	mux.HandleFunc("/api/resource/yaml", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			k8sService.GetResourceYAML(w, r)
		} else if r.Method == http.MethodPut {
			k8sService.UpdateResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/resource/import", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			k8sService.ImportResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	
	// API handlers - using services
	mux.HandleFunc("/api/apis", secure(apiService.ListAPIResources))
	mux.HandleFunc("/api/apis/resources", secure(apiService.ListAPIResourceObjects))
	mux.HandleFunc("/api/apis/yaml", secure(apiService.GetAPIResourceYAML))
	mux.HandleFunc("/api/crds", secure(apiService.GetCRDs))
	mux.HandleFunc("/api/crds/resources", secure(apiService.GetCRDResources))
	mux.HandleFunc("/api/crds/yaml", secure(apiService.GetCRDYaml))
	
	mux.HandleFunc("/api/scale", secure(k8sService.ScaleResource))
	mux.HandleFunc("/api/overview", secure(k8sService.GetClusterStats))
	
	// Helm handlers - using services
	mux.HandleFunc("/api/helm/releases", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			helmService.GetHelmReleases(w, r)
		} else if r.Method == http.MethodDelete {
			helmService.DeleteHelmRelease(w, r)
		} else if r.Method == http.MethodPost {
			helmService.UpgradeHelmRelease(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/helm/releases/install", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			helmService.InstallHelmRelease(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	
	// Pod handlers - using services
	mux.HandleFunc("/api/pods/logs", secure(podService.StreamPodLogs))
	mux.HandleFunc("/api/pods/events", secure(podService.GetPodEvents))
	mux.HandleFunc("/api/pods/exec", secure(podService.ExecIntoPod))
	
	mux.HandleFunc("/api/clusters", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clusterService.GetClusters(w, r)
		} else if r.Method == http.MethodPost {
			clusterService.AddCluster(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/resource", secure(k8sService.DeleteResource))
	mux.HandleFunc("/api/cronjobs/trigger", secure(k8sService.TriggerCronJob))
	mux.HandleFunc("/api/logo", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetLogo(w, r)
		} else if r.Method == http.MethodPost {
			h.UploadLogo(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/prometheus/status", secure(h.GetPrometheusStatus))
	mux.HandleFunc("/api/prometheus/metrics", secure(h.GetPrometheusMetrics))
	mux.HandleFunc("/api/prometheus/pod-metrics", secure(h.GetPrometheusPodMetrics))
	mux.HandleFunc("/api/prometheus/cluster-overview", secure(h.GetPrometheusClusterOverview))

	// Serve static files from frontend build
	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("Warning: static directory '%s' not found, frontend will not be served", staticDir)
	} else {
		// Middleware to add security headers and correct MIME types for static files
		secureStaticFiles := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Wrap ResponseWriter to fix Content-Type
				fixer := &contentTypeFixer{
					ResponseWriter: w,
					path:           r.URL.Path,
				}
				
				// Add security headers
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
				w.Header().Set("X-Frame-Options", "SAMEORIGIN")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
				if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
					w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
				}
				next.ServeHTTP(fixer, r)
			})
		}

		assetsDir := filepath.Join(staticDir, "assets")
		fileServer := http.FileServer(http.Dir(assetsDir))
		mux.Handle("/assets/", secureStaticFiles(http.StripPrefix("/assets/", fileServer)))
		
		mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, r, filepath.Join(staticDir, "favicon.ico"))
		})
		mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, r, filepath.Join(staticDir, "robots.txt"))
		})
		
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
			} else {
				http.NotFound(w, r)
			}
		})
	}

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

		if origin == "" && r.Method != "OPTIONS" {
			next(w, r)
			return
		}

		allowed := false
		if allowedOrigins != "" {
			origins := strings.Split(allowedOrigins, ",")
			for _, o := range origins {
				o = strings.TrimSpace(o)
				if o == origin {
					allowed = true
					break
				}
			}
		} else {
			if origin != "" {
				originURL, err := url.Parse(origin)
				if err == nil {
					host := r.Host
					if strings.Contains(host, ":") {
						host = strings.Split(host, ":")[0]
					}
					originHost := originURL.Host
					if strings.Contains(originHost, ":") {
						originHost = strings.Split(originHost, ":")[0]
					}
					if (originHost == "localhost" || originHost == "127.0.0.1" || originHost == host) &&
						(originURL.Scheme == "http" || originURL.Scheme == "https") {
						allowed = true
					}
				}
			}
		}

		if !allowed && origin != "" {
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
