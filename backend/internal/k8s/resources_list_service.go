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

// ListResources fetches and transforms Kubernetes resources of a specific kind
// This is the business logic layer that handles the transformation of different resource types
// It filters resources based on user permissions before returning them
func (s *ResourceListService) ListResources(ctx context.Context, req ListResourcesRequest) ([]models.Resource, error) {
	// Validate namespace access using helper
	listNamespace, err := s.validateNamespaceAccess(ctx, req.Namespace, req.AllNamespaces)
	if err != nil {
		return nil, err
	}

	// Create ListOptions
	listOpts := metav1.ListOptions{}
	if req.LabelSelector != "" {
		listOpts.LabelSelector = req.LabelSelector
	}

	var resources []models.Resource
	var listErr error

	// Delegate to specific transformer based on kind
	switch req.Kind {
	case "Deployment":
		resources, listErr = s.listDeployments(ctx, req.Client, listNamespace, listOpts)
	case "Node":
		resources, listErr = s.listNodes(ctx, req.Client, listOpts)
	case "Pod":
		resources, listErr = s.listPods(ctx, req.Client, req.MetricsClient, listNamespace, listOpts)
	case "ConfigMap":
		resources, listErr = s.listConfigMaps(ctx, req.Client, listNamespace, listOpts)
	case "Secret":
		resources, listErr = s.listSecrets(ctx, req.Client, listNamespace, listOpts)
	case "Job":
		resources, listErr = s.listJobs(ctx, req.Client, listNamespace, listOpts)
	case "CronJob":
		resources, listErr = s.listCronJobs(ctx, req.Client, listNamespace, listOpts)
	case "StatefulSet":
		resources, listErr = s.listStatefulSets(ctx, req.Client, listNamespace, listOpts)
	case "DaemonSet":
		resources, listErr = s.listDaemonSets(ctx, req.Client, listNamespace, listOpts)
	case "HorizontalPodAutoscaler":
		resources, listErr = s.listHPAs(ctx, req.Client, listNamespace, listOpts)
	case "Service":
		resources, listErr = s.listServices(ctx, req.Client, listNamespace, listOpts)
	case "Ingress":
		resources, listErr = s.listIngresses(ctx, req.Client, listNamespace, listOpts)
	case "ServiceAccount":
		resources, listErr = s.listServiceAccounts(ctx, req.Client, listNamespace, listOpts)
	case "Role":
		resources, listErr = s.listRoles(ctx, req.Client, listNamespace, listOpts)
	case "ClusterRole":
		resources, listErr = s.listClusterRoles(ctx, req.Client, listOpts)
	case "RoleBinding":
		resources, listErr = s.listRoleBindings(ctx, req.Client, listNamespace, listOpts)
	case "ClusterRoleBinding":
		resources, listErr = s.listClusterRoleBindings(ctx, req.Client, listOpts)
	case "NetworkPolicy":
		resources, listErr = s.listNetworkPolicies(ctx, req.Client, listNamespace, listOpts)
	case "PersistentVolumeClaim":
		resources, listErr = s.listPVCs(ctx, req.Client, listNamespace, listOpts)
	case "PersistentVolume":
		resources, listErr = s.listPVs(ctx, req.Client, listOpts)
	case "StorageClass":
		resources, listErr = s.listStorageClasses(ctx, req.Client, listOpts)
	case "ResourceQuota":
		resources, listErr = s.listResourceQuotas(ctx, req.Client, listNamespace, listOpts)
	case "LimitRange":
		resources, listErr = s.listLimitRanges(ctx, req.Client, listNamespace, listOpts)
	default:
		// Return empty list for unknown kinds
		resources = []models.Resource{}
	}

	if listErr != nil {
		return nil, fmt.Errorf("failed to list %s resources: %w", req.Kind, listErr)
	}

	// Filter resources based on user permissions (final check)
	// This ensures that even if we query all namespaces, we only return resources the user has access to
	filteredResources, err := permissions.FilterResources(ctx, resources)
	if err != nil {
		return nil, fmt.Errorf("failed to filter resources by permissions: %w", err)
	}

	return filteredResources, nil
}
