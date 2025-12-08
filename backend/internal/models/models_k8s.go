package models

import "k8s.io/apimachinery/pkg/runtime/schema"

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

// DeploymentDetails contiene detalles específicos de un Deployment
type DeploymentDetails struct {
	Replicas    int32             `json:"replicas"`
	Ready       int32             `json:"ready"`
	Images      []string          `json:"images"`
	ImageTag    string            `json:"imageTag,omitempty"`
	Ports       []int32           `json:"ports"`
	PVCs        []string          `json:"pvcs"`
	PodLabels   map[string]string `json:"podLabels"`
	Labels      map[string]string `json:"labels,omitempty"`
	RequestsCPU string            `json:"requestsCPU,omitempty"`
	RequestsMem string            `json:"requestsMem,omitempty"`
	LimitsCPU   string            `json:"limitsCPU,omitempty"`
	LimitsMem   string            `json:"limitsMem,omitempty"`
}

// ResourceMeta contiene metadatos sobre un tipo de recurso de Kubernetes
type ResourceMeta struct {
	Group      string
	Version    string
	Resource   string
	Namespaced bool
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
