package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespaceRepository defines the interface for fetching Kubernetes namespaces
type NamespaceRepository interface {
	List(ctx context.Context) ([]corev1.Namespace, error)
}

// K8sNamespaceRepository implements NamespaceRepository
type K8sNamespaceRepository struct {
	client kubernetes.Interface
}

// NewK8sNamespaceRepository creates a new K8sNamespaceRepository
func NewK8sNamespaceRepository(client kubernetes.Interface) *K8sNamespaceRepository {
	return &K8sNamespaceRepository{client: client}
}

// List fetches all namespaces
func (r *K8sNamespaceRepository) List(ctx context.Context) ([]corev1.Namespace, error) {
	namespaces, err := r.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	return namespaces.Items, nil
}

// ClusterStatsRepository defines the interface for fetching cluster statistics
type ClusterStatsRepository interface {
	GetNodeCount(ctx context.Context) (int, error)
	GetNamespaceCount(ctx context.Context) (int, error)
	GetPodCount(ctx context.Context) (int, error)
	GetDeploymentCount(ctx context.Context) (int, error)
	GetServiceCount(ctx context.Context) (int, error)
	GetIngressCount(ctx context.Context) (int, error)
	GetPVCCount(ctx context.Context) (int, error)
	GetPVCount(ctx context.Context) (int, error)
}

// K8sClusterStatsRepository implements ClusterStatsRepository
type K8sClusterStatsRepository struct {
	client kubernetes.Interface
}

// NewK8sClusterStatsRepository creates a new K8sClusterStatsRepository
func NewK8sClusterStatsRepository(client kubernetes.Interface) *K8sClusterStatsRepository {
	return &K8sClusterStatsRepository{client: client}
}

// isControlPlaneNode checks if a node is a control plane/master node
// Control plane nodes are identified by:
// - Taints with key "node-role.kubernetes.io/control-plane" or "node-role.kubernetes.io/master"
// - Labels with "node-role.kubernetes.io/control-plane" or "node-role.kubernetes.io/master"
func isControlPlaneNode(node corev1.Node) bool {
	// Check labels
	if val, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok && val != "" {
		return true
	}
	if val, ok := node.Labels["node-role.kubernetes.io/master"]; ok && val != "" {
		return true
	}

	// Check taints
	for _, taint := range node.Spec.Taints {
		if taint.Key == "node-role.kubernetes.io/control-plane" || taint.Key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false
}

// GetNodeCount returns the number of worker nodes (excluding control plane nodes)
func (r *K8sClusterStatsRepository) GetNodeCount(ctx context.Context) (int, error) {
	nodes, err := r.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get node count: %w", err)
	}

	// Filter out control plane nodes
	workerNodeCount := 0
	for _, node := range nodes.Items {
		if !isControlPlaneNode(node) {
			workerNodeCount++
		}
	}

	return workerNodeCount, nil
}

// GetNamespaceCount returns the number of namespaces
func (r *K8sClusterStatsRepository) GetNamespaceCount(ctx context.Context) (int, error) {
	namespaces, err := r.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get namespace count: %w", err)
	}
	return len(namespaces.Items), nil
}

// GetPodCount returns the number of pods
func (r *K8sClusterStatsRepository) GetPodCount(ctx context.Context) (int, error) {
	pods, err := r.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get pod count: %w", err)
	}
	return len(pods.Items), nil
}

// GetDeploymentCount returns the number of deployments
func (r *K8sClusterStatsRepository) GetDeploymentCount(ctx context.Context) (int, error) {
	deployments, err := r.client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get deployment count: %w", err)
	}
	return len(deployments.Items), nil
}

// GetServiceCount returns the number of services
func (r *K8sClusterStatsRepository) GetServiceCount(ctx context.Context) (int, error) {
	services, err := r.client.CoreV1().Services("").List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get service count: %w", err)
	}
	return len(services.Items), nil
}

// GetIngressCount returns the number of ingresses
func (r *K8sClusterStatsRepository) GetIngressCount(ctx context.Context) (int, error) {
	ingresses, err := r.client.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get ingress count: %w", err)
	}
	return len(ingresses.Items), nil
}

// GetPVCCount returns the number of PVCs
func (r *K8sClusterStatsRepository) GetPVCCount(ctx context.Context) (int, error) {
	pvcs, err := r.client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get PVC count: %w", err)
	}
	return len(pvcs.Items), nil
}

// GetPVCount returns the number of PVs
func (r *K8sClusterStatsRepository) GetPVCount(ctx context.Context) (int, error) {
	pvs, err := r.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return 0, fmt.Errorf("failed to get PV count: %w", err)
	}
	return len(pvs.Items), nil
}

// DeploymentRepository defines the interface for Deployment operations
type DeploymentRepository interface {
	GetScale(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error)
	UpdateScale(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error)
	GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error)
}

// K8sDeploymentRepository implements DeploymentRepository
type K8sDeploymentRepository struct {
	client kubernetes.Interface
}

// NewK8sDeploymentRepository creates a new K8sDeploymentRepository
func NewK8sDeploymentRepository(client kubernetes.Interface) *K8sDeploymentRepository {
	return &K8sDeploymentRepository{client: client}
}

// GetScale gets the scale of a deployment
func (r *K8sDeploymentRepository) GetScale(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error) {
	scale, err := r.client.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment scale: %w", err)
	}
	return scale, nil
}

// UpdateScale updates the scale of a deployment
func (r *K8sDeploymentRepository) UpdateScale(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error) {
	// Preserve ResourceVersion and other metadata for proper update
	updateOptions := metav1.UpdateOptions{}
	updatedScale, err := r.client.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment scale: %w", err)
	}
	return updatedScale, nil
}

// GetDeployment gets a deployment by name
func (r *K8sDeploymentRepository) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := r.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return deployment, nil
}

// UpdateDeployment updates a deployment
func (r *K8sDeploymentRepository) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	updatedDeployment, err := r.client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment: %w", err)
	}
	return updatedDeployment, nil
}
