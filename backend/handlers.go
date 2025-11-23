package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"regexp"

	"github.com/gorilla/websocket"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/yaml"
)

type Handlers struct {
	Clients       map[string]*kubernetes.Clientset
	Dynamics      map[string]dynamic.Interface
	Metrics       map[string]*metricsv.Clientset
	RESTConfigs   map[string]*rest.Config
	PrometheusURL string
	mu            sync.RWMutex
}

type ClusterConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Token    string `json:"token"`
	Insecure bool   `json:"insecure"`
}

type Namespace struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Labels  map[string]string `json:"labels"`
	Created string            `json:"created"`
}

type Resource struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status"`
	Created   string      `json:"created"`
	UID       string      `json:"uid"`
	Details   interface{} `json:"details,omitempty"`
}

type DeploymentDetails struct {
	Replicas  int32             `json:"replicas"`
	Ready     int32             `json:"ready"`
	Images    []string          `json:"images"`
	Ports     []int32           `json:"ports"`
	PVCs      []string          `json:"pvcs"`
	PodLabels map[string]string `json:"podLabels"`
}

type PodMetric struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type ResourceMeta struct {
	Group      string
	Version    string
	Resource   string
	Namespaced bool
}

var resourceMeta = map[string]ResourceMeta{
	"Deployment":              {Group: "apps", Version: "v1", Resource: "deployments", Namespaced: true},
	"Node":                    {Group: "", Version: "v1", Resource: "nodes", Namespaced: false},
	"Pod":                     {Group: "", Version: "v1", Resource: "pods", Namespaced: true},
	"ConfigMap":               {Group: "", Version: "v1", Resource: "configmaps", Namespaced: true},
	"Secret":                  {Group: "", Version: "v1", Resource: "secrets", Namespaced: true},
	"Job":                     {Group: "batch", Version: "v1", Resource: "jobs", Namespaced: true},
	"CronJob":                 {Group: "batch", Version: "v1", Resource: "cronjobs", Namespaced: true},
	"StatefulSet":             {Group: "apps", Version: "v1", Resource: "statefulsets", Namespaced: true},
	"DaemonSet":               {Group: "apps", Version: "v1", Resource: "daemonsets", Namespaced: true},
	"HorizontalPodAutoscaler": {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers", Namespaced: true},
	"Service":                 {Group: "", Version: "v1", Resource: "services", Namespaced: true},
	"Ingress":                 {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses", Namespaced: true},
	"NetworkPolicy":           {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies", Namespaced: true},
	"PersistentVolumeClaim":   {Group: "", Version: "v1", Resource: "persistentvolumeclaims", Namespaced: true},
	"PersistentVolume":        {Group: "", Version: "v1", Resource: "persistentvolumes", Namespaced: false},
	"StorageClass":            {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses", Namespaced: false},
	"ServiceAccount":          {Group: "", Version: "v1", Resource: "serviceaccounts", Namespaced: true},
	"Role":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles", Namespaced: true},
	"ClusterRole":             {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles", Namespaced: false},
	"RoleBinding":             {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings", Namespaced: true},
	"ClusterRoleBinding":      {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings", Namespaced: false},
	"ResourceQuota":           {Group: "", Version: "v1", Resource: "resourcequotas", Namespaced: true},
	"LimitRange":              {Group: "", Version: "v1", Resource: "limitranges", Namespaced: true},
}

func resolveGVR(kind string) (schema.GroupVersionResource, bool) {
	meta, ok := resourceMeta[kind]
	if !ok {
		return schema.GroupVersionResource{}, false
	}
	return schema.GroupVersionResource{
		Group:    meta.Group,
		Version:  meta.Version,
		Resource: meta.Resource,
	}, true
}

func (h *Handlers) GetClusters(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clusters []string
	for name := range h.Clients {
		clusters = append(clusters, name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// HealthHandler is an unauthenticated liveness endpoint
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handlers) AddCluster(w http.ResponseWriter, r *http.Request) {
	var config ClusterConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if config.Name == "" || config.Host == "" || config.Token == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Clients[config.Name]; exists {
		http.Error(w, "Cluster with this name already exists", http.StatusConflict)
		return
	}

	k8sConfig := &rest.Config{
		Host:        config.Host,
		BearerToken: config.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: config.Insecure,
		},
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create client: %v", err), http.StatusInternalServerError)
		return
	}

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create dynamic client: %v", err), http.StatusInternalServerError)
		return
	}

	metricsClient, _ := metricsv.NewForConfig(k8sConfig)

	h.Clients[config.Name] = clientset
	h.Dynamics[config.Name] = dynamicClient
	h.RESTConfigs[config.Name] = k8sConfig
	if metricsClient != nil {
		h.Metrics[config.Name] = metricsClient
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) getClient(r *http.Request) (*kubernetes.Clientset, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.Clients[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

func (h *Handlers) getDynamicClient(r *http.Request) (dynamic.Interface, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.Dynamics[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

func (h *Handlers) getMetricsClient(r *http.Request) *metricsv.Clientset {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.Metrics[cluster]
	if !ok {
		return nil
	}
	return client
}

func (h *Handlers) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []Namespace
	for _, ns := range namespaces.Items {
		result = append(result, Namespace{
			Name:    ns.Name,
			Status:  string(ns.Status.Phase),
			Labels:  ns.Labels,
			Created: ns.CreationTimestamp.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handlers) GetResources(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
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

	var resources []Resource

	switch kind {
	case "Deployment":
		list, e := client.AppsV1().Deployments(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
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

				details := DeploymentDetails{
					Replicas:  *i.Spec.Replicas,
					Ready:     i.Status.ReadyReplicas,
					Images:    images,
					Ports:     ports,
					PVCs:      pvcs,
					PodLabels: i.Spec.Selector.MatchLabels,
				}

				resources = append(resources, Resource{
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
		list, e := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		err = e
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

				resources = append(resources, Resource{
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
		list, e := client.CoreV1().Pods(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		metricsMap := make(map[string]PodMetric)
		if mclient := h.getMetricsClient(r); mclient != nil {
			if pmList, mErr := mclient.MetricsV1beta1().PodMetricses(listNamespace).List(context.TODO(), metav1.ListOptions{}); mErr == nil {
				for _, pm := range pmList.Items {
					var cpuMilli int64
					var memBytes int64
					for _, c := range pm.Containers {
						cpuMilli += c.Usage.Cpu().MilliValue()
						memBytes += c.Usage.Memory().Value()
					}
					metricsMap[pm.Name] = PodMetric{
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
				for _, s := range i.Status.ContainerStatuses {
					restarts += s.RestartCount
				}

				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: i.Namespace,
					Kind:      "Pod",
					Status:    string(i.Status.Phase),
					Created:   i.CreationTimestamp.Format(time.RFC3339),
					UID:       string(i.UID),
					Details: map[string]interface{}{
						"node":       i.Spec.NodeName,
						"ip":         i.Status.PodIP,
						"restarts":   restarts,
						"containers": containers,
						"metrics":    metricsMap[i.Name],
						"labels":     i.Labels,
					},
				})
			}
		}
	case "ConfigMap":
		list, e := client.CoreV1().ConfigMaps(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				resources = append(resources, Resource{
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
		list, e := client.CoreV1().Secrets(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				// Security: Do not expose secret data values
				keys := make([]string, 0, len(i.Data))
				for k := range i.Data {
					keys = append(keys, k)
				}
				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: i.Namespace,
					Kind:      "Secret",
					Status:    string(i.Type),
					Created:   i.CreationTimestamp.Format(time.RFC3339),
					UID:       string(i.UID),
					Details: map[string]interface{}{
						"type":      string(i.Type),
						"keys":      keys,
						"keysCount": len(keys),
					},
				})
			}
		}
	case "Job":
		list, e := client.BatchV1().Jobs(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
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
				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: i.Namespace,
					Kind:      "Job",
					Status:    status,
					Created:   i.CreationTimestamp.Format(time.RFC3339),
					UID:       string(i.UID),
					Details:   details,
				})
			}
		}
	case "CronJob":
		list, e := client.BatchV1().CronJobs(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
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
				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: i.Namespace,
					Kind:      "CronJob",
					Status:    i.Spec.Schedule,
					Created:   i.CreationTimestamp.Format(time.RFC3339),
					UID:       string(i.UID),
					Details:   details,
				})
			}
		}
	case "StatefulSet":
		list, e := client.AppsV1().StatefulSets(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
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
				resources = append(resources, Resource{
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
		list, e := client.AppsV1().DaemonSets(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
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
				resources = append(resources, Resource{
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
		list, e := client.AutoscalingV2().HorizontalPodAutoscalers(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
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
				resources = append(resources, Resource{
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
		list, e := client.CoreV1().Services(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				var ports []string
				for _, p := range i.Spec.Ports {
					ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.TargetPort.IntVal, p.Protocol))
				}
				resources = append(resources, Resource{
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
		list, e := client.NetworkingV1().Ingresses(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
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

				resources = append(resources, Resource{
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
		list, e := client.CoreV1().ServiceAccounts(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				details := map[string]interface{}{
					"secrets":          i.Secrets,
					"imagePullSecrets": i.ImagePullSecrets,
				}
				resources = append(resources, Resource{
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
		list, e := client.RbacV1().Roles(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				details := map[string]interface{}{
					"rules": i.Rules,
				}
				resources = append(resources, Resource{
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
		list, e := client.RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				details := map[string]interface{}{
					"rules": i.Rules,
				}
				resources = append(resources, Resource{
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
		list, e := client.RbacV1().RoleBindings(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				details := map[string]interface{}{
					"subjects": i.Subjects,
					"roleRef":  i.RoleRef,
				}
				resources = append(resources, Resource{
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
		list, e := client.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				details := map[string]interface{}{
					"subjects": i.Subjects,
					"roleRef":  i.RoleRef,
				}
				resources = append(resources, Resource{
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
		list, e := client.NetworkingV1().NetworkPolicies(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: i.Namespace,
					Kind:      "NetworkPolicy",
					Status:    "Active", // NetworkPolicies don't have a status field like others
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
		list, e := client.CoreV1().PersistentVolumeClaims(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: i.Namespace,
					Kind:      "PersistentVolumeClaim",
					Status:    string(i.Status.Phase),
					Created:   i.CreationTimestamp.Format(time.RFC3339),
					UID:       string(i.UID),
					Details: map[string]interface{}{
						"accessModes":      i.Spec.AccessModes,
						"capacity":         i.Status.Capacity.Storage().String(),
						"storageClassName": i.Spec.StorageClassName,
						"volumeName":       i.Spec.VolumeName,
					},
				})
			}
		}
	case "PersistentVolume":
		// PVs are cluster-scoped, so we ignore namespace
		list, e := client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				resources = append(resources, Resource{
					Name:      i.Name,
					Namespace: "", // Cluster scoped
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
		list, e := client.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
		err = e
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
				resources = append(resources, Resource{
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
		list, e := client.CoreV1().ResourceQuotas(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				resources = append(resources, Resource{
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
		list, e := client.CoreV1().LimitRanges(listNamespace).List(context.TODO(), metav1.ListOptions{})
		err = e
		if err == nil {
			for _, i := range list.Items {
				resources = append(resources, Resource{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

func (h *Handlers) GetResourceYAML(w http.ResponseWriter, r *http.Request) {
	dynamicClient, err := h.getDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := namespace == "all"

	if kind == "" || name == "" {
		http.Error(w, "Missing kind or name parameter", http.StatusBadRequest)
		return
	}

	if err := validateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if namespace != "" && namespace != "all" {
		if err := validateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Try to get metadata; use namespaced from query param if provided (for CRDs)
	namespacedParam := r.URL.Query().Get("namespaced")
	meta, ok := resourceMeta[kind]
	if !ok {
		// Default to namespaced=true unless explicitly told otherwise
		isNamespaced := true
		if namespacedParam == "false" {
			isNamespaced = false
		}
		meta = ResourceMeta{Namespaced: isNamespaced}
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
		// Fall back to static resolution for known types
		gvr, _ = resolveGVR(kind)
	}

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

	obj, err := res.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch resource: %v", err), http.StatusInternalServerError)
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

func (h *Handlers) UpdateResourceYAML(w http.ResponseWriter, r *http.Request) {
	dynamicClient, err := h.getDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")

	if kind == "" {
		http.Error(w, "Missing kind parameter", http.StatusBadRequest)
		return
	}

	// Try to get metadata; use namespaced from query param if provided (for CRDs)
	namespacedParam := r.URL.Query().Get("namespaced")
	meta, ok := resourceMeta[kind]
	if !ok {
		// Default to namespaced=true unless explicitly told otherwise
		isNamespaced := true
		if namespacedParam == "false" {
			isNamespaced = false
		}
		meta = ResourceMeta{Namespaced: isNamespaced}
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

	// Extract GVR from the object's apiVersion and kind (not from query param)
	// This allows us to support CRDs and any custom resources
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
	gvr, ok := resolveGVR(objKind)
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

	_, err = res.Update(context.TODO(), &obj, metav1.UpdateOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update resource: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"updated"}`))
}

func (h *Handlers) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
	dynamicClient, err := h.getDynamicClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Limit body size to 1MB to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read body or body too large: %v", err), http.StatusBadRequest)
		return
	}

	dec := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(body), 4096)
	var applied []string

	for {
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

		// Validate that the resource is allowed
		if !isResourceAllowed(obj) {
			http.Error(w, fmt.Sprintf("Resource type %s is not allowed", obj.GetKind()), http.StatusForbidden)
			return
		}

		// Validate namespace (prevent creation in system namespaces)
		if isSystemNamespace(obj.GetNamespace()) {
			http.Error(w, "Cannot create resources in system namespaces", http.StatusForbidden)
			return
		}

		gvr, ok := resolveGVR(kind)
		if !ok {
			http.Error(w, fmt.Sprintf("Unsupported kind: %s", kind), http.StatusBadRequest)
			return
		}
		meta := resourceMeta[kind]

		// Namespace defaults
		if meta.Namespaced {
			if obj.GetNamespace() == "" {
				obj.SetNamespace("default")
			}
		} else {
			obj.SetNamespace("")
		}

		// Clean noisy metadata fields
		unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
		unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
		unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
		unstructured.RemoveNestedField(obj.Object, "metadata", "uid")

		ctx := context.TODO()
		var res dynamic.ResourceInterface
		if meta.Namespaced {
			res = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())
		} else {
			res = dynamicClient.Resource(gvr)
		}

		_, err = res.Create(ctx, obj, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			existing, gerr := res.Get(ctx, obj.GetName(), metav1.GetOptions{})
			if gerr != nil {
				http.Error(w, fmt.Sprintf("Failed to fetch existing %s: %v", kind, gerr), http.StatusInternalServerError)
				return
			}
			obj.SetResourceVersion(existing.GetResourceVersion())
			if _, uerr := res.Update(ctx, obj, metav1.UpdateOptions{}); uerr != nil {
				http.Error(w, fmt.Sprintf("Failed to update %s/%s: %v", kind, obj.GetName(), uerr), http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create %s/%s: %v", kind, obj.GetName(), err), http.StatusInternalServerError)
			return
		}

		nsPart := obj.GetNamespace()
		if nsPart == "" {
			nsPart = "-"
		}
		applied = append(applied, fmt.Sprintf("%s/%s/%s", kind, nsPart, obj.GetName()))
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

func isResourceAllowed(obj *unstructured.Unstructured) bool {
	// Whitelist of allowed resources
	allowedKinds := map[string]bool{
		"ConfigMap":               true,
		"Secret":                  true,
		"Deployment":              true,
		"Service":                 true,
		"Ingress":                 true,
		"PersistentVolumeClaim":   true,
		"Job":                     true,
		"CronJob":                 true,
		"StatefulSet":             true,
		"DaemonSet":               true,
		"HorizontalPodAutoscaler": true,
		"NetworkPolicy":           true,
		"ServiceAccount":          true,
		"Role":                    true,
		"RoleBinding":             true,
	}
	return allowedKinds[obj.GetKind()]
}

func isSystemNamespace(ns string) bool {
	systemNamespaces := map[string]bool{
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
	}
	return systemNamespaces[ns]
}

func (h *Handlers) DeleteResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dynamicClient, err := h.getDynamicClient(r)
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

	if err := validateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if namespace != "" && namespace != "all" {
		if err := validateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	meta, ok := resourceMeta[kind]
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

	gvr, _ := resolveGVR(kind)
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

	if err := res.Delete(context.TODO(), name, delOpts); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete %s/%s: %v", kind, name, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"deleted"}`))
}

func (h *Handlers) ScaleResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, err := h.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	deltaStr := r.URL.Query().Get("delta")

	if kind != "Deployment" {
		http.Error(w, "Scaling supported only for Deployments", http.StatusBadRequest)
		return
	}
	if name == "" {
		http.Error(w, "Missing name", http.StatusBadRequest)
		return
	}

	if err := validateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if namespace != "" {
		if err := validateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if namespace == "" {
		namespace = "default"
	}

	delta, err := strconv.Atoi(deltaStr)
	if err != nil || delta == 0 {
		http.Error(w, "Invalid delta", http.StatusBadRequest)
		return
	}

	ctx := context.TODO()
	scale, err := client.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get scale: %v", err), http.StatusInternalServerError)
		return
	}

	newReplicas := int(scale.Spec.Replicas) + delta
	if newReplicas < 0 {
		newReplicas = 0
	}
	scale.Spec.Replicas = int32(newReplicas)

	if _, err := client.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update scale: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"replicas":%d}`, newReplicas)))
}

func (h *Handlers) WatchResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dynamicClient, err := h.getDynamicClient(r)
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

	meta, ok := resourceMeta[kind]
	if !ok {
		http.Error(w, "Unsupported kind", http.StatusBadRequest)
		return
	}

	gvr, _ := resolveGVR(kind)
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

	watcher, err := res.Watch(context.TODO(), metav1.ListOptions{
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

type APIResourceInfo struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Resource   string `json:"resource"`
	Kind       string `json:"kind"`
	Namespaced bool   `json:"namespaced"`
}

type APIResourceObject struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status,omitempty"`
	Created   string      `json:"created,omitempty"`
	Raw       interface{} `json:"raw,omitempty"`
}

func (h *Handlers) ListAPIResources(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
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

func (h *Handlers) ListAPIResourceObjects(w http.ResponseWriter, r *http.Request) {
	dynamicClient, err := h.getDynamicClient(r)
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

	list, err := res.List(context.TODO(), metav1.ListOptions{})
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

func (h *Handlers) GetAPIResourceYAML(w http.ResponseWriter, r *http.Request) {
	dynamicClient, err := h.getDynamicClient(r)
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

	obj, err := res.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch resource: %v", err), http.StatusInternalServerError)
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

type ClusterStats struct {
	Nodes       int `json:"nodes"`
	Namespaces  int `json:"namespaces"`
	Pods        int `json:"pods"`
	Deployments int `json:"deployments"`
	Services    int `json:"services"`
	Ingresses   int `json:"ingresses"`
	PVCs        int `json:"pvcs"`
	PVs         int `json:"pvs"`
}

func (h *Handlers) GetClusterStats(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.TODO()
	stats := ClusterStats{}

	// Nodes
	if nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err == nil {
		stats.Nodes = len(nodes.Items)
	}

	// Namespaces
	if namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err == nil {
		stats.Namespaces = len(namespaces.Items)
	}

	// Pods
	if pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.Pods = len(pods.Items)
	}

	// Deployments
	if deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.Deployments = len(deployments.Items)
	}

	// Services
	if services, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.Services = len(services.Items)
	}

	// Ingresses
	if ingresses, err := client.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.Ingresses = len(ingresses.Items)
	}

	// PVCs
	if pvcs, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.PVCs = len(pvcs.Items)
	}

	// PVs
	if pvs, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{}); err == nil {
		stats.PVs = len(pvs.Items)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handlers) StreamPodLogs(w http.ResponseWriter, r *http.Request) {
	if _, err := authenticateRequest(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client, err := h.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if ns == "" || podName == "" {
		http.Error(w, "Missing namespace or pod parameter", http.StatusBadRequest)
		return
	}

	if err := validateK8sName(ns, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateK8sName(podName, "pod"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if container != "" {
		if err := validateK8sName(container, "container"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	opts := &corev1.PodLogOptions{
		Follow: true,
	}
	if container != "" {
		opts.Container = container
	}

	req := client.CoreV1().Pods(ns).GetLogs(podName, opts)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open log stream: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			if _, wErr := w.Write(buf[:n]); wErr != nil {
				break
			}
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}

func (h *Handlers) ExecIntoPod(w http.ResponseWriter, r *http.Request) {
	if _, err := authenticateRequest(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client, err := h.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if ns == "" || podName == "" {
		http.Error(w, "Missing namespace or pod parameter", http.StatusBadRequest)
		return
	}

	if err := validateK8sName(ns, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateK8sName(podName, "pod"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if container != "" {
		if err := validateK8sName(container, "container"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return false // Do not allow empty origin
			}
			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}

			// Check allowed origins from env
			allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
			if allowedOrigins != "" {
				origins := strings.Split(allowedOrigins, ",")
				for _, allowed := range origins {
					allowed = strings.TrimSpace(allowed)
					allowedURL, err := url.Parse(allowed)
					if err != nil {
						continue
					}
					if originURL.Scheme == allowedURL.Scheme && originURL.Host == allowedURL.Host {
						return true
					}
				}
				return false
			}

			// If no ALLOWED_ORIGINS, allow same-origin, localhost
			host := r.Host
			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			return originURL.Host == host || originURL.Host == "localhost" || originURL.Host == "127.0.0.1"
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upgrade to WebSocket: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Create exec request
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(ns).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, runtime.NewParameterCodec(clientgoscheme.Scheme))

	// Get REST config for exec
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.mu.RLock()
	restConfig := h.RESTConfigs[cluster]
	h.mu.RUnlock()

	if restConfig == nil {
		conn.WriteMessage(websocket.TextMessage, []byte("REST config not found for cluster"))
		return
	}

	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to create executor: %v", err)))
		return
	}

	// Create pipes for stdin/stdout/stderr
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	// Handle WebSocket messages (send to pod stdin)
	go func() {
		defer stdinWriter.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			stdinWriter.Write(message)
		}
	}()

	// Read from pod stdout and send to WebSocket
	go func() {
		defer stdoutReader.Close()
		buf := make([]byte, 8192)
		for {
			n, err := stdoutReader.Read(buf)
			if n > 0 {
				if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Execute the command
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdinReader,
		Stdout: stdoutWriter,
		Stderr: stdoutWriter,
		Tty:    true,
	})

	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\nExec error: %v\r\n", err)))
	}
}

const DataDir = "./data"

func (h *Handlers) GetLogo(w http.ResponseWriter, r *http.Request) {
	// Check for supported extensions
	extensions := []string{".png", ".svg"}
	var foundPath string

	for _, ext := range extensions {
		path := filepath.Join(DataDir, "logo"+ext)
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		absPath, _ := filepath.Abs(filepath.Join(DataDir, "logo.png")) // Default for logging
		fmt.Printf("Logo file not found (checked .png and .svg) in: %s\n", filepath.Dir(absPath))
		http.Error(w, "Logo not found", http.StatusNotFound)
		return
	}

	absPath, _ := filepath.Abs(foundPath)
	fmt.Printf("Serving logo from: %s\n", absPath)
	http.ServeFile(w, r, foundPath)
}

func (h *Handlers) UploadLogo(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to 5MB
	r.ParseMultipartForm(5 << 20)

	file, handler, err := r.FormFile("logo")
	if err != nil {
		fmt.Printf("Error retrieving file: %v\n", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if handler.Size > 5<<20 {
		http.Error(w, "File too large (max 5MB)", http.StatusBadRequest)
		return
	}

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	file.Seek(0, 0) // Reset pointer

	contentType := http.DetectContentType(buffer)
	// allowedTypes map removed as it was unused. We validate specifically below.

	// Validate extension
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".png" && ext != ".svg" {
		http.Error(w, "Invalid file type. Only .png and .svg are allowed", http.StatusBadRequest)
		return
	}

	if ext == ".png" && contentType != "image/png" {
		http.Error(w, "Invalid file content (not a PNG)", http.StatusBadRequest)
		return
	}
	// For SVG, we might want to check if it looks like XML/SVG, but DetectContentType is limited.
	// We'll trust the extension + size limit for SVG for now to avoid breaking valid SVGs,
	// but in a real high-sec env we'd parse the XML.

	// Ensure data directory exists
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		http.Error(w, "Error creating data directory", http.StatusInternalServerError)
		return
	}

	// Remove existing logos to avoid conflicts
	os.Remove(filepath.Join(DataDir, "logo.png"))
	os.Remove(filepath.Join(DataDir, "logo.svg"))

	// Create destination file
	destPath := filepath.Join(DataDir, "logo"+ext)
	absPath, _ := filepath.Abs(destPath)
	fmt.Printf("Saving logo to: %s\n", absPath)

	dst, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("Error creating destination file: %v\n", err)
		http.Error(w, "Error creating destination file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		fmt.Printf("Error saving file content: %v\n", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logo uploaded successfully"))
}

// GetCRDs returns a list of all Custom Resource Definitions in the cluster
func (h *Handlers) GetCRDs(w http.ResponseWriter, r *http.Request) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	// Get dynamic client for actual CRD objects
	dynamic, err := h.getDynamicClient(r)
	if err != nil {
		http.Error(w, "Dynamic client not found", http.StatusInternalServerError)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	unstructuredList, err := dynamic.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type CRD struct {
		Name    string `json:"name"`
		Group   string `json:"group"`
		Version string `json:"version"`
		Kind    string `json:"kind"`
		Scope   string `json:"scope"`
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

// GetCRDResources returns instances of a specific CRD
func (h *Handlers) GetCRDResources(w http.ResponseWriter, r *http.Request) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")
	namespaced := r.URL.Query().Get("namespaced") == "true"

	if resource == "" || version == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	dynamic, err := h.getDynamicClient(r)
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

	if namespaced && namespace != "" {
		unstructuredList, err = dynamic.Resource(gvr).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	} else {
		unstructuredList, err = dynamic.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type CRInstance struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
		Created   string `json:"created"`
		Status    string `json:"status,omitempty"`
	}

	instances := []CRInstance{}
	for _, item := range unstructuredList.Items {
		created := item.GetCreationTimestamp().Format(time.RFC3339)

		// Try to extract status if available
		status := ""
		if statusObj, found := item.Object["status"]; found {
			if statusMap, ok := statusObj.(map[string]interface{}); ok {
				if phase, ok := statusMap["phase"].(string); ok {
					status = phase
				} else if state, ok := statusMap["state"].(string); ok {
					status = state
				}
			}
		}

		instances = append(instances, CRInstance{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Created:   created,
			Status:    status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

// GetCRDYaml returns the YAML for a specific custom resource instance
func (h *Handlers) GetCRDYaml(w http.ResponseWriter, r *http.Request) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

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
	if err := validateK8sName(name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Only validate namespace if it's provided and relevant
	if namespaced && namespace != "" {
		if err := validateK8sName(namespace, "namespace"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	dynamic, err := h.getDynamicClient(r)
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
		obj, err = dynamic.Resource(gvr).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	} else {
		obj, err = dynamic.Resource(gvr).Get(context.Background(), name, metav1.GetOptions{})
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

func (h *Handlers) TriggerCronJob(w http.ResponseWriter, r *http.Request) {
	if _, err := authenticateRequest(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client, err := h.getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cronJob, err := client.BatchV1().CronJobs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobName := fmt.Sprintf("%s-manual-%d", req.Name, time.Now().Unix())
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: req.Namespace,
			Annotations: map[string]string{
				"cronjob.kubernetes.io/instantiate": "manual",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cronJob, schema.GroupVersionKind{
					Group:   "batch",
					Version: "v1",
					Kind:    "CronJob",
				}),
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}

	_, err = client.BatchV1().Jobs(req.Namespace).Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"jobName": jobName})
}

func validateK8sName(name, paramName string) error {
	if name == "" {
		return fmt.Errorf("%s is required", paramName)
	}

	// Validate length
	if len(name) > 253 {
		return fmt.Errorf("invalid %s: too long (max 253 characters)", paramName)
	}

	// Validate according to RFC 1123 (Kubernetes names) using regex
	// Lowercase alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.
	var dns1123SubdomainRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	if !dns1123SubdomainRegexp.MatchString(name) {
		return fmt.Errorf("invalid %s: must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character", paramName)
	}

	return nil
}
