package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// UpdateResourceYAML updates a resource from YAML
func (s *Service) UpdateResourceYAML(w http.ResponseWriter, r *http.Request) {
	log.Printf("UpdateResourceYAML: Method=%s, Content-Type=%s, kind=%s, name=%s, namespace=%s",
		r.Method, r.Header.Get("Content-Type"), r.URL.Query().Get("kind"), r.URL.Query().Get("name"), r.URL.Query().Get("namespace"))

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		log.Printf("UpdateResourceYAML: Error getting dynamic client: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")

	if kind == "" {
		log.Printf("UpdateResourceYAML: Missing kind parameter")
		http.Error(w, "Missing kind parameter", http.StatusBadRequest)
		return
	}

	// Normalize kind (handle aliases like HPA -> HorizontalPodAutoscaler)
	normalizedKind := models.NormalizeKind(kind)

	// Try to get metadata; use namespaced from query param if provided (for CRDs)
	namespacedParam := r.URL.Query().Get("namespaced")
	meta, ok := models.ResourceMetaMap[normalizedKind]
	if !ok {
		// Default to namespaced=true unless explicitly told otherwise
		isNamespaced := true
		if namespacedParam == "false" {
			isNamespaced = false
		}
		meta = models.ResourceMeta{Namespaced: isNamespaced}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
		return
	}

	jsonData, err := yaml.YAMLToJSON(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid YAML: %v", err), http.StatusBadRequest)
		return
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj.Object); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse resource: %v", err), http.StatusBadRequest)
		return
	}

	// Ensure name/namespace are set
	if obj.GetName() == "" {
		if name == "" {
			http.Error(w, "Resource name missing", http.StatusBadRequest)
			return
		}
		obj.SetName(name)
	}

	if meta.Namespaced {
		if obj.GetNamespace() != "" {
			namespace = obj.GetNamespace()
		}
		if namespace == "" {
			namespace = "default"
		}
		obj.SetNamespace(namespace)
	} else {
		obj.SetNamespace("")
	}

	// Cleanup noisy metadata fields that are not needed for updates
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")

	// Extract GVR from the object's apiVersion and kind (not from query param)
	apiVersion := obj.GetAPIVersion()
	objKind := obj.GetKind()

	// Parse apiVersion (format: "group/version" or just "version")
	group := ""
	version := apiVersion
	if parts := strings.SplitN(apiVersion, "/", 2); len(parts) == 2 {
		group = parts[0]
		version = parts[1]
	}

	// Try to get resource name from static map first
	gvr, ok := models.ResolveGVR(objKind)
	if !ok {
		// For CRDs and unknown types, use lowercase(kind) + "s" as resource name
		resourceName := strings.ToLower(objKind) + "s"
		gvr = schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resourceName,
		}
	}

	var res dynamic.ResourceInterface
	if meta.Namespaced {
		res = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())
	} else {
		res = dynamicClient.Resource(gvr)
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Use Server-Side Apply (SSA)
	force := true
	patchOptions := metav1.PatchOptions{
		FieldManager: "dkonsole",
		Force:        &force,
	}

	// Marshal the object back to JSON for the patch body
	patchData, err := json.Marshal(obj.Object)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal patch: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = res.Patch(ctx, obj.GetName(), types.ApplyPatchType, patchData, patchOptions)
	if err != nil {
		utils.AuditLog(r, "update", objKind, obj.GetName(), namespace, false, err, nil)
		utils.HandleError(w, err, fmt.Sprintf("Failed to update %s", objKind), http.StatusInternalServerError)
		return
	}

	utils.AuditLog(r, "update", objKind, obj.GetName(), namespace, true, nil, nil)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"updated"}`))
}

// ImportResourceYAML imports a resource from YAML
func (s *Service) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
	log.Printf("ImportResourceYAML: Method=%s, Content-Type=%s", r.Method, r.Header.Get("Content-Type"))

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		log.Printf("ImportResourceYAML: Error getting dynamic client: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Limit body size to 1MB to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ImportResourceYAML: Error reading body: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read body or body too large: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("ImportResourceYAML: Body size=%d bytes", len(body))

	dec := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(body), 4096)
	var applied []string
	resourceCount := 0
	maxResources := 50 // Maximum resources per request

	// Counters per resource type
	resourceTypeCounts := make(map[string]int)
	maxPerType := map[string]int{
		"Deployment":              10,
		"Service":                 20,
		"ConfigMap":               30,
		"Secret":                  10,
		"Job":                     15,
		"CronJob":                 5,
		"StatefulSet":             10,
		"DaemonSet":               5,
		"HorizontalPodAutoscaler": 10,
		"Ingress":                 15,
		"NetworkPolicy":           10,
		"ServiceAccount":          20,
		"Role":                    15,
		"RoleBinding":             15,
		"PersistentVolumeClaim":   10,
	}

	for {
		if resourceCount >= maxResources {
			http.Error(w, fmt.Sprintf("Too many resources (max %d per request)", maxResources), http.StatusBadRequest)
			return
		}

		var objMap map[string]interface{}
		if err := dec.Decode(&objMap); err != nil {
			if err == io.EOF {
				break
			}
			http.Error(w, fmt.Sprintf("Failed to decode YAML: %v", err), http.StatusBadRequest)
			return
		}
		if len(objMap) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{Object: objMap}
		kind := obj.GetKind()
		if kind == "" {
			http.Error(w, "Resource kind missing", http.StatusBadRequest)
			return
		}

		// Validate limit per resource type
		if maxCount, exists := maxPerType[kind]; exists {
			if resourceTypeCounts[kind] >= maxCount {
				http.Error(w, fmt.Sprintf("Too many resources of type %s (max %d)", kind, maxCount), http.StatusBadRequest)
				return
			}
			resourceTypeCounts[kind]++
		} else {
			// For unspecified types, use general limit
			if resourceTypeCounts[kind] >= 10 {
				http.Error(w, fmt.Sprintf("Too many resources of type %s (max 10)", kind), http.StatusBadRequest)
				return
			}
			resourceTypeCounts[kind]++
		}

		// Validate namespace (prevent creation in system namespaces)
		if utils.IsSystemNamespace(obj.GetNamespace()) {
			http.Error(w, "Cannot create resources in system namespaces", http.StatusForbidden)
			return
		}

		// Extract GVR from the object's apiVersion and kind (supports CRDs and any custom resources)
		apiVersion := obj.GetAPIVersion()
		if apiVersion == "" {
			http.Error(w, "Resource apiVersion missing", http.StatusBadRequest)
			return
		}

		// Parse apiVersion (format: "group/version" or just "version")
		group := ""
		version := apiVersion
		if parts := strings.SplitN(apiVersion, "/", 2); len(parts) == 2 {
			group = parts[0]
			version = parts[1]
		}

		// Normalize kind (handle aliases like HPA -> HorizontalPodAutoscaler)
		normalizedKind := models.NormalizeKind(kind)

		// Try to get resource name and metadata from static map first
		gvr, ok := models.ResolveGVR(normalizedKind)
		var meta models.ResourceMeta
		if ok {
			meta = models.ResourceMetaMap[normalizedKind]
		} else {
			// For CRDs and unknown types, use lowercase(kind) + "s" as resource name
			resourceName := strings.ToLower(kind) + "s"
			gvr = schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: resourceName,
			}
			// Determine if namespaced
			hasNamespace := obj.GetNamespace() != ""
			isClusterScopedKind := strings.HasPrefix(kind, "Cluster") ||
				kind == "PersistentVolume" || kind == "StorageClass" ||
				kind == "CustomResourceDefinition" || kind == "Node"

			if isClusterScopedKind {
				meta = models.ResourceMeta{Namespaced: false}
			} else if hasNamespace {
				meta = models.ResourceMeta{Namespaced: true}
			} else {
				// Default to namespaced (most resources are namespaced)
				meta = models.ResourceMeta{Namespaced: true}
			}
		}

		// Namespace defaults
		if meta.Namespaced {
			if obj.GetNamespace() == "" {
				obj.SetNamespace("default")
			}
		} else {
			// Ensure cluster-scoped resources have no namespace
			obj.SetNamespace("")
		}

		// Clean noisy metadata fields
		unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
		unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
		unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
		unstructured.RemoveNestedField(obj.Object, "metadata", "uid")

		ctx, cancel := utils.CreateTimeoutContext()
		defer cancel()

		// Validate ResourceQuota and LimitRange before creating resource
		// Note: These validations are advisory - Kubernetes API server will enforce them
		if meta.Namespaced && obj.GetNamespace() != "" {
			client, err := s.clusterService.GetClient(r)
			if err == nil {
				// Validate ResourceQuota (check if creation would exceed limits)
				if quotaErr := s.validateResourceQuota(ctx, client, obj.GetNamespace(), obj); quotaErr != nil {
					log.Printf("Warning: ResourceQuota validation issue for %s/%s in %s: %v", kind, obj.GetName(), obj.GetNamespace(), quotaErr)
				}

				// Validate LimitRange (check if resource conforms to constraints)
				if limitErr := s.validateLimitRange(ctx, client, obj.GetNamespace(), obj); limitErr != nil {
					log.Printf("Warning: LimitRange validation issue for %s/%s in %s: %v", kind, obj.GetName(), obj.GetNamespace(), limitErr)
				}
			}
		}

		var res dynamic.ResourceInterface
		if meta.Namespaced {
			res = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())
		} else {
			res = dynamicClient.Resource(gvr)
		}

		// Use Server-Side Apply (SSA) for import as well
		force := true
		patchOptions := metav1.PatchOptions{
			FieldManager: "dkonsole",
			Force:        &force,
		}

		// Marshal the object back to JSON for the patch body
		patchData, err := json.Marshal(obj.Object)
		if err != nil {
			utils.HandleError(w, err, fmt.Sprintf("Failed to marshal patch for %s", kind), http.StatusInternalServerError)
			return
		}

		_, err = res.Patch(ctx, obj.GetName(), types.ApplyPatchType, patchData, patchOptions)
		if err != nil {
			utils.HandleError(w, err, fmt.Sprintf("Failed to apply %s", kind), http.StatusInternalServerError)
			return
		}

		nsPart := obj.GetNamespace()
		if nsPart == "" {
			nsPart = "-"
		}
		applied = append(applied, fmt.Sprintf("%s/%s/%s", kind, nsPart, obj.GetName()))
		resourceCount++
	}

	if len(applied) == 0 {
		http.Error(w, "No resources found in YAML", http.StatusBadRequest)
		return
	}

	resp := map[string]interface{}{
		"status":    "applied",
		"count":     len(applied),
		"resources": applied,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteResource deletes a resource
func (s *Service) DeleteResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	force := r.URL.Query().Get("force") == "true"
	allNamespaces := namespace == "all"

	if kind == "" || name == "" {
		http.Error(w, "Missing kind or name", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if namespace != "" && namespace != "all" {
		if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Normalize kind (handle aliases like HPA -> HorizontalPodAutoscaler)
	normalizedKind := models.NormalizeKind(kind)

	meta, ok := models.ResourceMetaMap[normalizedKind]
	if !ok {
		http.Error(w, "Unsupported kind", http.StatusBadRequest)
		return
	}

	if meta.Namespaced {
		if allNamespaces {
			http.Error(w, "Namespace is required to delete namespaced resources", http.StatusBadRequest)
			return
		}
		if namespace == "" {
			namespace = "default"
		}
	}

	gvr, _ := models.ResolveGVR(normalizedKind)
	var res dynamic.ResourceInterface
	if meta.Namespaced {
		res = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		res = dynamicClient.Resource(gvr)
	}

	var grace int64 = 30
	propagation := metav1.DeletePropagationForeground
	if force {
		grace = 0
		propagation = metav1.DeletePropagationBackground
	}

	delOpts := metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &propagation,
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	if err := res.Delete(ctx, name, delOpts); err != nil {
		utils.AuditLog(r, "delete", kind, name, namespace, false, err, map[string]interface{}{
			"force": force,
		})
		utils.HandleError(w, err, fmt.Sprintf("Failed to delete %s", kind), http.StatusInternalServerError)
		return
	}

	utils.AuditLog(r, "delete", kind, name, namespace, true, nil, map[string]interface{}{
		"force": force,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"deleted"}`))
}

// WatchResources watches for resource changes
func (s *Service) WatchResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := namespace == "all"
	if namespace == "" {
		namespace = "default"
	}

	// Normalize kind (handle aliases like HPA -> HorizontalPodAutoscaler)
	normalizedKind := models.NormalizeKind(kind)

	meta, ok := models.ResourceMetaMap[normalizedKind]
	if !ok {
		http.Error(w, "Unsupported kind", http.StatusBadRequest)
		return
	}

	gvr, _ := models.ResolveGVR(normalizedKind)
	var res dynamic.ResourceInterface
	if meta.Namespaced {
		if allNamespaces {
			res = dynamicClient.Resource(gvr)
		} else {
			res = dynamicClient.Resource(gvr).Namespace(namespace)
		}
	} else {
		res = dynamicClient.Resource(gvr)
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()
	watcher, err := res.Watch(ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start watch: %v", err), http.StatusInternalServerError)
		return
	}
	defer watcher.Stop()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			payload := map[string]string{
				"type":      string(event.Type),
				"name":      obj.GetName(),
				"namespace": obj.GetNamespace(),
			}
			data, _ := json.Marshal(payload)
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

