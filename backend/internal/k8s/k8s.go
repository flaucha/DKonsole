package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/utils"
)

// Service provides Kubernetes resource operations
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
}

// NewService creates a new Kubernetes service
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
	}
}

// GetNamespaces returns a list of all namespaces
func (s *Service) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []models.Namespace
	for _, ns := range namespaces.Items {
		result = append(result, models.Namespace{
			Name:    ns.Name,
			Status:  string(ns.Status.Phase),
			Labels:  ns.Labels,
			Created: ns.CreationTimestamp.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetResources is implemented in resources.go

// ScaleResource scales a deployment
func (s *Service) ScaleResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	deltaStr := r.URL.Query().Get("delta")

	if kind != "Deployment" {
		http.Error(w, "Scaling supported only for Deployments", http.StatusBadRequest)
		return
	}
	if name == "" {
		http.Error(w, "Missing name", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if namespace != "" {
		if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if namespace == "" {
		namespace = "default"
	}

	delta, err := strconv.Atoi(deltaStr)
	if err != nil || delta == 0 {
		http.Error(w, "Invalid delta", http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()
	scale, err := client.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get scale: %v", err), http.StatusInternalServerError)
		return
	}

	newReplicas := int(scale.Spec.Replicas) + delta
	if newReplicas < 0 {
		newReplicas = 0
	}
	scale.Spec.Replicas = int32(newReplicas)

	if _, err := client.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update scale: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"replicas":%d}`, newReplicas)))
}

// WatchResources is implemented in resource_operations.go

// GetClusterStats returns cluster statistics
func (s *Service) GetClusterStats(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout so cancellation propagates
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()
	
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats := models.ClusterStats{}

	// Create ListOptions with limit to prevent OOM in large clusters
	// Note: For stats we only need counts, so we can use a reasonable limit
	listOpts := metav1.ListOptions{
		Limit: 500, // Limit to prevent OOM, but enough for accurate counts
	}

	// Nodes
	if nodes, err := client.CoreV1().Nodes().List(ctx, listOpts); err == nil {
		stats.Nodes = len(nodes.Items)
	}

	// Namespaces
	if namespaces, err := client.CoreV1().Namespaces().List(ctx, listOpts); err == nil {
		stats.Namespaces = len(namespaces.Items)
	}

	// Pods
	if pods, err := client.CoreV1().Pods("").List(ctx, listOpts); err == nil {
		stats.Pods = len(pods.Items)
	}

	// Deployments
	if deployments, err := client.AppsV1().Deployments("").List(ctx, listOpts); err == nil {
		stats.Deployments = len(deployments.Items)
	}

	// Services
	if services, err := client.CoreV1().Services("").List(ctx, listOpts); err == nil {
		stats.Services = len(services.Items)
	}

	// Ingresses
	if ingresses, err := client.NetworkingV1().Ingresses("").List(ctx, listOpts); err == nil {
		stats.Ingresses = len(ingresses.Items)
	}

	// PVCs
	if pvcs, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, listOpts); err == nil {
		stats.PVCs = len(pvcs.Items)
	}

	// PVs
	if pvs, err := client.CoreV1().PersistentVolumes().List(ctx, listOpts); err == nil {
		stats.PVs = len(pvs.Items)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// TriggerCronJob is implemented in cronjob.go
