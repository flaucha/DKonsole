package models

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Handlers es la estructura principal que contiene los clients de Kubernetes
type Handlers struct {
	Clients       map[string]*kubernetes.Clientset
	Dynamics      map[string]dynamic.Interface
	Metrics       map[string]*metricsv.Clientset
	RESTConfigs   map[string]*rest.Config
	PrometheusURL string
	mu            sync.RWMutex
}

// Lock locks the handlers mutex for writing
func (h *Handlers) Lock() {
	h.mu.Lock()
}

// Unlock unlocks the handlers mutex
func (h *Handlers) Unlock() {
	h.mu.Unlock()
}

// RLock locks the handlers mutex for reading
func (h *Handlers) RLock() {
	h.mu.RLock()
}

// RUnlock unlocks the handlers mutex for reading
func (h *Handlers) RUnlock() {
	h.mu.RUnlock()
}

// ClusterConfig representa la configuración de un cluster de Kubernetes
type ClusterConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Token    string `json:"token"`
	Insecure bool   `json:"insecure"`
}

// Namespace representa un namespace de Kubernetes
type Namespace struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Labels  map[string]string `json:"labels"`
	Created string            `json:"created"`
}

// Resource representa un recurso genérico de Kubernetes
type Resource struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status"`
	Created   string      `json:"created"`
	UID       string      `json:"uid"`
	Details   interface{} `json:"details,omitempty"`
}

// PaginatedResources representa una respuesta paginada de recursos
type PaginatedResources struct {
	Resources []Resource `json:"resources"`
	Continue  string     `json:"continue,omitempty"`  // Token para la siguiente página
	Remaining int        `json:"remaining,omitempty"` // Cantidad estimada de recursos restantes
}

// DeploymentDetails contiene detalles específicos de un Deployment
type DeploymentDetails struct {
	Replicas  int32             `json:"replicas"`
	Ready     int32             `json:"ready"`
	Images    []string          `json:"images"`
	Ports     []int32           `json:"ports"`
	PVCs      []string          `json:"pvcs"`
	PodLabels map[string]string `json:"podLabels"`
}

// PodMetric representa métricas de CPU y memoria de un Pod
type PodMetric struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// ResourceMeta contiene metadatos sobre un tipo de recurso de Kubernetes
type ResourceMeta struct {
	Group      string
	Version    string
	Resource   string
	Namespaced bool
}

// APIResourceInfo contiene información sobre un recurso de API descubierto
type APIResourceInfo struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Resource   string `json:"resource"`
	Kind       string `json:"kind"`
	Namespaced bool   `json:"namespaced"`
}

// APIResourceObject representa un objeto de un recurso de API
type APIResourceObject struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status,omitempty"`
	Created   string      `json:"created,omitempty"`
	Raw       interface{} `json:"raw,omitempty"`
}

// ClusterStats contiene estadísticas agregadas del cluster
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

// HelmRelease representa un release de Helm
type HelmRelease struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Chart       string `json:"chart"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	Revision    int    `json:"revision"`
	Updated     string `json:"updated"`
	AppVersion  string `json:"appVersion,omitempty"`
	Description string `json:"description,omitempty"`
}

// Credentials representa las credenciales de autenticación
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IDP      string `json:"idp,omitempty"` // "core" or "ldap" - optional, defaults to trying both
}

// Claims representa los claims del JWT
type Claims struct {
	Username    string            `json:"username"`
	Role        string            `json:"role"`
	Permissions map[string]string `json:"permissions,omitempty"` // namespace -> permission (view/edit)
	RegisteredClaims interface{} `json:"-"` // Se manejará con jwt.RegisteredClaims en el paquete auth
}

// PrometheusQueryResult representa el resultado de una consulta a Prometheus
type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// MetricDataPoint representa un punto de datos de métrica con timestamp y valor
type MetricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// DeploymentMetricsResponse contiene métricas de CPU y memoria de un Deployment
type DeploymentMetricsResponse struct {
	CPU    []MetricDataPoint `json:"cpu"`
	Memory []MetricDataPoint `json:"memory"`
}

// PodMetricsResponse incluye todas las métricas de un Pod
type PodMetricsResponse struct {
	CPU       []MetricDataPoint `json:"cpu"`
	Memory    []MetricDataPoint `json:"memory"`
	NetworkRx []MetricDataPoint `json:"networkRx"`
	NetworkTx []MetricDataPoint `json:"networkTx"`
	PVCUsage  []MetricDataPoint `json:"pvcUsage"`
}

// ClusterOverviewResponse incluye métricas a nivel de cluster
type ClusterOverviewResponse struct {
	NodeMetrics  []NodeMetric            `json:"nodeMetrics"`
	ClusterStats *PrometheusClusterStats `json:"clusterStats"`
}

// NodeMetric representa métricas para un nodo individual
type NodeMetric struct {
	Name      string  `json:"name"`
	CPUUsage  float64 `json:"cpuUsage"`
	MemUsage  float64 `json:"memoryUsage"`
	DiskUsage float64 `json:"diskUsage"`
	NetworkRx float64 `json:"networkRx"`
	NetworkTx float64 `json:"networkTx"`
	Status    string  `json:"status"`
}

// PrometheusClusterStats representa estadísticas agregadas del cluster desde Prometheus
type PrometheusClusterStats struct {
	TotalNodes     int     `json:"totalNodes"`
	AvgCPUUsage    float64 `json:"avgCpuUsage"`
	AvgMemoryUsage float64 `json:"avgMemoryUsage"`
	NetworkTraffic float64 `json:"networkTraffic"`
	CPUTrend       float64 `json:"cpuTrend"`
	MemoryTrend    float64 `json:"memoryTrend"`
}

// ResourceMetaMap contiene el mapeo de tipos de recursos a sus metadatos
var ResourceMetaMap = map[string]ResourceMeta{
	"Deployment":              {Group: "apps", Version: "v1", Resource: "deployments", Namespaced: true},
	"Node":                    {Group: "", Version: "v1", Resource: "nodes", Namespaced: false},
	"Namespace":               {Group: "", Version: "v1", Resource: "namespaces", Namespaced: false},
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

// KindAliases mapea alias comunes a nombres completos de recursos
var KindAliases = map[string]string{
	"HPA": "HorizontalPodAutoscaler",
	"PVC": "PersistentVolumeClaim",
	"PV":  "PersistentVolume",
	"SC":  "StorageClass",
	"SA":  "ServiceAccount",
	"CR":  "ClusterRole",
	"CRB": "ClusterRoleBinding",
	"RB":  "RoleBinding",
}

// ResolveGVR resuelve un tipo de recurso a su GroupVersionResource
func ResolveGVR(kind string) (schema.GroupVersionResource, bool) {
	meta, ok := ResourceMetaMap[kind]
	if !ok {
		return schema.GroupVersionResource{}, false
	}
	return schema.GroupVersionResource{
		Group:    meta.Group,
		Version:  meta.Version,
		Resource: meta.Resource,
	}, true
}

// NormalizeKind normaliza un alias de tipo de recurso a su nombre completo
func NormalizeKind(kind string) string {
	if alias, ok := KindAliases[kind]; ok {
		return alias
	}
	return kind
}

// LDAPConfig representa la configuración del servidor LDAP
type LDAPConfig struct {
	Enabled      bool     `json:"enabled"`
	URL          string   `json:"url"`
	BaseDN       string   `json:"baseDN"`
	UserDN       string   `json:"userDN"`
	GroupDN      string   `json:"groupDN"`
	UserFilter   string   `json:"userFilter,omitempty"`
	RequiredGroup string   `json:"requiredGroup,omitempty"` // Grupo requerido para acceso (opcional)
	AdminGroups  []string `json:"adminGroups,omitempty"`    // Grupos LDAP que tienen acceso de admin al cluster
}

// LDAPGroupPermission representa los permisos de un grupo LDAP para un namespace
type LDAPGroupPermission struct {
	Namespace string `json:"namespace"`
	Permission string `json:"permission"` // "view", "edit"
}

// LDAPGroup representa un grupo LDAP con sus permisos
type LDAPGroup struct {
	Name        string                `json:"name"`
	Permissions []LDAPGroupPermission `json:"permissions"`
}

// LDAPGroupsConfig representa la configuración de grupos LDAP
type LDAPGroupsConfig struct {
	Groups []LDAPGroup `json:"groups"`
}
