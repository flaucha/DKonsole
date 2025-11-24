package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/utils"
)

// Service provides Helm release operations
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
}

// NewService creates a new Helm service
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
	}
}

// HelmRelease represents a Helm release
type HelmRelease struct {
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	Chart       string    `json:"chart"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	Revision    int       `json:"revision"`
	Updated     string    `json:"updated"`
	AppVersion  string    `json:"appVersion,omitempty"`
	Description string    `json:"description,omitempty"`
}

// GetHelmReleases lists all Helm releases in the cluster
func (s *Service) GetHelmReleases(w http.ResponseWriter, r *http.Request) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	releasesMap := make(map[string]*HelmRelease)

	// Helm 3 stores releases in Secrets (default) or ConfigMaps
	// Check Secrets across all namespaces
	secrets, err := client.CoreV1().Secrets("").List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err == nil {
		for _, secret := range secrets.Items {
			// Extract release name from labels or annotations
			releaseName := secret.Labels["name"]
			if releaseName == "" {
				releaseName = secret.Annotations["meta.helm.sh/release-name"]
			}
			releaseNamespace := secret.Namespace
			if releaseNamespace == "" {
				releaseNamespace = secret.Annotations["meta.helm.sh/release-namespace"]
			}

			if releaseName != "" && releaseNamespace != "" {
				key := fmt.Sprintf("%s/%s", releaseNamespace, releaseName)

				// Parse Helm release data from secret
				if releaseData, ok := secret.Data["release"]; ok {
					decoded := releaseData

					// Try to decode as base64 string first
					if decodedStr, err := base64.StdEncoding.DecodeString(string(releaseData)); err == nil && len(decodedStr) > 0 {
						decoded = decodedStr
					}

					// Try to decompress gzip
					reader := bytes.NewReader(decoded)
					gzReader, err := gzip.NewReader(reader)
					var releaseInfo map[string]interface{}

					if err == nil {
						decompressed, err := io.ReadAll(gzReader)
						gzReader.Close()
						if err == nil {
							decoded = decompressed
						}
					}

					// Try to parse as JSON
					if err := json.Unmarshal(decoded, &releaseInfo); err != nil {
						log.Printf("Failed to unmarshal release JSON for %s/%s: %v", releaseNamespace, releaseName, err)
						continue
					}

					chartName := ""
					chartVersion := ""
					appVersion := ""

					if chart, ok := releaseInfo["chart"].(map[string]interface{}); ok {
						if metadata, ok := chart["metadata"].(map[string]interface{}); ok {
							if name, ok := metadata["name"].(string); ok {
								chartName = name
							}
							if version, ok := metadata["version"].(string); ok {
								chartVersion = version
							}
							if av, ok := metadata["appVersion"].(string); ok {
								appVersion = av
							}
						}
					}

					// Get status and revision from labels first
					status := secret.Labels["status"]
					if status == "" {
						status = "unknown"
					}

					revision := 0
					if revStr, ok := secret.Labels["version"]; ok {
						if rev, err := strconv.Atoi(revStr); err == nil {
							revision = rev
						}
					}

					updated := ""
					description := ""

					if info, ok := releaseInfo["info"].(map[string]interface{}); ok {
						if status == "" || status == "superseded" {
							if s, ok := info["status"].(string); ok {
								status = s
							}
						}

						if revision == 0 {
							if r, ok := info["revision"].(float64); ok {
								revision = int(r)
							}
						}

						if u, ok := info["last_deployed"].(map[string]interface{}); ok {
							if uStr, ok := u["Time"].(string); ok {
								updated = uStr
							} else if uStr, ok := info["last_deployed"].(string); ok {
								updated = uStr
							}
						} else if uStr, ok := info["last_deployed"].(string); ok {
							updated = uStr
						}

						if d, ok := info["description"].(string); ok {
							description = d
						}
					}

					// Only update if this is a newer revision or doesn't exist
					if existing, exists := releasesMap[key]; !exists || revision > existing.Revision || (revision == existing.Revision && status == "deployed" && existing.Status != "deployed") {
						releasesMap[key] = &HelmRelease{
							Name:        releaseName,
							Namespace:   releaseNamespace,
							Chart:       chartName,
							Version:     chartVersion,
							Status:      status,
							Revision:    revision,
							Updated:     updated,
							AppVersion:  appVersion,
							Description: description,
						}
					}
				}
			}
		}
	}

	// Also check ConfigMaps (some Helm installations use ConfigMaps)
	configMaps, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err == nil {
		for _, cm := range configMaps.Items {
			releaseName := cm.Labels["name"]
			if releaseName == "" {
				releaseName = cm.Annotations["meta.helm.sh/release-name"]
			}
			releaseNamespace := cm.Namespace
			if releaseNamespace == "" {
				releaseNamespace = cm.Annotations["meta.helm.sh/release-namespace"]
			}

			if releaseName != "" && releaseNamespace != "" {
				key := fmt.Sprintf("%s/%s", releaseNamespace, releaseName)

				if _, exists := releasesMap[key]; exists {
					continue
				}

				chartName := cm.Labels["name"]
				chartVersion := cm.Labels["version"]

				status := cm.Labels["status"]
				if status == "" {
					status = "unknown"
				}

				revision := 0
				if revStr, ok := cm.Labels["version"]; ok {
					if rev, err := strconv.Atoi(revStr); err == nil {
						revision = rev
					}
				}

				releasesMap[key] = &HelmRelease{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Chart:     chartName,
					Version:   chartVersion,
					Status:    status,
					Revision:  revision,
					Updated:   cm.CreationTimestamp.Format(time.RFC3339),
				}
			}
		}
	}

	// Convert map to slice
	releases := make([]HelmRelease, 0, len(releasesMap))
	for _, release := range releasesMap {
		releases = append(releases, *release)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(releases)
}

// DeleteHelmRelease uninstalls a Helm release by deleting its Secrets
func (s *Service) DeleteHelmRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	releaseName := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")

	if releaseName == "" || namespace == "" {
		http.Error(w, "Missing name or namespace parameter", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(releaseName, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Find all Secrets related to this Helm release
	secrets, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list secrets: %v", err), http.StatusInternalServerError)
		return
	}

	deletedCount := 0
	for _, secret := range secrets.Items {
		releaseNameFromAnnotation := secret.Annotations["meta.helm.sh/release-name"]
		releaseNamespaceFromAnnotation := secret.Annotations["meta.helm.sh/release-namespace"]

		secretNameMatches := strings.HasPrefix(secret.Name, fmt.Sprintf("sh.helm.release.v1.%s.v", releaseName))

		if (releaseNameFromAnnotation == releaseName && releaseNamespaceFromAnnotation == namespace) || secretNameMatches {
			if err := client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Printf("Failed to delete Helm release secret %s: %v", secret.Name, err)
				}
			} else {
				deletedCount++
			}
		}
	}

	// Also check ConfigMaps
	configMaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, cm := range configMaps.Items {
			releaseNameFromAnnotation := cm.Annotations["meta.helm.sh/release-name"]
			releaseNamespaceFromAnnotation := cm.Annotations["meta.helm.sh/release-namespace"]

			if releaseNameFromAnnotation == releaseName && releaseNamespaceFromAnnotation == namespace {
				if err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, cm.Name, metav1.DeleteOptions{}); err != nil {
					if !apierrors.IsNotFound(err) {
						log.Printf("Failed to delete Helm release configmap %s: %v", cm.Name, err)
					}
				} else {
					deletedCount++
				}
			}
		}
	}

	if deletedCount == 0 {
		http.Error(w, "No Helm release secrets found", http.StatusNotFound)
		return
	}

	utils.AuditLog(r, "delete", "HelmRelease", releaseName, namespace, true, nil, map[string]interface{}{
		"secrets_deleted": deletedCount,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "deleted",
		"secrets_deleted": deletedCount,
	})
}
