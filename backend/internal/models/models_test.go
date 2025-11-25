package models

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestHandlers_Lock(t *testing.T) {
	h := &Handlers{}
	h.Lock()
	// If we get here without deadlock, the test passes
	h.Unlock()
}

func TestHandlers_RLock(t *testing.T) {
	h := &Handlers{}
	h.RLock()
	// If we get here without deadlock, the test passes
	h.RUnlock()
}

func TestNormalizeKind(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"HPA alias", "HPA", "HorizontalPodAutoscaler"},
		{"PVC alias", "PVC", "PersistentVolumeClaim"},
		{"PV alias", "PV", "PersistentVolume"},
		{"SC alias", "SC", "StorageClass"},
		{"SA alias", "SA", "ServiceAccount"},
		{"CR alias", "CR", "ClusterRole"},
		{"CRB alias", "CRB", "ClusterRoleBinding"},
		{"RB alias", "RB", "RoleBinding"},
		{"No alias", "Pod", "Pod"},
		{"No alias Deployment", "Deployment", "Deployment"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeKind(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeKind(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolveGVR(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		wantGVR schema.GroupVersionResource
		wantOk  bool
	}{
		{"Pod", "Pod", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, true},
		{"Deployment", "Deployment", schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, true},
		{"Service", "Service", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}, true},
		{"ConfigMap", "ConfigMap", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, true},
		{"Secret", "Secret", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}, true},
		{"Unknown", "UnknownKind", schema.GroupVersionResource{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gvr, ok := ResolveGVR(tt.kind)
			if ok != tt.wantOk {
				t.Errorf("ResolveGVR() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if tt.wantOk {
				if gvr.Group != tt.wantGVR.Group || gvr.Version != tt.wantGVR.Version || gvr.Resource != tt.wantGVR.Resource {
					t.Errorf("ResolveGVR() = %v, want %v", gvr, tt.wantGVR)
				}
			}
		})
	}
}

func TestResourceMetaMap(t *testing.T) {
	// Test that ResourceMetaMap is populated
	if len(ResourceMetaMap) == 0 {
		t.Error("ResourceMetaMap should not be empty")
	}

	// Test some common resources
	commonResources := []string{"Pod", "Deployment", "Service", "ConfigMap", "Secret"}
	for _, resource := range commonResources {
		if meta, exists := ResourceMetaMap[resource]; !exists {
			t.Errorf("ResourceMetaMap should contain %s", resource)
		} else {
			if meta.Resource == "" {
				t.Errorf("ResourceMetaMap[%s].Resource should not be empty", resource)
			}
		}
	}
}

func TestKindAliases(t *testing.T) {
	// Test that KindAliases is populated
	if len(KindAliases) == 0 {
		t.Error("KindAliases should not be empty")
	}

	// Test some common aliases
	testCases := map[string]string{
		"HPA": "HorizontalPodAutoscaler",
		"PVC": "PersistentVolumeClaim",
		"PV":  "PersistentVolume",
		"SC":  "StorageClass",
		"SA":  "ServiceAccount",
		"CR":  "ClusterRole",
		"CRB": "ClusterRoleBinding",
		"RB":  "RoleBinding",
	}

	for alias, expectedKind := range testCases {
		if kind, exists := KindAliases[alias]; !exists {
			t.Errorf("KindAliases should contain alias %s", alias)
		} else {
			if kind != expectedKind {
				t.Errorf("KindAliases[%s] = %q, want %q", alias, kind, expectedKind)
			}
		}
	}
}

func TestClusterConfig(t *testing.T) {
	config := ClusterConfig{
		Name:     "test-cluster",
		Host:     "https://kubernetes.example.com",
		Token:    "test-token",
		Insecure: false,
	}

	if config.Name != "test-cluster" {
		t.Errorf("ClusterConfig.Name = %q, want %q", config.Name, "test-cluster")
	}
	if config.Host != "https://kubernetes.example.com" {
		t.Errorf("ClusterConfig.Host = %q, want %q", config.Host, "https://kubernetes.example.com")
	}
	if config.Token != "test-token" {
		t.Errorf("ClusterConfig.Token = %q, want %q", config.Token, "test-token")
	}
	if config.Insecure {
		t.Error("ClusterConfig.Insecure = true, want false")
	}
}

func TestNamespace(t *testing.T) {
	ns := Namespace{
		Name:    "default",
		Status:  "Active",
		Labels:  map[string]string{"env": "production"},
		Created: "2024-01-01T00:00:00Z",
	}

	if ns.Name != "default" {
		t.Errorf("Namespace.Name = %q, want %q", ns.Name, "default")
	}
	if ns.Status != "Active" {
		t.Errorf("Namespace.Status = %q, want %q", ns.Status, "Active")
	}
	if len(ns.Labels) != 1 {
		t.Errorf("Namespace.Labels length = %d, want 1", len(ns.Labels))
	}
}

func TestResource(t *testing.T) {
	resource := Resource{
		Name:      "test-pod",
		Namespace: "default",
		Kind:      "Pod",
		Status:    "Running",
		Created:   "2024-01-01T00:00:00Z",
		UID:       "test-uid",
	}

	if resource.Name != "test-pod" {
		t.Errorf("Resource.Name = %q, want %q", resource.Name, "test-pod")
	}
	if resource.Namespace != "default" {
		t.Errorf("Resource.Namespace = %q, want %q", resource.Namespace, "default")
	}
	if resource.Kind != "Pod" {
		t.Errorf("Resource.Kind = %q, want %q", resource.Kind, "Pod")
	}
	if resource.Status != "Running" {
		t.Errorf("Resource.Status = %q, want %q", resource.Status, "Running")
	}
}

func TestDeploymentDetails(t *testing.T) {
	details := DeploymentDetails{
		Replicas:  3,
		Ready:     2,
		Images:    []string{"nginx:latest"},
		Ports:     []int32{80, 443},
		PVCs:      []string{"pvc1"},
		PodLabels: map[string]string{"app": "nginx"},
	}

	if details.Replicas != 3 {
		t.Errorf("DeploymentDetails.Replicas = %d, want 3", details.Replicas)
	}
	if details.Ready != 2 {
		t.Errorf("DeploymentDetails.Ready = %d, want 2", details.Ready)
	}
	if len(details.Images) != 1 {
		t.Errorf("DeploymentDetails.Images length = %d, want 1", len(details.Images))
	}
}

func TestClusterStats(t *testing.T) {
	stats := ClusterStats{
		Nodes:       3,
		Namespaces:  5,
		Pods:        10,
		Deployments: 4,
		Services:    6,
		Ingresses:   2,
		PVCs:        8,
		PVs:         5,
	}

	if stats.Nodes != 3 {
		t.Errorf("ClusterStats.Nodes = %d, want 3", stats.Nodes)
	}
	if stats.Namespaces != 5 {
		t.Errorf("ClusterStats.Namespaces = %d, want 5", stats.Namespaces)
	}
	if stats.Pods != 10 {
		t.Errorf("ClusterStats.Pods = %d, want 10", stats.Pods)
	}
}

func TestCredentials(t *testing.T) {
	creds := Credentials{
		Username: "testuser",
		Password: "testpass",
	}

	if creds.Username != "testuser" {
		t.Errorf("Credentials.Username = %q, want %q", creds.Username, "testuser")
	}
	if creds.Password != "testpass" {
		t.Errorf("Credentials.Password = %q, want %q", creds.Password, "testpass")
	}
}

func TestClaims(t *testing.T) {
	claims := Claims{
		Username: "testuser",
		Role:     "admin",
	}

	if claims.Username != "testuser" {
		t.Errorf("Claims.Username = %q, want %q", claims.Username, "testuser")
	}
	if claims.Role != "admin" {
		t.Errorf("Claims.Role = %q, want %q", claims.Role, "admin")
	}
}
