package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/utils"
)

const (
	// DefaultListLimit is the maximum number of items to fetch in a single list operation
	// This prevents OOM issues in large clusters
	DefaultListLimit = int64(500)
)

// Service provides API resource and CRD operations
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
}

// NewService creates a new API service
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
	}
}

// APIResourceInfo represents information about an API resource
type APIResourceInfo struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Resource   string `json:"resource"`
	Kind       string `json:"kind"`
	Namespaced bool   `json:"namespaced"`
}

// APIResourceObject represents an instance of an API resource
type APIResourceObject struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status,omitempty"`
	Created   string      `json:"created,omitempty"`
	Raw       interface{} `json:"raw,omitempty"`
}

// ListAPIResources lists all available API resources in the cluster
func (s *Service) ListAPIResources(w http.ResponseWriter, r *http.Request) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lists, err := client.Discovery().ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		http.Error(w, fmt.Sprintf("Failed to discover APIs: %v", err), http.StatusInternalServerError)
		return
	}

	var result []APIResourceInfo
	for _, l := range lists {
		gv, _ := schema.ParseGroupVersion(l.GroupVersion)
		for _, ar := range l.APIResources {
			if strings.Contains(ar.Name, "/") { // skip subresources
				continue
			}
			result = append(result, APIResourceInfo{
				Group:      gv.Group,
				Version:    gv.Version,
				Resource:   ar.Name,
				Kind:       ar.Kind,
				Namespaced: ar.Namespaced,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListAPIResourceObjects lists instances of a specific API resource
func (s *Service) ListAPIResourceObjects(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout so cancellation propagates to Kubernetes API calls
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()
	
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")
	namespacedParam := r.URL.Query().Get("namespaced") == "true"

	if namespacedParam && namespace == "" {
		namespace = "default"
	}

	if resource == "" || version == "" {
		http.Error(w, "Missing resource or version", http.StatusBadRequest)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	var res dynamic.ResourceInterface
	// Cluster-scoped resources or namespaced resources across all namespaces
	if !namespacedParam || namespace == "all" || namespace == "" {
		res = dynamicClient.Resource(gvr)
	} else {
		res = dynamicClient.Resource(gvr).Namespace(namespace)
	}

	// Create ListOptions with limit to prevent OOM in large clusters
	listOpts := metav1.ListOptions{
		Limit: DefaultListLimit,
	}

	list, err := res.List(ctx, listOpts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list resource: %v", err), http.StatusInternalServerError)
		return
	}

	var objects []APIResourceObject
	for _, item := range list.Items {
		obj := APIResourceObject{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Kind:      item.GetKind(),
			Created:   item.GetCreationTimestamp().Format(time.RFC3339),
		}
		if status, ok := item.Object["status"]; ok {
			if statusMap, ok := status.(map[string]interface{}); ok {
				if phase, ok := statusMap["phase"].(string); ok {
					obj.Status = phase
				} else if conditions, ok := statusMap["conditions"].([]interface{}); ok && len(conditions) > 0 {
					last := conditions[len(conditions)-1]
					if cond, ok := last.(map[string]interface{}); ok {
						if t, ok := cond["type"].(string); ok {
							obj.Status = t
							if s, ok := cond["status"].(string); ok && s != "" {
								obj.Status += fmt.Sprintf(" (%s)", s)
							}
						}
					}
				}
			}
		}
		obj.Raw = item.Object
		objects = append(objects, obj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(objects)
}

// GetAPIResourceYAML returns the YAML representation of an API resource
func (s *Service) GetAPIResourceYAML(w http.ResponseWriter, r *http.Request) {
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	namespaced := r.URL.Query().Get("namespaced") == "true"

	if resource == "" || version == "" || name == "" {
		http.Error(w, "Missing resource, version, or name", http.StatusBadRequest)
		return
	}

	if namespaced && namespace == "" {
		namespace = "default"
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	var res dynamic.ResourceInterface
	if namespaced {
		res = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		res = dynamicClient.Resource(gvr)
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	obj, err := res.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		utils.HandleError(w, err, fmt.Sprintf("Failed to fetch resource"), http.StatusInternalServerError)
		return
	}

	jsonData, err := obj.MarshalJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal resource: %v", err), http.StatusInternalServerError)
		return
	}

	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode YAML: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write(yamlData)
}

// CRD represents a Custom Resource Definition
type CRD struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Scope   string `json:"scope"`
}

// GetCRDs returns a list of all Custom Resource Definitions in the cluster
func (s *Service) GetCRDs(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()
	
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, "Dynamic client not found", http.StatusInternalServerError)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	// Create ListOptions with limit to prevent OOM in large clusters
	listOpts := metav1.ListOptions{
		Limit: DefaultListLimit,
	}

	unstructuredList, err := dynamicClient.Resource(gvr).List(ctx, listOpts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	crds := []CRD{}
	for _, item := range unstructuredList.Items {
		spec, _ := item.Object["spec"].(map[string]interface{})
		group, _ := spec["group"].(string)
		names, _ := spec["names"].(map[string]interface{})
		kind, _ := names["kind"].(string)
		scope, _ := spec["scope"].(string)

		// Get versions
		versions, _ := spec["versions"].([]interface{})
		for _, v := range versions {
			vMap, _ := v.(map[string]interface{})
			version, _ := vMap["name"].(string)
			served, _ := vMap["served"].(bool)
			if served {
				crds = append(crds, CRD{
					Name:    item.GetName(),
					Group:   group,
					Version: version,
					Kind:    kind,
					Scope:   scope,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(crds)
}

// CRInstance represents an instance of a Custom Resource
type CRInstance struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace,omitempty"`
	Created   string                 `json:"created"`
	Raw       map[string]interface{} `json:"raw"`
}

// GetCRDResources returns instances of a specific CRD
func (s *Service) GetCRDResources(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")
	namespaced := r.URL.Query().Get("namespaced") == "true"

	if resource == "" || version == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, "Dynamic client not found", http.StatusInternalServerError)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	var unstructuredList *unstructured.UnstructuredList

	// Use request context with timeout
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()
	
	// Create ListOptions with limit to prevent OOM in large clusters
	listOpts := metav1.ListOptions{
		Limit: DefaultListLimit,
	}
	
	if namespaced && namespace != "" {
		unstructuredList, err = dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, listOpts)
	} else {
		unstructuredList, err = dynamicClient.Resource(gvr).List(ctx, listOpts)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	instances := []CRInstance{}
	for _, item := range unstructuredList.Items {
		instances = append(instances, CRInstance{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Created:   item.GetCreationTimestamp().Format(time.RFC3339),
			Raw:       item.Object,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

// GetCRDYaml returns the YAML representation of a CRD instance
func (s *Service) GetCRDYaml(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()
	
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	namespaced := r.URL.Query().Get("namespaced") == "true"

	if resource == "" || version == "" || name == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Add validation for name and namespace
	if err := utils.ValidateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Only validate namespace if it's provided and relevant
	if namespaced && namespace != "" {
		if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, "Dynamic client not found", http.StatusInternalServerError)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	var obj *unstructured.Unstructured

	if namespaced && namespace != "" {
		obj, err = dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write(yamlBytes)
}
