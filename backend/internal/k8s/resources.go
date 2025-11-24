package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

const (
	// DefaultListLimit is the maximum number of items to fetch in a single list operation
	// This prevents OOM issues in large clusters
	DefaultListLimit = int64(500)
)

// validatePromQLParam validates and escapes PromQL parameters to prevent injection
func validatePromQLParam(param, paramName string) (string, error) {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validPattern.MatchString(param) {
		return "", fmt.Errorf("invalid %s: contains invalid characters", paramName)
	}
	if len(param) > 253 {
		return "", fmt.Errorf("invalid %s: too long", paramName)
	}
	escaped := strings.ReplaceAll(param, `"`, `\"`)
	return escaped, nil
}

// createSecureHTTPClient creates an HTTP client with proper TLS certificate validation
func createSecureHTTPClient() *http.Client {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	config := &tls.Config{
		RootCAs: rootCAs,
	}

	transport := &http.Transport{TLSClientConfig: config}
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// queryPrometheusInstant queries Prometheus for instant metrics
func (s *Service) queryPrometheusInstant(query string) []map[string]interface{} {
	if s.handlers.PrometheusURL == "" {
		return []map[string]interface{}{}
	}

	promURL := fmt.Sprintf("%s/api/v1/query", s.handlers.PrometheusURL)
	params := url.Values{}
	params.Add("query", query)
	fullURL := fmt.Sprintf("%s?%s", promURL, params.Encode())

	client := createSecureHTTPClient()
	resp, err := client.Get(fullURL)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer resp.Body.Close()

	maxResponseSize := int64(10 << 20) // 10MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		return []map[string]interface{}{}
	}

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return []map[string]interface{}{}
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
		Error     string `json:"error"`
		ErrorType string `json:"errorType"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return []map[string]interface{}{}
	}

	if result.Status != "success" {
		return []map[string]interface{}{}
	}

	var results []map[string]interface{}
	for _, r := range result.Data.Result {
		resultMap := make(map[string]interface{})
		for k, v := range r.Metric {
			resultMap[k] = v
		}
		if len(r.Value) >= 2 {
			if valueStr, ok := r.Value[1].(string); ok {
				var floatValue float64
				fmt.Sscanf(valueStr, "%f", &floatValue)
				resultMap["value"] = floatValue
			}
		}
		results = append(results, resultMap)
	}

	return results
}

// GetResources returns resources of a specific kind
func (s *Service) GetResources(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout so cancellation propagates to Kubernetes API calls
	// This ensures that if the user cancels the HTTP request, the K8s API call is also cancelled
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()
	
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	allNamespaces := ns == "all"
	if ns == "" {
		ns = "default"
	}
	listNamespace := ns
	if allNamespaces {
		listNamespace = ""
	}

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := DefaultListLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 && parsedLimit <= 5000 {
			limit = parsedLimit
		}
	}
	continueToken := r.URL.Query().Get("continue")

	// Create ListOptions with limit and continue token for pagination
	listOpts := metav1.ListOptions{
		Limit:    limit,
		Continue: continueToken,
	}

	var resources []models.Resource
	var lastContinueToken string

	switch kind {
	case "Deployment":
		list, err := client.AppsV1().Deployments(listNamespace).List(ctx, listOpts)
		if err == nil {
			lastContinueToken = list.Continue
		if err == nil {
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

				details := models.DeploymentDetails{
					Replicas:  *i.Spec.Replicas,
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
					Status:    fmt.Sprintf("%d/%d", i.Status.ReadyReplicas, i.Status.Replicas),
					Created:   i.CreationTimestamp.Format(time.RFC3339),
					UID:       string(i.UID),
					Details:   details,
				})
			}
		}
	case "Node":
		list, err := client.CoreV1().Nodes().List(ctx, listOpts)
		if err == nil {
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
		}
	case "Pod":
		list, err := client.CoreV1().Pods(listNamespace).List(ctx, listOpts)
		if err == nil {
			lastContinueToken = list.Continue
		}
		metricsMap := make(map[string]models.PodMetric)
		if mclient := s.clusterService.GetMetricsClient(r); mclient != nil {
			if pmList, mErr := mclient.MetricsV1beta1().PodMetricses(listNamespace).List(ctx, listOpts); mErr == nil {
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
		if err == nil {
			for _, i := range list.Items {
				var containers []string
				for _, c := range i.Spec.Containers {
					containers = append(containers, c.Name)
				}

				restarts := int32(0)
				readyCount := int32(0)
				totalContainers := int32(len(i.Spec.Containers))
				var containerStatuses []map[string]interface{}

				for _, s := range i.Status.ContainerStatuses {
					restarts += s.RestartCount
					if s.Ready {
						readyCount++
					}

					containerStatus := map[string]interface{}{
						"name":         s.Name,
						"ready":        s.Ready,
						"restartCount": s.RestartCount,
						"image":        s.Image,
					}

					if s.State.Waiting != nil {
						containerStatus["state"] = "Waiting"
						containerStatus["reason"] = s.State.Waiting.Reason
						containerStatus["message"] = s.State.Waiting.Message
					} else if s.State.Running != nil {
						containerStatus["state"] = "Running"
						containerStatus["startedAt"] = s.State.Running.StartedAt.Format(time.RFC3339)
					} else if s.State.Terminated != nil {
						containerStatus["state"] = "Terminated"
						containerStatus["reason"] = s.State.Terminated.Reason
						containerStatus["exitCode"] = s.State.Terminated.ExitCode
						if !s.State.Terminated.StartedAt.IsZero() {
							containerStatus["startedAt"] = s.State.Terminated.StartedAt.Format(time.RFC3339)
						}
						if !s.State.Terminated.FinishedAt.IsZero() {
							containerStatus["finishedAt"] = s.State.Terminated.FinishedAt.Format(time.RFC3339)
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
		}
	case "ConfigMap":
		list, err := client.CoreV1().ConfigMaps(listNamespace).List(ctx, listOpts)
		if err == nil {
			lastContinueToken = list.Continue
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
		}
	case "Secret":
		list, err := client.CoreV1().Secrets(listNamespace).List(ctx, listOpts)
		if err == nil {
			lastContinueToken = list.Continue
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
		}
	case "Job":
		list, err := client.BatchV1().Jobs(listNamespace).List(ctx, listOpts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list Jobs: %v", err), http.StatusInternalServerError)
			return
		}
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
	case "CronJob":
		list, err := client.BatchV1().CronJobs(listNamespace).List(ctx, listOpts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list CronJobs: %v", err), http.StatusInternalServerError)
			return
		}
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
	case "StatefulSet":
		list, err := client.AppsV1().StatefulSets(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "DaemonSet":
		list, err := client.AppsV1().DaemonSets(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "HorizontalPodAutoscaler":
		list, err := client.AutoscalingV2().HorizontalPodAutoscalers(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "Service":
		list, err := client.CoreV1().Services(listNamespace).List(ctx, listOpts)
		if err == nil {
			for _, i := range list.Items {
				var ports []string
				for _, p := range i.Spec.Ports {
					ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.TargetPort.IntVal, p.Protocol))
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
		}
	case "Ingress":
		list, err := client.NetworkingV1().Ingresses(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "ServiceAccount":
		list, err := client.CoreV1().ServiceAccounts(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "Role":
		list, err := client.RbacV1().Roles(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "ClusterRole":
		list, err := client.RbacV1().ClusterRoles().List(ctx, listOpts)
		if err == nil {
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
		}
	case "RoleBinding":
		list, err := client.RbacV1().RoleBindings(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "ClusterRoleBinding":
		list, err := client.RbacV1().ClusterRoleBindings().List(ctx, listOpts)
		if err == nil {
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
		}
	case "NetworkPolicy":
		list, err := client.NetworkingV1().NetworkPolicies(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "PersistentVolumeClaim":
		list, err := client.CoreV1().PersistentVolumeClaims(listNamespace).List(ctx, listOpts)
		if err == nil {
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

				var usedBytes int64
				var totalBytes int64
				var usagePercent float64

				if s.handlers.PrometheusURL != "" && allocated != "" {
					validatedNamespace, _ := validatePromQLParam(i.Namespace, "namespace")
					validatedPVCName, _ := validatePromQLParam(i.Name, "pvc")

					usedQuery := fmt.Sprintf(
						`kubelet_volume_stats_used_bytes{persistentvolumeclaim="%s",namespace="%s"}`,
						validatedPVCName, validatedNamespace,
					)
					capacityQuery := fmt.Sprintf(
						`kubelet_volume_stats_capacity_bytes{persistentvolumeclaim="%s",namespace="%s"}`,
						validatedPVCName, validatedNamespace,
					)

					usedResult := s.queryPrometheusInstant(usedQuery)
					capacityResult := s.queryPrometheusInstant(capacityQuery)

					if len(usedResult) > 0 && len(capacityResult) > 0 {
						if usedVal, ok := usedResult[0]["value"].(float64); ok {
							usedBytes = int64(usedVal)
						}
						if capVal, ok := capacityResult[0]["value"].(float64); ok {
							totalBytes = int64(capVal)
							if totalBytes > 0 {
								usagePercent = (float64(usedBytes) / float64(totalBytes)) * 100
							}
						}
					}
				}

				details := map[string]interface{}{
					"accessModes":      i.Spec.AccessModes,
					"capacity":         allocated,
					"requested":        requested,
					"storageClassName": i.Spec.StorageClassName,
					"volumeName":       i.Spec.VolumeName,
				}

				if usedBytes > 0 && totalBytes > 0 {
					details["used"] = fmt.Sprintf("%.2fGi", float64(usedBytes)/(1024*1024*1024))
					details["total"] = fmt.Sprintf("%.2fGi", float64(totalBytes)/(1024*1024*1024))
					details["usagePercent"] = usagePercent
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
		}
	case "PersistentVolume":
		list, err := client.CoreV1().PersistentVolumes().List(ctx, listOpts)
		if err == nil {
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
		}
	case "StorageClass":
		list, err := client.StorageV1().StorageClasses().List(ctx, listOpts)
		if err == nil {
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
		}
	case "ResourceQuota":
		list, err := client.CoreV1().ResourceQuotas(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	case "LimitRange":
		list, err := client.CoreV1().LimitRanges(listNamespace).List(ctx, listOpts)
		if err == nil {
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
		}
	default:
		// Return empty list for unknown kinds
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine continue token and remaining count from the last list operation
	// Note: We need to track this per resource type, but for simplicity we'll use a generic approach
	var continueToken string
	var remaining int
	
	// Try to get continue token from the last list call
	// Since we have multiple switch cases, we'll need to handle this differently
	// For now, we'll return the continue token if we have one from listOpts
	// In a real implementation, each list call would return metadata with continue token
	
	// Create paginated response
	response := models.PaginatedResources{
		Resources: resources,
		Continue:  lastContinueToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

