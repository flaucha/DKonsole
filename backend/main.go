package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

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
		Clients:     make(map[string]*kubernetes.Clientset),
		Dynamics:    make(map[string]dynamic.Interface),
		Metrics:     make(map[string]*metricsv.Clientset),
		RESTConfigs: make(map[string]*rest.Config),
	}
	h.Clients["default"] = clientset
	h.Dynamics["default"] = dynamicClient
	h.RESTConfigs["default"] = config
	if metricsClient != nil {
		h.Metrics["default"] = metricsClient
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", enableCors(h.LoginHandler))
	mux.HandleFunc("/healthz", h.HealthHandler)
	mux.HandleFunc("/api/namespaces", enableCors(AuthMiddleware(h.GetNamespaces)))
	mux.HandleFunc("/api/resources", enableCors(AuthMiddleware(h.GetResources)))
	mux.HandleFunc("/api/resources/watch", enableCors(AuthMiddleware(h.WatchResources)))
	mux.HandleFunc("/api/resource/yaml", enableCors(AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetResourceYAML(w, r)
		} else if r.Method == http.MethodPut {
			h.UpdateResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/api/resource/import", enableCors(AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.ImportResourceYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/api/apis", enableCors(AuthMiddleware(h.ListAPIResources)))
	mux.HandleFunc("/api/apis/resources", enableCors(AuthMiddleware(h.ListAPIResourceObjects)))
	mux.HandleFunc("/api/apis/yaml", enableCors(AuthMiddleware(h.GetAPIResourceYAML)))
	mux.HandleFunc("/api/scale", enableCors(AuthMiddleware(h.ScaleResource)))
	mux.HandleFunc("/api/overview", enableCors(AuthMiddleware(h.GetClusterStats)))
	mux.HandleFunc("/api/pods/logs", enableCors(AuthMiddleware(h.StreamPodLogs)))
	mux.HandleFunc("/api/pods/exec", enableCors(AuthMiddleware(h.ExecIntoPod))) // WebSocket - now protected
	mux.HandleFunc("/api/clusters", enableCors(AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetClusters(w, r)
		} else if r.Method == http.MethodPost {
			h.AddCluster(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/api/resource", enableCors(AuthMiddleware(h.DeleteResource)))
	mux.HandleFunc("/api/logo", enableCors(AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetLogo(w, r)
		} else if r.Method == http.MethodPost {
			h.UploadLogo(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
