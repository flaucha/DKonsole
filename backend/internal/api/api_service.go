package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
)

// APIService provides business logic for API resource operations
type APIService struct {
	discoveryRepo APIResourceRepository
	resourceRepo  DynamicResourceRepository
}

// NewAPIService creates a new APIService
func NewAPIService(discoveryRepo APIResourceRepository, resourceRepo DynamicResourceRepository) *APIService {
	return &APIService{
		discoveryRepo: discoveryRepo,
		resourceRepo:  resourceRepo,
	}
}

// ListAPIResources returns all available API resources
func (s *APIService) ListAPIResources(ctx context.Context) ([]APIResourceInfo, error) {
	if s.discoveryRepo == nil {
		return nil, fmt.Errorf("discovery repository not set")
	}
	lists, err := s.discoveryRepo.ListAPIResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list API resources: %w", err)
	}
	var result []APIResourceInfo
	for _, l := range lists {
		gv, _ := schema.ParseGroupVersion(l.GroupVersion)
		for _, ar := range l.APIResources {
			if strings.Contains(ar.Name, "/") {
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
	return result, nil
}

// ListAPIResourceObjectsRequest represents parameters for listing API resource objects
type ListAPIResourceObjectsRequest struct {
	Group         string
	Version       string
	Resource      string
	Namespace     string
	Namespaced    bool
	Limit         int64
	ContinueToken string
}

// ListAPIResourceObjectsResponse represents the paginated response
type ListAPIResourceObjectsResponse struct {
	Resources []APIResourceObject `json:"resources"`
	Continue  string              `json:"continue,omitempty"`
}

// ListAPIResourceObjects lists instances of a specific API resource
// It filters resources based on user permissions before returning them
func (s *APIService) ListAPIResourceObjects(ctx context.Context, req ListAPIResourceObjectsRequest) (*ListAPIResourceObjectsResponse, error) {
	if s.resourceRepo == nil {
		return nil, fmt.Errorf("resource repository not set")
	}

	// Validate namespace access if resource is namespaced
	if req.Namespaced && req.Namespace != "" {
		hasAccess, err := permissions.HasNamespaceAccess(ctx, req.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to check namespace access: %w", err)
		}
		if !hasAccess {
			return nil, fmt.Errorf("access denied to namespace: %s", req.Namespace)
		}
	}

	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}
	list, continueToken, err := s.resourceRepo.List(ctx, gvr, req.Namespace, req.Namespaced, req.Limit, req.ContinueToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	var objects []APIResourceObject
	for _, item := range list.Items {
		// Filter by namespace permissions
		itemNamespace := item.GetNamespace()
		if req.Namespaced && itemNamespace != "" {
			hasAccess, err := permissions.HasNamespaceAccess(ctx, itemNamespace)
			if err != nil {
				continue // Skip if we can't check permissions
			}
			if !hasAccess {
				continue // Skip resources in namespaces the user doesn't have access to
			}
		}

		obj := APIResourceObject{
			Name:      item.GetName(),
			Namespace: itemNamespace,
			Kind:      item.GetKind(),
			Created:   item.GetCreationTimestamp().Format(time.RFC3339),
		}
		if status, ok := item.Object["status"].(map[string]interface{}); ok {
			if phase, ok := status["phase"].(string); ok {
				obj.Status = phase
			} else if conditions, ok := status["conditions"].([]interface{}); ok && len(conditions) > 0 {
				if last, ok := conditions[len(conditions)-1].(map[string]interface{}); ok {
					if t, ok := last["type"].(string); ok {
						obj.Status = t
					}
				}
			}
		}
		obj.Raw = item.Object
		objects = append(objects, obj)
	}
	return &ListAPIResourceObjectsResponse{
		Resources: objects,
		Continue:  continueToken,
	}, nil
}

// GetResourceYAMLRequest represents parameters for getting resource YAML
type GetResourceYAMLRequest struct {
	Group      string
	Version    string
	Resource   string
	Name       string
	Namespace  string
	Namespaced bool
}

// GetResourceYAML returns the YAML representation of a resource
// It validates namespace permissions before retrieving.
func (s *APIService) GetResourceYAML(ctx context.Context, req GetResourceYAMLRequest) (string, error) {
	if s.resourceRepo == nil {
		return "", fmt.Errorf("resource repository not set")
	}

	// Validate namespace access if resource is namespaced
	if req.Namespaced && req.Namespace != "" {
		hasAccess, err := permissions.HasNamespaceAccess(ctx, req.Namespace)
		if err != nil {
			return "", fmt.Errorf("failed to check namespace access: %w", err)
		}
		if !hasAccess {
			return "", fmt.Errorf("access denied to namespace: %s", req.Namespace)
		}
	}

	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}
	obj, err := s.resourceRepo.Get(ctx, gvr, req.Name, req.Namespace, req.Namespaced)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %w", err)
	}
	jsonData, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource: %w", err)
	}
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to convert to YAML: %w", err)
	}
	return string(yamlData), nil
}

// CRDService provides business logic for CRD operations
type CRDService struct {
	repo CRDRepository
}

// NewCRDService creates a new CRDService
func NewCRDService(repo CRDRepository) *CRDService {
	return &CRDService{repo: repo}
}

// CRD represents a Custom Resource Definition
type CRD struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Scope   string `json:"scope"`
}

// GetCRDsRequest represents parameters for getting CRDs
type GetCRDsRequest struct {
	Limit         int64
	ContinueToken string
}

// GetCRDsResponse represents the paginated response for CRDs
type GetCRDsResponse struct {
	CRDs     []CRD  `json:"crds"`
	Continue string `json:"continue,omitempty"`
}

// GetCRDs returns all Custom Resource Definitions
func (s *CRDService) GetCRDs(ctx context.Context, req GetCRDsRequest) (*GetCRDsResponse, error) {
	crdList, err := s.repo.ListCRDs(ctx, req.Limit, req.ContinueToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}
	var crds []CRD
	for _, item := range crdList.Items {
		spec, ok := item.Object["spec"].(map[string]interface{})
		if !ok {
			continue
		}
		group, _ := spec["group"].(string)
		names, _ := spec["names"].(map[string]interface{})
		kind, _ := names["kind"].(string)
		scope, _ := spec["scope"].(string)
		versions, _ := spec["versions"].([]interface{})
		for _, v := range versions {
			if version, ok := v.(map[string]interface{}); ok {
				versionName, _ := version["name"].(string)
				served, _ := version["served"].(bool)
				if served && group != "" && kind != "" {
					crds = append(crds, CRD{
						Name:    item.GetName(),
						Group:   group,
						Version: versionName,
						Kind:    kind,
						Scope:   scope,
					})
				}
			}
		}
	}
	return &GetCRDsResponse{
		CRDs:     crds,
		Continue: crdList.GetContinue(),
	}, nil
}

// CRInstance represents an instance of a Custom Resource
type CRInstance struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace,omitempty"`
	Created   string                 `json:"created"`
	Raw       map[string]interface{} `json:"raw"`
}

// GetCRDResourcesRequest represents parameters for listing CRD instances
type GetCRDResourcesRequest struct {
	Group         string
	Version       string
	Resource      string
	Namespace     string
	Namespaced    bool
	Limit         int64
	ContinueToken string
}

// GetCRDResourcesResponse represents the paginated response for CRD instances
type GetCRDResourcesResponse struct {
	Resources []CRInstance `json:"resources"`
	Continue  string       `json:"continue,omitempty"`
}

// GetCRDResources lists instances of a specific Custom Resource Definition
func (s *APIService) GetCRDResources(ctx context.Context, req GetCRDResourcesRequest) (*GetCRDResourcesResponse, error) {
	if s.resourceRepo == nil {
		return nil, fmt.Errorf("resource repository not set")
	}
	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}
	list, continueToken, err := s.resourceRepo.List(ctx, gvr, req.Namespace, req.Namespaced, req.Limit, req.ContinueToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list CRD resources: %w", err)
	}
	var instances []CRInstance
	for _, item := range list.Items {
		instances = append(instances, CRInstance{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Created:   item.GetCreationTimestamp().Format(time.RFC3339),
			Raw:       item.Object,
		})
	}
	return &GetCRDResourcesResponse{
		Resources: instances,
		Continue:  continueToken,
	}, nil
}
