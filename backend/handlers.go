package main

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/example/k8s-view/internal/models"
)

// Handlers is kept for backward compatibility with handlers not yet migrated
// New code should use the service modules directly
// This embeds models.Handlers to maintain compatibility
type Handlers struct {
	*models.Handlers
}

// getClient, getDynamicClient, getMetricsClient are kept for backward compatibility
// They are used by Prometheus handlers
func (h *Handlers) getClient(r *http.Request) (*kubernetes.Clientset, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.RLock()
	defer h.RUnlock()
	client, ok := h.Clients[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

func (h *Handlers) getDynamicClient(r *http.Request) (dynamic.Interface, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.RLock()
	defer h.RUnlock()
	client, ok := h.Dynamics[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

func (h *Handlers) getMetricsClient(r *http.Request) *metricsv.Clientset {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.RLock()
	defer h.RUnlock()
	client, ok := h.Metrics[cluster]
	if !ok {
		return nil
	}
	return client
}

// Type aliases for backward compatibility
type ClusterConfig = models.ClusterConfig
type Namespace = models.Namespace
type Resource = models.Resource
type DeploymentDetails = models.DeploymentDetails
type PodMetric = models.PodMetric
type ResourceMeta = models.ResourceMeta

var resourceMeta = models.ResourceMetaMap
var kindAliases = models.KindAliases

// Functions for backward compatibility
func normalizeKind(kind string) string {
	return models.NormalizeKind(kind)
}

func resolveGVR(kind string) (schema.GroupVersionResource, bool) {
	return models.ResolveGVR(kind)
}

// HealthHandler is an unauthenticated liveness endpoint
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
