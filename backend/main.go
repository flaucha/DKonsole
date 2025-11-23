package main

import (
	"flag"
	"fmt"
	"log"
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

	mux.HandleFunc("/api/login", public(h.LoginHandler))
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
