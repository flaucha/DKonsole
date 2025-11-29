package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// ImportService provides business logic for importing Kubernetes resources
type ImportService struct {
	resourceRepo ResourceRepository
	gvrResolver  GVRResolver
	k8sClient    kubernetes.Interface
}

// NewImportService creates a new ImportService
func NewImportService(resourceRepo ResourceRepository, gvrResolver GVRResolver, k8sClient kubernetes.Interface) *ImportService {
	return &ImportService{
		resourceRepo: resourceRepo,
		gvrResolver:  gvrResolver,
		k8sClient:    k8sClient,
	}
}

// ImportResourceRequest represents parameters for importing resources
type ImportResourceRequest struct {
	YAMLContent []byte
}

// ImportResourceResponse represents the result of an import operation
type ImportResourceResponse struct {
	Status  string   `json:"status"`
	Count   int      `json:"count"`
	Applied []string `json:"resources"`
}

// ImportResources parses multi-document YAML and applies resources
func (s *ImportService) ImportResources(ctx context.Context, req ImportResourceRequest) (*ImportResourceResponse, error) {
	dec := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(req.YAMLContent), 4096)
	var applied []string
	resourceCount := 0

	for {
		if resourceCount >= 50 {
			return nil, fmt.Errorf("too many resources (max 50 per request)")
		}

		var objMap map[string]interface{}
		if err := dec.Decode(&objMap); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}
		if len(objMap) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{Object: objMap}
		kind := obj.GetKind()
		if kind == "" {
			return nil, fmt.Errorf("resource kind missing")
		}

		// Validate namespace (prevent creation in system namespaces)
		if utils.IsSystemNamespace(obj.GetNamespace()) {
			return nil, fmt.Errorf("cannot create resources in system namespace %s", obj.GetNamespace())
		}

		// Resolve GVR
		gvr, meta, err := s.gvrResolver.ResolveGVR(ctx, kind, obj.GetAPIVersion(), "")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve GVR for kind %s: %w", kind, err)
		}

		// Namespace defaults and permission validation
		if meta.Namespaced {
			if obj.GetNamespace() == "" {
				obj.SetNamespace("default")
			}
			// Validate edit permission for namespaced resources
			canEdit, err := permissions.CanPerformAction(ctx, obj.GetNamespace(), "edit")
			if err != nil {
				return nil, fmt.Errorf("failed to check permissions for namespace %s: %w", obj.GetNamespace(), err)
			}
			if !canEdit {
				return nil, fmt.Errorf("edit permission required for namespace: %s", obj.GetNamespace())
			}
		} else {
			obj.SetNamespace("")
			// Cluster-scoped resources require admin access
			// Check if user is core admin (pass nil for LDAP checker, will only check core admin)
			var ldapChecker permissions.LDAPAdminChecker = nil
			isAdmin, err := permissions.IsAdmin(ctx, ldapChecker)
			if err != nil {
				return nil, fmt.Errorf("failed to check admin status: %w", err)
			}
			if !isAdmin {
				return nil, fmt.Errorf("admin access required for cluster-scoped resources")
			}
		}

		// Cleanup metadata
		unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
		unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
		unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
		unstructured.RemoveNestedField(obj.Object, "metadata", "uid")

		// Marshal for patch
		patchData, err := json.Marshal(obj.Object)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal patch: %w", err)
		}

		// Use Server-Side Apply
		force := true
		patchOptions := metav1.PatchOptions{
			FieldManager: "dkonsole",
			Force:        &force,
		}

		_, err = s.resourceRepo.Patch(ctx, gvr, obj.GetName(), obj.GetNamespace(), meta.Namespaced, patchData, types.ApplyPatchType, patchOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to apply %s %s/%s: %w", kind, obj.GetNamespace(), obj.GetName(), err)
		}

		nsPart := obj.GetNamespace()
		if nsPart == "" {
			nsPart = "-"
		}
		applied = append(applied, fmt.Sprintf("%s/%s/%s", kind, nsPart, obj.GetName()))
		resourceCount++
	}

	if len(applied) == 0 {
		return nil, fmt.Errorf("no resources found in YAML")
	}

	return &ImportResourceResponse{
		Status:  "applied",
		Count:   len(applied),
		Applied: applied,
	}, nil
}
