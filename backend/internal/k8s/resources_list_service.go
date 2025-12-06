package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
)

// ResourceListService provides business logic for listing Kubernetes resources
type ResourceListService struct {
	clusterService *cluster.Service
	prometheusURL  string // URL for Prometheus queries (optional)
}

// NewResourceListService creates a new ResourceListService
func NewResourceListService(clusterService *cluster.Service, prometheusURL string) *ResourceListService {
	return &ResourceListService{
		clusterService: clusterService,
		prometheusURL:  prometheusURL,
	}
}

// ListResourcesRequest represents parameters for listing resources
type ListResourcesRequest struct {
	Kind          string
	Namespace     string
	AllNamespaces bool
	LabelSelector string
	Client        kubernetes.Interface
	MetricsClient *metricsv.Clientset
}

// isClusterScoped returns true if the resource kind is not namespaced
func isClusterScoped(kind string) bool {
	switch kind {
	case "Node", "ClusterRole", "ClusterRoleBinding", "PersistentVolume", "StorageClass":
		return true
	default:
		return false
	}
}

// ListResources fetches and transforms Kubernetes resources of a specific kind
// This is the business logic layer that handles the transformation of different resource types
// It filters resources based on user permissions before returning them
func (s *ResourceListService) ListResources(ctx context.Context, req ListResourcesRequest) ([]models.Resource, error) {
	// Validate namespace access using helper
	// Now returns a slice of namespaces to query (optimization for restricted users)
	targetNamespaces, err := s.validateNamespaceAccess(ctx, req.Namespace, req.AllNamespaces)
	if err != nil {
		return nil, err
	}

	// Create ListOptions
	listOpts := metav1.ListOptions{}
	if req.LabelSelector != "" {
		listOpts.LabelSelector = req.LabelSelector
	}

	var allResources []models.Resource

	// Optimization: For cluster-scoped resources, we effectively query once.
	// For namespaced resources, we iterate over the allowed namespaces.
	if isClusterScoped(req.Kind) {
		resources, err := s.fetchResourcesByKind(ctx, req, "", listOpts)
		if err != nil {
			return nil, err
		}
		allResources = resources
	} else {
		for _, ns := range targetNamespaces {
			resources, err := s.fetchResourcesByKind(ctx, req, ns, listOpts)
			if err != nil {
				return nil, fmt.Errorf("failed to list %s resources in namespace %s: %w", req.Kind, ns, err)
			}
			allResources = append(allResources, resources...)
		}
	}

	// Filter resources based on user permissions (final check)
	// This ensures that even if we query all namespaces, we only return resources the user has access to
	filteredResources, err := permissions.FilterResources(ctx, allResources)
	if err != nil {
		return nil, fmt.Errorf("failed to filter resources by permissions: %w", err)
	}

	return filteredResources, nil
}

// fetchResourcesByKind handles the specific listing logic for a single namespace scope
func (s *ResourceListService) fetchResourcesByKind(ctx context.Context, req ListResourcesRequest, namespace string, listOpts metav1.ListOptions) ([]models.Resource, error) {
	var resources []models.Resource
	var listErr error

	// Delegate to specific transformer based on kind
	switch req.Kind {
	case "Deployment":
		resources, listErr = s.listDeployments(ctx, req.Client, namespace, listOpts)
	case "Node":
		resources, listErr = s.listNodes(ctx, req.Client, listOpts)
	case "Pod":
		resources, listErr = s.listPods(ctx, req.Client, req.MetricsClient, namespace, listOpts)
	case "ConfigMap":
		resources, listErr = s.listConfigMaps(ctx, req.Client, namespace, listOpts)
	case "Secret":
		resources, listErr = s.listSecrets(ctx, req.Client, namespace, listOpts)
	case "Job":
		resources, listErr = s.listJobs(ctx, req.Client, namespace, listOpts)
	case "CronJob":
		resources, listErr = s.listCronJobs(ctx, req.Client, namespace, listOpts)
	case "StatefulSet":
		resources, listErr = s.listStatefulSets(ctx, req.Client, namespace, listOpts)
	case "DaemonSet":
		resources, listErr = s.listDaemonSets(ctx, req.Client, namespace, listOpts)
	case "HorizontalPodAutoscaler":
		resources, listErr = s.listHPAs(ctx, req.Client, namespace, listOpts)
	case "Service":
		resources, listErr = s.listServices(ctx, req.Client, namespace, listOpts)
	case "Ingress":
		resources, listErr = s.listIngresses(ctx, req.Client, namespace, listOpts)
	case "ServiceAccount":
		resources, listErr = s.listServiceAccounts(ctx, req.Client, namespace, listOpts)
	case "Role":
		resources, listErr = s.listRoles(ctx, req.Client, namespace, listOpts)
	case "ClusterRole":
		resources, listErr = s.listClusterRoles(ctx, req.Client, listOpts)
	case "RoleBinding":
		resources, listErr = s.listRoleBindings(ctx, req.Client, namespace, listOpts)
	case "ClusterRoleBinding":
		resources, listErr = s.listClusterRoleBindings(ctx, req.Client, listOpts)
	case "NetworkPolicy":
		resources, listErr = s.listNetworkPolicies(ctx, req.Client, namespace, listOpts)
	case "PersistentVolumeClaim":
		resources, listErr = s.listPVCs(ctx, req.Client, namespace, listOpts)
	case "PersistentVolume":
		resources, listErr = s.listPVs(ctx, req.Client, listOpts)
	case "StorageClass":
		resources, listErr = s.listStorageClasses(ctx, req.Client, listOpts)
	case "ResourceQuota":
		resources, listErr = s.listResourceQuotas(ctx, req.Client, namespace, listOpts)
	case "LimitRange":
		resources, listErr = s.listLimitRanges(ctx, req.Client, namespace, listOpts)
	default:
		// Return empty list for unknown kinds
		resources = []models.Resource{}
	}

	if listErr != nil {
		return nil, listErr
	}

	return resources, nil
}
