package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

// DecodeHelmReleaseData decodes and decompresses Helm release data from a Secret (exported for reuse)
func (s *HelmReleaseService) DecodeHelmReleaseData(releaseData []byte) (map[string]interface{}, error) {
	decoded := releaseData

	// Try to decode as base64 string first
	if decodedStr, err := base64.StdEncoding.DecodeString(string(releaseData)); err == nil && len(decodedStr) > 0 {
		decoded = decodedStr
	}

	// Try to decompress gzip
	reader := bytes.NewReader(decoded)
	gzReader, err := gzip.NewReader(reader)
	if err == nil {
		decompressed, err := io.ReadAll(gzReader)
		gzReader.Close()
		if err == nil {
			decoded = decompressed
		}
	}

	// Parse as JSON
	var releaseInfo map[string]interface{}
	if err := json.Unmarshal(decoded, &releaseInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal release JSON: %w", err)
	}

	return releaseInfo, nil
}

// extractChartInfo extracts chart information from release data
func (s *HelmReleaseService) extractChartInfo(releaseInfo map[string]interface{}) (chartName, chartVersion, appVersion string) {
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
	return
}

// extractReleaseInfo extracts release information from release data and secret labels
func (s *HelmReleaseService) extractReleaseInfo(releaseInfo map[string]interface{}, secret corev1.Secret) (status string, revision int, updated, description string) {
	// Get status and revision from labels first
	status = secret.Labels["status"]
	if status == "" {
		status = "unknown"
	}

	if revStr, ok := secret.Labels["version"]; ok {
		if rev, err := strconv.Atoi(revStr); err == nil {
			revision = rev
		}
	}

	// Override from release info if available
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

	return status, revision, updated, description
}
