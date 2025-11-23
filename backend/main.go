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

	h := &Handlers{
		Clients:       make(map[string]*kubernetes.Clientset),
		Dynamics:      make(map[string]dynamic.Interface),
		Metrics:       make(map[string]*metricsv.Clientset),
		RESTConfigs:   make(map[string]*rest.Config),
		PrometheusURL: os.Getenv("PROMETHEUS_URL"),
	}
	h.Clients["default"] = clientset
	h.Dynamics["default"] = dynamicClient
	h.RESTConfigs["default"] = config
	if metricsClient != nil {
		h.Metrics["default"] = metricsClient
	}

	mux := http.NewServeMux()
	// Helper for authenticated routes
	secure := func(h http.HandlerFunc) http.HandlerFunc {
		return enableCors(RateLimitMiddleware(AuditMiddleware(AuthMiddleware(h))))
	}

	// Helper for public routes
	public := func(h http.HandlerFunc) http.HandlerFunc {
		return enableCors(RateLimitMiddleware(AuditMiddleware(h)))
	}

	mux.HandleFunc("/api/login", enableCors(LoginRateLimitMiddleware(AuditMiddleware(h.LoginHandler))))
	mux.HandleFunc("/api/logout", public(h.LogoutHandler)) // Logout doesn't strictly need auth check if we just clear cookie, but usually it's fine.
	mux.HandleFunc("/api/me", secure(h.MeHandler))
	mux.HandleFunc("/healthz", h.HealthHandler) // Health check usually no auth/audit/limit or loose limit
	mux.HandleFunc("/api/namespaces", secure(h.GetNamespaces))
	mux.HandleFunc("/api/resources", secure(h.GetResources))
	mux.HandleFunc("/api/resources/watch", secure(h.WatchResources))
	mux.HandleFunc("/api/resource/yaml", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetResourceYAML(w, r)
		} else if r.Method == http.MethodPut {
			h.UpdateResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/resource/import", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.ImportResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/apis", secure(h.ListAPIResources))
	mux.HandleFunc("/api/apis/resources", secure(h.ListAPIResourceObjects))
	mux.HandleFunc("/api/apis/yaml", secure(h.GetAPIResourceYAML))
	mux.HandleFunc("/api/scale", secure(h.ScaleResource))
	mux.HandleFunc("/api/overview", secure(h.GetClusterStats))
	// StreamPodLogs and ExecIntoPod handle their own auth/upgrade, but we can wrap them with Audit/RateLimit.
	// Note: ExecIntoPod uses WebSocket, RateLimit should skip or be high. Audit is good.
	// They use "authenticateRequest" internally? Let's check.
	// Yes, they do. So we can use secure() but we need to make sure AuthMiddleware doesn't break WS upgrade or double-check.
	// AuthMiddleware checks token. ExecIntoPod checks token. Double check is fine.
	// However, ExecIntoPod might need to handle the error differently (WS close).
	// Let's stick to the existing pattern for Pods but add Audit/RateLimit.
	// Actually, StreamPodLogs is HTTP stream, Exec is WS.
	// I'll wrap them with secure() for consistency, assuming AuthMiddleware handles the token check fine.
	mux.HandleFunc("/api/pods/logs", secure(h.StreamPodLogs))
	mux.HandleFunc("/api/pods/exec", func(w http.ResponseWriter, r *http.Request) {
		// WebSocket endpoint needs CORS for browser connections
		// Auth is handled inside ExecIntoPod
		enableCors(RateLimitMiddleware(AuditMiddleware(h.ExecIntoPod)))(w, r)
	})
	mux.HandleFunc("/api/clusters", secure(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetClusters(w, r)
		} else if r.Method == http.MethodPost {
			h.AddCluster(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/resource", secure(h.DeleteResource))
	mux.HandleFunc("/api/cronjobs/trigger", secure(h.TriggerCronJob))
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
				// Allow Google Fonts for font-src and style-src
				// Allow Monaco Editor from cdn.jsdelivr.net
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
				w.Header().Set("X-Frame-Options", "SAMEORIGIN")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
				// HSTS - only if HTTPS is detected
				if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
					w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
				}
				next.ServeHTTP(fixer, r)
			})
		}

		// Serve static assets (JS, CSS, images, etc.) - Vite puts them in /assets/
		// FileServer points to staticDir/assets because Vite outputs to dist/assets/
		assetsDir := filepath.Join(staticDir, "assets")
		fileServer := http.FileServer(http.Dir(assetsDir))
		mux.Handle("/assets/", secureStaticFiles(http.StripPrefix("/assets/", fileServer)))
		
		// Serve other static files directly from root (favicon, robots.txt, etc.)
		mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, r, filepath.Join(staticDir, "favicon.ico"))
		})
		mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, r, filepath.Join(staticDir, "robots.txt"))
		})
		
		// SPA fallback: serve index.html for all non-API routes
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Don't serve index.html for API routes
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}
			// Don't serve index.html for asset requests
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				http.NotFound(w, r)
				return
			}
			// Add security headers
			// Allow Google Fonts for font-src and style-src
			// Allow Monaco Editor from cdn.jsdelivr.net
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			// HSTS - only if HTTPS is detected
			if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			// Serve index.html for all other routes (SPA routing)
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

		// If no origin (and not OPTIONS), allow if it's not a browser request or handle as same-origin
		// However, for strict security, we might want to block if we expect only browser traffic.
		// For now, we follow the recommendation:
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
			// If no ALLOWED_ORIGINS set, allow same-origin exact match or localhost/127.0.0.1 for dev
			// In production, you should set ALLOWED_ORIGINS
			if origin != "" {
				originURL, err := url.Parse(origin)
				if err == nil {
					host := r.Host
					// Remove port for comparison if present
					if strings.Contains(host, ":") {
						host = strings.Split(host, ":")[0]
					}
					originHost := originURL.Host
					if strings.Contains(originHost, ":") {
						originHost = strings.Split(originHost, ":")[0]
					}
					// Only allow exact match: localhost, 127.0.0.1, or same host
					// Also validate scheme is http/https
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
