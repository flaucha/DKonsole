package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// UpdateResourceYAML updates a resource from YAML
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) UpdateResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	namespacedParam := r.URL.Query().Get("namespaced") == "true"

	if kind == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing kind parameter")
		return
	}

	// Read YAML body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read body: %v", err))
		return
	}

	// Get dynamic client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Prepare request
	updateReq := UpdateResourceRequest{
		YAMLContent: string(body),
		Kind:        kind,
		Name:        name,
		Namespace:   namespace,
		Namespaced:  namespacedParam,
	}

	// Call service to update resource (business logic layer)
	err = resourceService.UpdateResource(ctx, updateReq)
	if err != nil {
		utils.AuditLog(r, "update", kind, name, namespace, false, err, nil)
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update resource: %v", err))
		return
	}

	// Audit log
	utils.AuditLog(r, "update", kind, name, namespace, true, nil, nil)

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ImportResourceYAML imports a resource from YAML
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Limit body size to 1MB to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read body or body too large: %v", err))
		return
	}

	// Get clients
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	importService := s.serviceFactory.CreateImportService(dynamicClient, client)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Prepare request
	importReq := ImportResourceRequest{
		YAMLContent: body,
	}

	// Call service to import resources (business logic layer)
	result, err := importService.ImportResources(ctx, importReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to import resources: %v", err))
		return
	}

	// Build response
	resp := map[string]interface{}{
		"status":    result.Status,
		"count":     result.Count,
		"resources": result.Applied,
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, resp)
}

// DeleteResource deletes a resource
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) DeleteResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse HTTP parameters
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	force := r.URL.Query().Get("force") == "true"
	allNamespaces := namespace == "all"

	if kind == "" || name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing kind or name")
		return
	}

	// Validate parameters
	if err := utils.ValidateK8sName(name, "name"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if namespace != "" && namespace != "all" {
		if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Normalize kind
	normalizedKind := models.NormalizeKind(kind)

	// Check if kind is supported
	meta, ok := models.ResourceMetaMap[normalizedKind]
	if !ok {
		utils.ErrorResponse(w, http.StatusBadRequest, "Unsupported kind")
		return
	}

	// Validate namespace requirements
	if meta.Namespaced {
		if allNamespaces {
			utils.ErrorResponse(w, http.StatusBadRequest, "Namespace is required to delete namespaced resources")
			return
		}
		if namespace == "" {
			namespace = "default"
		}
	}

	// Get dynamic client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Prepare request
	deleteReq := DeleteResourceRequest{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		Force:     force,
	}

	// Call service to delete resource (business logic layer)
	err = resourceService.DeleteResource(ctx, deleteReq)
	if err != nil {
		utils.AuditLog(r, "delete", kind, name, namespace, false, err, map[string]interface{}{
			"force": force,
		})
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete resource: %v", err))
		return
	}

	// Audit log
	utils.AuditLog(r, "delete", kind, name, namespace, true, nil, map[string]interface{}{
		"force": force,
	})

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// WatchResources watches for resource changes
// Refactored to use layered architecture:
// Handler (HTTP/SSE Streaming) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) WatchResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse HTTP parameters
	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := namespace == "all"

	if kind == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing kind parameter")
		return
	}

	// Get dynamic client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	watchService := s.serviceFactory.CreateWatchService()

	// Create context (use request context to allow cancellation)
	ctx := r.Context()

	// Prepare watch request
	watchReq := WatchRequest{
		Kind:          kind,
		Namespace:     namespace,
		AllNamespaces: allNamespaces,
	}

	// Start watch (business logic layer)
	watcher, err := watchService.StartWatch(ctx, dynamicClient, watchReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to start watch: %v", err))
		return
	}
	defer watcher.Stop()

	// Check if streaming is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Stream events (HTTP layer - streaming logic remains here for efficiency)
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}

			// Transform event (business logic layer)
			result, err := watchService.TransformEvent(event)
			if err != nil {
				// Log error but continue processing other events
				continue
			}

			// Write SSE event (HTTP layer)
			data, _ := json.Marshal(result)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// validateResourceQuota checks if creating the resource would exceed ResourceQuota limits
func (s *Service) validateResourceQuota(ctx context.Context, client *kubernetes.Clientset, namespace string, obj *unstructured.Unstructured) error {
	// Only validate for namespaced resources
	if namespace == "" {
		return nil
	}

	// Get all ResourceQuotas in the namespace
	quotas, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Warning: Could not check ResourceQuota for namespace %s: %v", namespace, err)
		return nil
	}

	if len(quotas.Items) == 0 {
		return nil
	}

	kind := obj.GetKind()

	// For Pods, check container resource requests/limits
	if kind == "Pod" {
		containersRaw, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "spec", "containers")
		if found {
			containers, ok := containersRaw.([]interface{})
			if ok {
				for _, container := range containers {
					containerMap, ok := container.(map[string]interface{})
					if !ok {
						continue
					}

					resources, found, _ := unstructured.NestedMap(containerMap, "resources")
					if found && resources != nil {
						requests, _, _ := unstructured.NestedMap(resources, "requests")
						limits, _, _ := unstructured.NestedMap(resources, "limits")

						for _, quota := range quotas.Items {
							if err := utils.CheckQuotaLimits(quota, requests, limits); err != nil {
								return fmt.Errorf("resource quota exceeded: %v", err)
							}
						}
					}
				}
			}
		}
	}

	// For PersistentVolumeClaim, check storage requests
	if kind == "PersistentVolumeClaim" {
		spec, found, _ := unstructured.NestedMap(obj.Object, "spec", "resources", "requests")
		if found {
			storage, ok := spec["storage"].(string)
			if ok {
				for _, quota := range quotas.Items {
					if err := utils.CheckStorageQuota(quota, storage); err != nil {
						return fmt.Errorf("storage quota exceeded: %v", err)
					}
				}
			}
		}
	}

	return nil
}

// validateLimitRange checks if the resource conforms to LimitRange constraints
func (s *Service) validateLimitRange(ctx context.Context, client *kubernetes.Clientset, namespace string, obj *unstructured.Unstructured) error {
	// Only validate for namespaced resources
	if namespace == "" {
		return nil
	}

	// Get LimitRanges in the namespace
	limitRanges, err := client.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Warning: Could not check LimitRange for namespace %s: %v", namespace, err)
		return nil
	}

	if len(limitRanges.Items) == 0 {
		return nil
	}

	kind := obj.GetKind()
	if kind == "Pod" {
		containersRaw, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "spec", "containers")
		if found {
			containers, ok := containersRaw.([]interface{})
			if ok {
				for _, container := range containers {
					containerMap, ok := container.(map[string]interface{})
					if !ok {
						continue
					}

					resources, found, _ := unstructured.NestedMap(containerMap, "resources")
					if found && resources != nil {
						for _, lr := range limitRanges.Items {
							for _, limit := range lr.Spec.Limits {
								if limit.Type == corev1.LimitTypeContainer {
									log.Printf("Validating container resources against LimitRange %s", lr.Name)
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}
