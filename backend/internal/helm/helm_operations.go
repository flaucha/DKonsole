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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/example/k8s-view/internal/utils"
)

// UpgradeHelmRelease upgrades a Helm release to a new version
func (s *Service) UpgradeHelmRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Chart      string                 `json:"chart,omitempty"`
		Version    string                 `json:"version,omitempty"`
		Repo       string                 `json:"repo,omitempty"`
		Values     map[string]interface{} `json:"values,omitempty"`
		ValuesYAML string                 `json:"valuesYaml,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Namespace == "" {
		http.Error(w, "Missing name or namespace", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(req.Name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(req.Namespace, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get REST config for the cluster
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	s.handlers.RLock()
	restConfig := s.handlers.RESTConfigs[cluster]
	s.handlers.RUnlock()

	if restConfig == nil {
		http.Error(w, "REST config not found for cluster", http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If chart not specified, get it from the existing release
	chartName := req.Chart
	var existingRepo string

	if chartName == "" || req.Repo == "" {
		secrets, err := client.CoreV1().Secrets(req.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("owner=helm,name=%s", req.Name),
		})
		if err == nil && len(secrets.Items) > 0 {
			var latestSecret *corev1.Secret
			latestRevision := 0
			for i := range secrets.Items {
				secret := &secrets.Items[i]
				if revStr, ok := secret.Labels["version"]; ok {
					if rev, err := strconv.Atoi(revStr); err == nil && rev > latestRevision {
						latestRevision = rev
						latestSecret = secret
					}
				}
			}

			if latestSecret != nil {
				if releaseData, ok := latestSecret.Data["release"]; ok {
					decoded := releaseData
					if decodedStr, err := base64.StdEncoding.DecodeString(string(releaseData)); err == nil && len(decodedStr) > 0 {
						decoded = decodedStr
					}

					reader := bytes.NewReader(decoded)
					if gzReader, err := gzip.NewReader(reader); err == nil {
						if decompressed, err := io.ReadAll(gzReader); err == nil {
							gzReader.Close()
							decoded = decompressed
						}
					}

					var releaseInfo map[string]interface{}
					if err := json.Unmarshal(decoded, &releaseInfo); err == nil {
						if chart, ok := releaseInfo["chart"].(map[string]interface{}); ok {
							if metadata, ok := chart["metadata"].(map[string]interface{}); ok {
								if name, ok := metadata["name"].(string); ok && chartName == "" {
									chartName = name
								}
								if repo, ok := metadata["repository"].(string); ok {
									if repo != "" {
										existingRepo = repo
									}
								}
							}
							if sources, ok := chart["sources"].([]interface{}); ok && existingRepo == "" {
								for _, source := range sources {
									if sourceStr, ok := source.(string); ok && sourceStr != "" {
										if strings.HasPrefix(sourceStr, "http://") || strings.HasPrefix(sourceStr, "https://") {
											existingRepo = sourceStr
											break
										}
									}
								}
							}
						}
					}
				}

				if chartName == "" {
					chartName = latestSecret.Labels["name"]
				}
			}
		}

		if chartName == "" {
			http.Error(w, "Chart name is required for upgrade. Could not determine chart from existing release.", http.StatusBadRequest)
			return
		}
	}

	if existingRepo != "" {
		req.Repo = existingRepo
	}

	// Use the ServiceAccount from dkonsole namespace
	dkonsoleNamespace := "dkonsole"
	saName := "dkonsole"

	_, err = client.CoreV1().ServiceAccounts(dkonsoleNamespace).Get(ctx, saName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Warning: ServiceAccount %s/%s not found, trying 'default': %v", dkonsoleNamespace, saName, err)
		saName = "default"
		_, err = client.CoreV1().ServiceAccounts(dkonsoleNamespace).Get(ctx, saName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error: Could not find ServiceAccount in namespace %s: %v", dkonsoleNamespace, err)
			http.Error(w, fmt.Sprintf("ServiceAccount not found in namespace %s", dkonsoleNamespace), http.StatusInternalServerError)
			return
		}
	}

	jobName := fmt.Sprintf("helm-upgrade-%s-%d", req.Name, time.Now().Unix())
	jobTimestamp := time.Now().Unix()

	valuesCMName := ""
	if req.ValuesYAML != "" {
		valuesCMName = fmt.Sprintf("helm-upgrade-%s-%d", req.Name, jobTimestamp)
		valuesCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      valuesCMName,
				Namespace: dkonsoleNamespace,
			},
			Data: map[string]string{
				"values.yaml": req.ValuesYAML,
			},
		}

		_, err = client.CoreV1().ConfigMaps(dkonsoleNamespace).Create(ctx, valuesCM, metav1.CreateOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create values ConfigMap: %v", err), http.StatusInternalServerError)
			return
		}
	}

	commonRepos := []struct {
		name string
		url  string
	}{
		{"prometheus-community", "https://prometheus-community.github.io/helm-charts"},
		{"bitnami", "https://charts.bitnami.com/bitnami"},
		{"ingress-nginx", "https://kubernetes.github.io/ingress-nginx"},
		{"stable", "https://charts.helm.sh/stable"},
	}

	var helmCmd []string
	var repoName string

	if req.Repo != "" {
		repoURL := strings.ToLower(req.Repo)

		if strings.Contains(repoURL, "prometheus-community") {
			repoName = "prometheus-community"
		} else if strings.Contains(repoURL, "bitnami") {
			repoName = "bitnami"
		} else if strings.Contains(repoURL, "kubernetes.github.io") {
			if strings.Contains(repoURL, "ingress-nginx") {
				repoName = "ingress-nginx"
			} else {
				parts := strings.Split(strings.Trim(req.Repo, "/"), "/")
				if len(parts) > 0 {
					repoName = parts[len(parts)-1]
				} else {
					repoName = "kubernetes"
				}
			}
		} else if strings.Contains(repoURL, "stable") {
			repoName = "stable"
		} else {
			repoName = strings.ReplaceAll(strings.ReplaceAll(req.Repo, "https://", ""), "http://", "")
			repoName = strings.ReplaceAll(repoName, ".github.io", "")
			parts := strings.Split(strings.Trim(repoName, "/"), "/")
			if len(parts) > 0 {
				repoName = parts[len(parts)-1]
			}
			repoName = strings.ReplaceAll(repoName, ".", "-")
			if len(repoName) > 50 {
				repoName = repoName[:50]
			}
			if repoName == "" {
				repoName = "temp-repo"
			}
		}

		upgradeCmd := fmt.Sprintf("helm upgrade %s %s/%s --namespace %s", req.Name, repoName, chartName, req.Namespace)

		if req.Version != "" {
			upgradeCmd += fmt.Sprintf(" --version %s", req.Version)
		}

		if req.ValuesYAML != "" && valuesCMName != "" {
			upgradeCmd += " -f /tmp/values/values.yaml"
		}

		cmdParts := []string{
			fmt.Sprintf("helm repo add %s %s 2>/dev/null || true", repoName, req.Repo),
			"helm repo update",
			upgradeCmd,
		}

		helmCmd = []string{"/bin/sh", "-c", strings.Join(cmdParts, " && ")}
	} else {
		repoAddCmds := []string{}
		for _, repo := range commonRepos {
			repoAddCmds = append(repoAddCmds, fmt.Sprintf("helm repo add %s %s 2>/dev/null || true", repo.name, repo.url))
		}

		searchCmd := fmt.Sprintf("CHART_REPO=$(helm search repo %s --output json 2>/dev/null | grep -o '\"name\":\"[^\"]*/%s\"' | head -1 | sed 's/\"name\":\"\\([^/]*\\)\\/.*/\\1/') || echo ''", chartName, chartName)

		upgradeCmd := fmt.Sprintf("if [ -n \"$CHART_REPO\" ] && [ \"$CHART_REPO\" != \"\" ]; then helm upgrade %s $CHART_REPO/%s --namespace %s", req.Name, chartName, req.Namespace)
		if req.Version != "" {
			upgradeCmd += fmt.Sprintf(" --version %s", req.Version)
		}
		if req.ValuesYAML != "" && valuesCMName != "" {
			upgradeCmd += " -f /tmp/values/values.yaml"
		}
		upgradeCmd += fmt.Sprintf("; else helm upgrade %s %s --namespace %s", req.Name, chartName, req.Namespace)
		if req.Version != "" {
			upgradeCmd += fmt.Sprintf(" --version %s", req.Version)
		}
		if req.ValuesYAML != "" && valuesCMName != "" {
			upgradeCmd += " -f /tmp/values/values.yaml"
		}
		upgradeCmd += "; fi"

		cmdParts := []string{
			strings.Join(repoAddCmds, " && "),
			"helm repo update",
			searchCmd,
			upgradeCmd,
		}

		helmCmd = []string{"/bin/sh", "-c", strings.Join(cmdParts, " && ")}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: dkonsoleNamespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: func() *int32 { t := int32(300); return &t }(),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: saName,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "helm",
							Image: "alpine/helm:latest",
						},
					},
				},
			},
		},
	}

	if len(helmCmd) == 3 && helmCmd[0] == "/bin/sh" && helmCmd[1] == "-c" {
		job.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh", "-c"}
		job.Spec.Template.Spec.Containers[0].Args = []string{helmCmd[2]}
	} else {
		job.Spec.Template.Spec.Containers[0].Command = helmCmd[0:1]
		job.Spec.Template.Spec.Containers[0].Args = helmCmd[1:]
	}

	if req.ValuesYAML != "" && valuesCMName != "" {
		job.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "values",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: valuesCMName,
						},
					},
				},
			},
		}
		job.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "values",
				MountPath: "/tmp/values",
			},
		}
	}

	_, err = client.BatchV1().Jobs(dkonsoleNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create upgrade job: %v", err), http.StatusInternalServerError)
		return
	}

	utils.AuditLog(r, "upgrade", "HelmRelease", req.Name, req.Namespace, true, nil, map[string]interface{}{
		"chart":   req.Chart,
		"version": req.Version,
		"job":     jobName,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "upgrade_initiated",
		"message": fmt.Sprintf("Helm upgrade job created: %s", jobName),
		"job":     jobName,
	})
}

// InstallHelmRelease installs a new Helm chart
func (s *Service) InstallHelmRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Chart      string                 `json:"chart"`
		Version    string                 `json:"version,omitempty"`
		Repo       string                 `json:"repo,omitempty"`
		Values     map[string]interface{} `json:"values,omitempty"`
		ValuesYAML string                 `json:"valuesYaml,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Namespace == "" || req.Chart == "" {
		http.Error(w, "Missing required fields: name, namespace, or chart", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(req.Name, "name"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(req.Namespace, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	s.handlers.RLock()
	restConfig := s.handlers.RESTConfigs[cluster]
	s.handlers.RUnlock()

	if restConfig == nil {
		http.Error(w, "REST config not found for cluster", http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chartName := req.Chart

	dkonsoleNamespace := "dkonsole"
	saName := "dkonsole"

	_, err = client.CoreV1().ServiceAccounts(dkonsoleNamespace).Get(ctx, saName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Warning: ServiceAccount %s/%s not found, trying 'default': %v", dkonsoleNamespace, saName, err)
		saName = "default"
		_, err = client.CoreV1().ServiceAccounts(dkonsoleNamespace).Get(ctx, saName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error: Could not find ServiceAccount in namespace %s: %v", dkonsoleNamespace, err)
			http.Error(w, fmt.Sprintf("ServiceAccount not found in namespace %s", dkonsoleNamespace), http.StatusInternalServerError)
			return
		}
	}

	jobName := fmt.Sprintf("helm-install-%s-%d", req.Name, time.Now().Unix())
	jobTimestamp := time.Now().Unix()

	valuesCMName := ""
	if req.ValuesYAML != "" {
		valuesCMName = fmt.Sprintf("helm-install-%s-%d", req.Name, jobTimestamp)
		valuesCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      valuesCMName,
				Namespace: dkonsoleNamespace,
			},
			Data: map[string]string{
				"values.yaml": req.ValuesYAML,
			},
		}

		_, err = client.CoreV1().ConfigMaps(dkonsoleNamespace).Create(ctx, valuesCM, metav1.CreateOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create values ConfigMap: %v", err), http.StatusInternalServerError)
			return
		}
	}

	commonRepos := []struct {
		name string
		url  string
	}{
		{"prometheus-community", "https://prometheus-community.github.io/helm-charts"},
		{"bitnami", "https://charts.bitnami.com/bitnami"},
		{"ingress-nginx", "https://kubernetes.github.io/ingress-nginx"},
		{"stable", "https://charts.helm.sh/stable"},
	}

	var helmCmd []string
	var repoName string

	if req.Repo != "" {
		repoURL := strings.ToLower(req.Repo)

		if strings.Contains(repoURL, "prometheus-community") {
			repoName = "prometheus-community"
		} else if strings.Contains(repoURL, "bitnami") {
			repoName = "bitnami"
		} else if strings.Contains(repoURL, "kubernetes.github.io") {
			if strings.Contains(repoURL, "ingress-nginx") {
				repoName = "ingress-nginx"
			} else {
				parts := strings.Split(strings.Trim(req.Repo, "/"), "/")
				if len(parts) > 0 {
					repoName = parts[len(parts)-1]
				} else {
					repoName = "kubernetes"
				}
			}
		} else if strings.Contains(repoURL, "stable") {
			repoName = "stable"
		} else {
			repoName = strings.ReplaceAll(strings.ReplaceAll(req.Repo, "https://", ""), "http://", "")
			repoName = strings.ReplaceAll(repoName, ".github.io", "")
			parts := strings.Split(strings.Trim(repoName, "/"), "/")
			if len(parts) > 0 {
				repoName = parts[len(parts)-1]
			}
			repoName = strings.ReplaceAll(repoName, ".", "-")
			if len(repoName) > 50 {
				repoName = repoName[:50]
			}
			if repoName == "" {
				repoName = "temp-repo"
			}
		}

		installCmd := fmt.Sprintf("helm install %s %s/%s --namespace %s --create-namespace", req.Name, repoName, chartName, req.Namespace)

		if req.Version != "" {
			installCmd += fmt.Sprintf(" --version %s", req.Version)
		}

		if req.ValuesYAML != "" && valuesCMName != "" {
			installCmd += " -f /tmp/values/values.yaml"
		}

		cmdParts := []string{
			fmt.Sprintf("helm repo add %s %s 2>/dev/null || true", repoName, req.Repo),
			"helm repo update",
			installCmd,
		}

		helmCmd = []string{"/bin/sh", "-c", strings.Join(cmdParts, " && ")}
	} else {
		repoAddCmds := []string{}
		for _, repo := range commonRepos {
			repoAddCmds = append(repoAddCmds, fmt.Sprintf("helm repo add %s %s 2>/dev/null || true", repo.name, repo.url))
		}

		searchCmd := fmt.Sprintf("CHART_REPO=$(helm search repo %s --output json 2>/dev/null | grep -o '\"name\":\"[^\"]*/%s\"' | head -1 | sed 's/\"name\":\"\\([^/]*\\)\\/.*/\\1/') || echo ''", chartName, chartName)

		installCmd := fmt.Sprintf("if [ -n \"$CHART_REPO\" ] && [ \"$CHART_REPO\" != \"\" ]; then helm install %s $CHART_REPO/%s --namespace %s --create-namespace", req.Name, chartName, req.Namespace)
		if req.Version != "" {
			installCmd += fmt.Sprintf(" --version %s", req.Version)
		}
		if req.ValuesYAML != "" && valuesCMName != "" {
			installCmd += " -f /tmp/values/values.yaml"
		}
		installCmd += fmt.Sprintf("; else helm install %s %s --namespace %s --create-namespace", req.Name, chartName, req.Namespace)
		if req.Version != "" {
			installCmd += fmt.Sprintf(" --version %s", req.Version)
		}
		if req.ValuesYAML != "" && valuesCMName != "" {
			installCmd += " -f /tmp/values/values.yaml"
		}
		installCmd += "; fi"

		cmdParts := []string{
			strings.Join(repoAddCmds, " && "),
			"helm repo update",
			searchCmd,
			installCmd,
		}

		helmCmd = []string{"/bin/sh", "-c", strings.Join(cmdParts, " && ")}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: dkonsoleNamespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: func() *int32 { t := int32(300); return &t }(),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: saName,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "helm",
							Image: "alpine/helm:latest",
						},
					},
				},
			},
		},
	}

	if len(helmCmd) == 3 && helmCmd[0] == "/bin/sh" && helmCmd[1] == "-c" {
		job.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh", "-c"}
		job.Spec.Template.Spec.Containers[0].Args = []string{helmCmd[2]}
	} else {
		job.Spec.Template.Spec.Containers[0].Command = helmCmd[0:1]
		job.Spec.Template.Spec.Containers[0].Args = helmCmd[1:]
	}

	if req.ValuesYAML != "" && valuesCMName != "" {
		job.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "values",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: valuesCMName,
						},
					},
				},
			},
		}
		job.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "values",
				MountPath: "/tmp/values",
			},
		}
	}

	_, err = client.BatchV1().Jobs(dkonsoleNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create install job: %v", err), http.StatusInternalServerError)
		return
	}

	utils.AuditLog(r, "install", "HelmRelease", req.Name, req.Namespace, true, nil, map[string]interface{}{
		"chart":   req.Chart,
		"version": req.Version,
		"job":     jobName,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "install_initiated",
		"message": fmt.Sprintf("Helm install job created: %s", jobName),
		"job":     jobName,
	})
}

