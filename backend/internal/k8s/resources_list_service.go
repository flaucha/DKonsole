package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/models"
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
	Client        kubernetes.Interface
	MetricsClient *metricsv.Clientset
}

// ListResources fetches and transforms Kubernetes resources of a specific kind
// This is the business logic layer that handles the transformation of different resource types
func (s *ResourceListService) ListResources(ctx context.Context, req ListResourcesRequest) ([]models.Resource, error) {
	ns := req.Namespace
	allNamespaces := req.AllNamespaces
	if ns == "" {
		ns = "default"
	}
	listNamespace := ns
	if allNamespaces {
		listNamespace = ""
	}

	// Create ListOptions - no pagination limits, load all resources
	listOpts := metav1.ListOptions{}

	var resources []models.Resource
	var err error

	// Delegate to specific transformer based on kind
	switch req.Kind {
	case "Deployment":
		resources, err = s.listDeployments(ctx, req.Client, listNamespace, listOpts)
	case "Node":
		resources, err = s.listNodes(ctx, req.Client, listOpts)
	case "Pod":
		resources, err = s.listPods(ctx, req.Client, req.MetricsClient, listNamespace, listOpts)
	case "ConfigMap":
		resources, err = s.listConfigMaps(ctx, req.Client, listNamespace, listOpts)
	case "Secret":
		resources, err = s.listSecrets(ctx, req.Client, listNamespace, listOpts)
	case "Job":
		resources, err = s.listJobs(ctx, req.Client, listNamespace, listOpts)
	case "CronJob":
		resources, err = s.listCronJobs(ctx, req.Client, listNamespace, listOpts)
	case "StatefulSet":
		resources, err = s.listStatefulSets(ctx, req.Client, listNamespace, listOpts)
	case "DaemonSet":
		resources, err = s.listDaemonSets(ctx, req.Client, listNamespace, listOpts)
	case "HorizontalPodAutoscaler":
		resources, err = s.listHPAs(ctx, req.Client, listNamespace, listOpts)
	case "Service":
		resources, err = s.listServices(ctx, req.Client, listNamespace, listOpts)
	case "Ingress":
		resources, err = s.listIngresses(ctx, req.Client, listNamespace, listOpts)
	case "ServiceAccount":
		resources, err = s.listServiceAccounts(ctx, req.Client, listNamespace, listOpts)
	case "Role":
		resources, err = s.listRoles(ctx, req.Client, listNamespace, listOpts)
	case "ClusterRole":
		resources, err = s.listClusterRoles(ctx, req.Client, listOpts)
	case "RoleBinding":
		resources, err = s.listRoleBindings(ctx, req.Client, listNamespace, listOpts)
	case "ClusterRoleBinding":
		resources, err = s.listClusterRoleBindings(ctx, req.Client, listOpts)
	case "NetworkPolicy":
		resources, err = s.listNetworkPolicies(ctx, req.Client, listNamespace, listOpts)
	case "PersistentVolumeClaim":
		resources, err = s.listPVCs(ctx, req.Client, listNamespace, listOpts)
	case "PersistentVolume":
		resources, err = s.listPVs(ctx, req.Client, listOpts)
	case "StorageClass":
		resources, err = s.listStorageClasses(ctx, req.Client, listOpts)
	case "ResourceQuota":
		resources, err = s.listResourceQuotas(ctx, req.Client, listNamespace, listOpts)
	case "LimitRange":
		resources, err = s.listLimitRanges(ctx, req.Client, listNamespace, listOpts)
	default:
		// Return empty list for unknown kinds
		resources = []models.Resource{}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list %s resources: %w", req.Kind, err)
	}

	return resources, nil
}

// Transformer functions - Each function transforms a specific Kubernetes resource type
// to the models.Resource format used by the frontend

func (s *ResourceListService) listDeployments(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var images []string
		var ports []int32
		var pvcs []string

		for _, c := range i.Spec.Template.Spec.Containers {
			images = append(images, c.Image)
			for _, p := range c.Ports {
				ports = append(ports, p.ContainerPort)
			}
		}
		for _, v := range i.Spec.Template.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				pvcs = append(pvcs, v.PersistentVolumeClaim.ClaimName)
			}
		}

		var replicas int32
		if i.Spec.Replicas != nil {
			replicas = *i.Spec.Replicas
		}

		details := models.DeploymentDetails{
			Replicas:  replicas,
			Ready:     i.Status.ReadyReplicas,
			Images:    images,
			Ports:     ports,
			PVCs:      pvcs,
			PodLabels: i.Spec.Selector.MatchLabels,
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Deployment",
			Status:    fmt.Sprintf("%d/%d", i.Status.ReadyReplicas, replicas),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listNodes(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Nodes().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		conditions := make(map[string]string)
		for _, c := range i.Status.Conditions {
			conditions[string(c.Type)] = string(c.Status)
		}

		var images []string
		for _, img := range i.Status.Images {
			if len(img.Names) > 0 {
				images = append(images, img.Names[0])
			}
		}

		details := map[string]interface{}{
			"addresses":     i.Status.Addresses,
			"nodeInfo":      i.Status.NodeInfo,
			"capacity":      i.Status.Capacity,
			"allocatable":   i.Status.Allocatable,
			"conditions":    conditions,
			"images":        images,
			"taints":        i.Spec.Taints,
			"podCIDR":       i.Spec.PodCIDR,
			"podCIDRs":      i.Spec.PodCIDRs,
			"unschedulable": i.Spec.Unschedulable,
			"labels":        i.Labels,
			"annotations":   i.Annotations,
		}

		status := "Ready"
		for _, c := range i.Status.Conditions {
			if c.Type == corev1.NodeReady && c.Status != corev1.ConditionTrue {
				status = "NotReady"
			}
		}
		if i.Spec.Unschedulable {
			status += ",SchedulingDisabled"
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "Node",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listPods(ctx context.Context, client kubernetes.Interface, metricsClient *metricsv.Clientset, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Build metrics map
	metricsMap := make(map[string]models.PodMetric)
	if metricsClient != nil {
		if pmList, mErr := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts); mErr == nil {
			for _, pm := range pmList.Items {
				var cpuMilli int64
				var memBytes int64
				for _, c := range pm.Containers {
					cpuMilli += c.Usage.Cpu().MilliValue()
					memBytes += c.Usage.Memory().Value()
				}
				metricsMap[pm.Name] = models.PodMetric{
					CPU:    fmt.Sprintf("%dm", cpuMilli),
					Memory: fmt.Sprintf("%.1fMi", float64(memBytes)/(1024*1024)),
				}
			}
		}
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var containers []string
		for _, c := range i.Spec.Containers {
			containers = append(containers, c.Name)
		}

		restarts := int32(0)
		readyCount := int32(0)
		totalContainers := int32(len(i.Spec.Containers))
		var containerStatuses []map[string]interface{}

		for _, cs := range i.Status.ContainerStatuses {
			restarts += cs.RestartCount
			if cs.Ready {
				readyCount++
			}

			containerStatus := map[string]interface{}{
				"name":         cs.Name,
				"ready":        cs.Ready,
				"restartCount": cs.RestartCount,
				"image":        cs.Image,
			}

			if cs.State.Waiting != nil {
				containerStatus["state"] = "Waiting"
				containerStatus["reason"] = cs.State.Waiting.Reason
				containerStatus["message"] = cs.State.Waiting.Message
			} else if cs.State.Running != nil {
				containerStatus["state"] = "Running"
				containerStatus["startedAt"] = cs.State.Running.StartedAt.Format(time.RFC3339)
			} else if cs.State.Terminated != nil {
				containerStatus["state"] = "Terminated"
				containerStatus["reason"] = cs.State.Terminated.Reason
				containerStatus["exitCode"] = cs.State.Terminated.ExitCode
				if !cs.State.Terminated.StartedAt.IsZero() {
					containerStatus["startedAt"] = cs.State.Terminated.StartedAt.Format(time.RFC3339)
				}
				if !cs.State.Terminated.FinishedAt.IsZero() {
					containerStatus["finishedAt"] = cs.State.Terminated.FinishedAt.Format(time.RFC3339)
				}
			}

			containerStatuses = append(containerStatuses, containerStatus)
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Pod",
			Status:    string(i.Status.Phase),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"node":              i.Spec.NodeName,
				"ip":                i.Status.PodIP,
				"restarts":          restarts,
				"ready":             fmt.Sprintf("%d/%d", readyCount, totalContainers),
				"readyCount":        readyCount,
				"totalContainers":   totalContainers,
				"containers":        containers,
				"containerStatuses": containerStatuses,
				"metrics":           metricsMap[i.Name],
				"labels":            i.Labels,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listConfigMaps(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().ConfigMaps(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "ConfigMap",
			Status:    fmt.Sprintf("%d keys", len(i.Data)),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"data": i.Data,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listSecrets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Secrets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		data := make(map[string]string)
		for k, v := range i.Data {
			data[k] = string(v)
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Secret",
			Status:    string(i.Type),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"type":      string(i.Type),
				"data":      data,
				"keysCount": len(data),
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listJobs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.BatchV1().Jobs(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		status := "Running"
		if i.Status.Succeeded > 0 {
			status = "Completed"
		} else if i.Status.Failed > 0 {
			status = "Failed"
		}
		details := map[string]interface{}{
			"active":         i.Status.Active,
			"succeeded":      i.Status.Succeeded,
			"failed":         i.Status.Failed,
			"startTime":      i.Status.StartTime,
			"completionTime": i.Status.CompletionTime,
			"parallelism":    i.Spec.Parallelism,
			"completions":    i.Spec.Completions,
			"backoffLimit":   i.Spec.BackoffLimit,
			"activeDeadline": i.Spec.ActiveDeadlineSeconds,
			"podSelector":    i.Spec.Selector,
			"podTemplate":    i.Spec.Template.Spec,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Job",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listCronJobs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.BatchV1().CronJobs(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var lastSchedule string
		if i.Status.LastScheduleTime != nil {
			lastSchedule = i.Status.LastScheduleTime.Format(time.RFC3339)
		}
		details := map[string]interface{}{
			"schedule":         i.Spec.Schedule,
			"suspend":          i.Spec.Suspend,
			"concurrency":      i.Spec.ConcurrencyPolicy,
			"startingDeadline": i.Spec.StartingDeadlineSeconds,
			"lastSchedule":     lastSchedule,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "CronJob",
			Status:    i.Spec.Schedule,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listStatefulSets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().StatefulSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"replicas":       i.Status.Replicas,
			"ready":          i.Status.ReadyReplicas,
			"current":        i.Status.CurrentReplicas,
			"update":         i.Status.UpdatedReplicas,
			"serviceName":    i.Spec.ServiceName,
			"podManagement":  i.Spec.PodManagementPolicy,
			"updateStrategy": i.Spec.UpdateStrategy,
			"volumeClaims":   i.Spec.VolumeClaimTemplates,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "StatefulSet",
			Status:    fmt.Sprintf("%d/%d", i.Status.ReadyReplicas, i.Status.Replicas),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listDaemonSets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().DaemonSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"desired":      i.Status.DesiredNumberScheduled,
			"current":      i.Status.CurrentNumberScheduled,
			"ready":        i.Status.NumberReady,
			"available":    i.Status.NumberAvailable,
			"updated":      i.Status.UpdatedNumberScheduled,
			"misscheduled": i.Status.NumberMisscheduled,
			"nodeSelector": i.Spec.Template.Spec.NodeSelector,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "DaemonSet",
			Status:    fmt.Sprintf("%d/%d", i.Status.NumberReady, i.Status.DesiredNumberScheduled),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listHPAs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		status := fmt.Sprintf("%d/%d replicas", i.Status.CurrentReplicas, i.Status.DesiredReplicas)
		details := map[string]interface{}{
			"minReplicas":   i.Spec.MinReplicas,
			"maxReplicas":   i.Spec.MaxReplicas,
			"current":       i.Status.CurrentReplicas,
			"desired":       i.Status.DesiredReplicas,
			"metrics":       i.Spec.Metrics,
			"lastScaleTime": i.Status.LastScaleTime,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "HPA",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listServices(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var ports []string
		for _, p := range i.Spec.Ports {
			var targetPort string
			if p.TargetPort.Type == 0 { // IntOrString with int value
				targetPort = fmt.Sprintf("%d", p.TargetPort.IntVal)
			} else {
				targetPort = p.TargetPort.StrVal
			}
			ports = append(ports, fmt.Sprintf("%d:%s/%s", p.Port, targetPort, p.Protocol))
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Service",
			Status:    string(i.Spec.Type),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"clusterIP": i.Spec.ClusterIP,
				"ports":     ports,
				"selector":  i.Spec.Selector,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listIngresses(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var tls []map[string]interface{}
		for _, t := range i.Spec.TLS {
			tls = append(tls, map[string]interface{}{
				"hosts":      t.Hosts,
				"secretName": t.SecretName,
			})
		}

		var rules []map[string]interface{}
		for _, r := range i.Spec.Rules {
			var paths []string
			if r.HTTP != nil {
				for _, p := range r.HTTP.Paths {
					svcName := "unknown"
					svcPort := int32(0)
					if p.Backend.Service != nil {
						svcName = p.Backend.Service.Name
						svcPort = p.Backend.Service.Port.Number
					}
					paths = append(paths, fmt.Sprintf("%s -> %s:%d", p.Path, svcName, svcPort))
				}
			}
			rules = append(rules, map[string]interface{}{
				"host":  r.Host,
				"paths": paths,
			})
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Ingress",
			Status:    fmt.Sprintf("%d rules", len(i.Spec.Rules)),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"rules":        rules,
				"tls":          tls,
				"annotations":  i.Annotations,
				"loadBalancer": i.Status.LoadBalancer.Ingress,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listServiceAccounts(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().ServiceAccounts(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"secrets":          i.Secrets,
			"imagePullSecrets": i.ImagePullSecrets,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "ServiceAccount",
			Status:    fmt.Sprintf("%d secrets", len(i.Secrets)),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listRoles(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().Roles(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"rules": i.Rules,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Role",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listClusterRoles(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().ClusterRoles().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"rules": i.Rules,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "ClusterRole",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listRoleBindings(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().RoleBindings(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"subjects": i.Subjects,
			"roleRef":  i.RoleRef,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "RoleBinding",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listClusterRoleBindings(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().ClusterRoleBindings().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"subjects": i.Subjects,
			"roleRef":  i.RoleRef,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "ClusterRoleBinding",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listNetworkPolicies(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "NetworkPolicy",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"podSelector": i.Spec.PodSelector.MatchLabels,
				"policyTypes": i.Spec.PolicyTypes,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listPVCs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		requested := ""
		if i.Spec.Resources.Requests != nil {
			if storage, ok := i.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				requested = storage.String()
			}
		}

		allocated := ""
		if i.Status.Capacity != nil {
			if storage, ok := i.Status.Capacity[corev1.ResourceStorage]; ok {
				allocated = storage.String()
			}
		}

		// Note: Prometheus metrics for PVC usage would require queryPrometheusInstant
		// For now, we skip Prometheus metrics in the refactored version
		// This can be added later if needed by injecting a Prometheus client

		details := map[string]interface{}{
			"accessModes":      i.Spec.AccessModes,
			"capacity":         allocated,
			"requested":        requested,
			"storageClassName": i.Spec.StorageClassName,
			"volumeName":       i.Spec.VolumeName,
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "PersistentVolumeClaim",
			Status:    string(i.Status.Phase),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listPVs(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().PersistentVolumes().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "PersistentVolume",
			Status:    string(i.Status.Phase),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"accessModes":      i.Spec.AccessModes,
				"capacity":         i.Spec.Capacity.Storage().String(),
				"storageClassName": i.Spec.StorageClassName,
				"claimRef":         i.Spec.ClaimRef,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listStorageClasses(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.StorageV1().StorageClasses().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		reclaim := ""
		if i.ReclaimPolicy != nil {
			reclaim = string(*i.ReclaimPolicy)
		}
		volumeBinding := ""
		if i.VolumeBindingMode != nil {
			volumeBinding = string(*i.VolumeBindingMode)
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "StorageClass",
			Status:    reclaim,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"provisioner":          i.Provisioner,
				"reclaimPolicy":        reclaim,
				"volumeBindingMode":    volumeBinding,
				"allowVolumeExpansion": i.AllowVolumeExpansion,
				"parameters":           i.Parameters,
				"mountOptions":         i.MountOptions,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listResourceQuotas(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "ResourceQuota",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"hard": i.Status.Hard,
				"used": i.Status.Used,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listLimitRanges(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().LimitRanges(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "LimitRange",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"limits": i.Spec.Limits,
			},
		})
	}

	return resources, nil
}
