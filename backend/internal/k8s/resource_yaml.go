package k8s

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// GetResourceYAML returns the YAML representation of a resource
func (s *Service) GetResourceYAML(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetResourceYAML called: kind=%s, name=%s, namespace=%s",
		r.URL.Query().Get("kind"), r.URL.Query().Get("name"), r.URL.Query().Get("namespace"))

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		log.Printf("GetResourceYAML: Error getting dynamic client: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := namespace == "all"

	if kind == "" || name == "" {
		log.Printf("GetResourceYAML: Missing parameters - kind=%s, name=%s", kind, name)
		http.Error(w, "Missing kind or name parameter", http.StatusBadRequest)
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

	if meta.Namespaced && namespace == "" {
		namespace = "default"
	}

	// Try to get GVR from query parameters (passed by API Explorer for CRDs)
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resourceName := r.URL.Query().Get("resource")

	var gvr schema.GroupVersionResource
	if group != "" || version != "" || resourceName != "" {
		// Use explicit GVR from query params (for CRDs)
		gvr = schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resourceName,
		}
	} else {
		// Fall back to static resolution for known types (with alias normalization)
		var ok bool

		// Force discovery for HPA to ensure we get the correct server-preferred version
		if normalizedKind != "HorizontalPodAutoscaler" {
			gvr, ok = models.ResolveGVR(normalizedKind)
		}

		log.Printf("GetResourceYAML: resolveGVR result - ok=%v, gvr=%+v for normalizedKind=%s", ok, gvr, normalizedKind)

		if !ok {
			// Try to find GVR using discovery API
			log.Printf("GetResourceYAML: Resource not in static map, trying discovery API for kind=%s", normalizedKind)
			client, err := s.clusterService.GetClient(r)
			if err != nil {
				log.Printf("GetResourceYAML: Error getting client for discovery: %v", err)
			} else {
				lists, err := client.Discovery().ServerPreferredResources()
				if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
					log.Printf("GetResourceYAML: Discovery error: %v", err)
				} else {
					log.Printf("GetResourceYAML: Discovery returned %d API groups", len(lists))
					// Search for the resource kind in discovered resources
					for _, list := range lists {
						gv, _ := schema.ParseGroupVersion(list.GroupVersion)
						for _, ar := range list.APIResources {
							if ar.Kind == normalizedKind && !strings.Contains(ar.Name, "/") {
								gvr = schema.GroupVersionResource{
									Group:    gv.Group,
									Version:  gv.Version,
									Resource: ar.Name,
								}
								ok = true
								log.Printf("GetResourceYAML: Found GVR via discovery: %+v for kind=%s (namespaced=%v)", gvr, normalizedKind, ar.Namespaced)
								// Update meta if we found it via discovery
								if !meta.Namespaced && ar.Namespaced {
									meta.Namespaced = ar.Namespaced
								}
								break
							}
						}
						if ok {
							break
						}
					}
					if !ok {
						log.Printf("GetResourceYAML: Resource kind=%s not found in discovery API", normalizedKind)
					}
				}
			}

			// If still not found, try common patterns
			if !ok {
				if normalizedKind == "HorizontalPodAutoscaler" {
					// Try v2 first, then v1
					gvr = schema.GroupVersionResource{
						Group:    "autoscaling",
						Version:  "v2",
						Resource: "horizontalpodautoscalers",
					}
				} else {
					// For unknown types, try to infer from kind name
					resourceName := strings.ToLower(normalizedKind) + "s"
					gvr = schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: resourceName,
					}
				}
			}
		}
	}

	// Validate GVR is not empty
	if gvr.Resource == "" {
		log.Printf("GetResourceYAML error: Empty GVR for kind=%s, normalizedKind=%s", kind, normalizedKind)
		http.Error(w, fmt.Sprintf("Unable to resolve resource type: %s", kind), http.StatusBadRequest)
		return
	}

	log.Printf("GetResourceYAML: kind=%s, normalizedKind=%s, gvr=%+v, namespace=%s, name=%s, namespaced=%v",
		kind, normalizedKind, gvr, namespace, name, meta.Namespaced)

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

	obj, err := res.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		log.Printf("GetResourceYAML: First attempt failed - kind=%s, normalizedKind=%s, gvr=%+v, namespace=%s, name=%s, error=%v",
			kind, normalizedKind, gvr, namespace, name, err)

		// If HPA fails, try different API versions
		if normalizedKind == "HorizontalPodAutoscaler" {
			// Try all known HPA versions in order of preference
			versions := []string{"v2", "v2beta2", "v2beta1", "v1"}

			for _, ver := range versions {
				// Skip if it's the version we already tried
				if ver == gvr.Version {
					continue
				}

				log.Printf("GetResourceYAML: Error encountered (%v), trying HPA with version %s", err, ver)
				gvr.Version = ver
				if meta.Namespaced {
					res = dynamicClient.Resource(gvr).Namespace(namespace)
				} else {
					res = dynamicClient.Resource(gvr)
				}
				obj, err = res.Get(ctx, name, metav1.GetOptions{})
				if err == nil {
					log.Printf("GetResourceYAML: Success with version %s", ver)
					break
				}
				log.Printf("GetResourceYAML: Version %s also failed: %v", ver, err)
			}
		}

		if err != nil {
			log.Printf("GetResourceYAML FINAL ERROR: kind=%s, normalizedKind=%s, gvr=%+v, namespace=%s, name=%s, namespaced=%v, error=%v",
				kind, normalizedKind, gvr, namespace, name, meta.Namespaced, err)

			statusCode := http.StatusInternalServerError
			if apierrors.IsNotFound(err) {
				statusCode = http.StatusNotFound
			} else if apierrors.IsForbidden(err) {
				statusCode = http.StatusForbidden
			} else if apierrors.IsBadRequest(err) {
				statusCode = http.StatusBadRequest
			}

			utils.HandleError(w, err, fmt.Sprintf("Failed to fetch resource: %v", err), statusCode)
			return
		}
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

