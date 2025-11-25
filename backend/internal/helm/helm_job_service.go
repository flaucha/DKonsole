package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HelmJobRepository defines the interface for creating Kubernetes Jobs and ConfigMaps
type HelmJobRepository interface {
	CreateConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap) error
	CreateJob(ctx context.Context, namespace string, job *batchv1.Job) error
	GetServiceAccount(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error)
}

// K8sHelmJobRepository implements HelmJobRepository
type K8sHelmJobRepository struct {
	client kubernetes.Interface
}

// NewK8sHelmJobRepository creates a new K8sHelmJobRepository
func NewK8sHelmJobRepository(client kubernetes.Interface) *K8sHelmJobRepository {
	return &K8sHelmJobRepository{client: client}
}

// CreateConfigMap creates a ConfigMap
func (r *K8sHelmJobRepository) CreateConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
	_, err := r.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}
	return nil
}

// CreateJob creates a Job
func (r *K8sHelmJobRepository) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) error {
	_, err := r.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	return nil
}

// GetServiceAccount gets a ServiceAccount
func (r *K8sHelmJobRepository) GetServiceAccount(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
	sa, err := r.client.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get serviceaccount: %w", err)
	}
	return sa, nil
}

// HelmJobService provides business logic for creating Helm Jobs
type HelmJobService struct {
	repo HelmJobRepository
}

// NewHelmJobService creates a new HelmJobService
func NewHelmJobService(repo HelmJobRepository) *HelmJobService {
	return &HelmJobService{repo: repo}
}

// CommonHelmRepo represents a common Helm repository
type CommonHelmRepo struct {
	Name string
	URL  string
}

var commonRepos = []CommonHelmRepo{
	{"prometheus-community", "https://prometheus-community.github.io/helm-charts"},
	{"bitnami", "https://charts.bitnami.com/bitnami"},
	{"ingress-nginx", "https://kubernetes.github.io/ingress-nginx"},
	{"stable", "https://charts.helm.sh/stable"},
}

// BuildHelmRepoName builds a repository name from a URL
func (s *HelmJobService) BuildHelmRepoName(repoURL string) string {
	repoURLLower := strings.ToLower(repoURL)

	if strings.Contains(repoURLLower, "prometheus-community") {
		return "prometheus-community"
	} else if strings.Contains(repoURLLower, "bitnami") {
		return "bitnami"
	} else if strings.Contains(repoURLLower, "kubernetes.github.io") {
		if strings.Contains(repoURLLower, "ingress-nginx") {
			return "ingress-nginx"
		} else {
			parts := strings.Split(strings.Trim(repoURL, "/"), "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
			return "kubernetes"
		}
	} else if strings.Contains(repoURLLower, "stable") {
		return "stable"
	} else {
		repoName := strings.ReplaceAll(strings.ReplaceAll(repoURL, "https://", ""), "http://", "")
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
		return repoName
	}
}

// BuildHelmCommand builds a Helm command for install/upgrade
func (s *HelmJobService) BuildHelmCommand(req HelmCommandRequest) []string {
	var cmdParts []string

	if req.Repo != "" {
		repoName := s.BuildHelmRepoName(req.Repo)
		cmdParts = append(cmdParts, fmt.Sprintf("helm repo add %s %s 2>/dev/null || true", repoName, req.Repo))
		cmdParts = append(cmdParts, "helm repo update")

		helmCmd := fmt.Sprintf("helm %s %s %s/%s --namespace %s", req.Operation, req.ReleaseName, repoName, req.ChartName, req.Namespace)
		
		if req.Operation == "install" {
			helmCmd += " --create-namespace"
		}

		if req.Version != "" {
			helmCmd += fmt.Sprintf(" --version %s", req.Version)
		}

		if req.ValuesYAML != "" && req.ValuesCMName != "" {
			helmCmd += " -f /tmp/values/values.yaml"
		}

		cmdParts = append(cmdParts, helmCmd)
	} else {
		// Add common repos
		repoAddCmds := []string{}
		for _, repo := range commonRepos {
			repoAddCmds = append(repoAddCmds, fmt.Sprintf("helm repo add %s %s 2>/dev/null || true", repo.Name, repo.URL))
		}
		cmdParts = append(cmdParts, strings.Join(repoAddCmds, " && "))
		cmdParts = append(cmdParts, "helm repo update")

		// Search for chart in repos
		searchCmd := fmt.Sprintf("CHART_REPO=$(helm search repo %s --output json 2>/dev/null | grep -o '\"name\":\"[^\"]*/%s\"' | head -1 | sed 's/\"name\":\"\\([^/]*\\)\\/.*/\\1/') || echo ''", req.ChartName, req.ChartName)
		cmdParts = append(cmdParts, searchCmd)

		// Build command with fallback
		helmCmd := fmt.Sprintf("if [ -n \"$CHART_REPO\" ] && [ \"$CHART_REPO\" != \"\" ]; then helm %s %s $CHART_REPO/%s --namespace %s", req.Operation, req.ReleaseName, req.ChartName, req.Namespace)
		if req.Operation == "install" {
			helmCmd += " --create-namespace"
		}
		if req.Version != "" {
			helmCmd += fmt.Sprintf(" --version %s", req.Version)
		}
		if req.ValuesYAML != "" && req.ValuesCMName != "" {
			helmCmd += " -f /tmp/values/values.yaml"
		}
		helmCmd += fmt.Sprintf("; else helm %s %s %s --namespace %s", req.Operation, req.ReleaseName, req.ChartName, req.Namespace)
		if req.Operation == "install" {
			helmCmd += " --create-namespace"
		}
		if req.Version != "" {
			helmCmd += fmt.Sprintf(" --version %s", req.Version)
		}
		if req.ValuesYAML != "" && req.ValuesCMName != "" {
			helmCmd += " -f /tmp/values/values.yaml"
		}
		helmCmd += "; fi"
		cmdParts = append(cmdParts, helmCmd)
	}

	return []string{"/bin/sh", "-c", strings.Join(cmdParts, " && ")}
}

// HelmCommandRequest represents parameters for building a Helm command
type HelmCommandRequest struct {
	Operation    string // "install" or "upgrade"
	ReleaseName  string
	Namespace    string
	ChartName    string
	Version      string
	Repo         string
	ValuesYAML   string
	ValuesCMName string
}

// CreateValuesConfigMap creates a ConfigMap for Helm values
func (s *HelmJobService) CreateValuesConfigMap(ctx context.Context, namespace, name, valuesYAML string) (string, error) {
	if valuesYAML == "" {
		return "", nil
	}

	cmName := fmt.Sprintf("%s-%d", name, time.Now().Unix())
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"values.yaml": valuesYAML,
		},
	}

	if err := s.repo.CreateConfigMap(ctx, namespace, cm); err != nil {
		return "", fmt.Errorf("failed to create values configmap: %w", err)
	}

	return cmName, nil
}

// CreateHelmJob creates a Kubernetes Job for running Helm commands
func (s *HelmJobService) CreateHelmJob(ctx context.Context, req CreateHelmJobRequest) (string, error) {
	// Get or validate service account
	saName := req.ServiceAccountName
	dkonsoleNamespace := req.DkonsoleNamespace

	if saName != "" {
		_, err := s.repo.GetServiceAccount(ctx, dkonsoleNamespace, saName)
		if err != nil {
			// Try default
			_, err = s.repo.GetServiceAccount(ctx, dkonsoleNamespace, "default")
			if err != nil {
				return "", fmt.Errorf("serviceaccount not found in namespace %s", dkonsoleNamespace)
			}
			saName = "default"
		}
	}

	jobName := fmt.Sprintf("helm-%s-%s-%d", req.Operation, req.ReleaseName, time.Now().Unix())

	// Build Helm command
	helmCmd := s.BuildHelmCommand(HelmCommandRequest{
		Operation:    req.Operation,
		ReleaseName:  req.ReleaseName,
		Namespace:    req.Namespace,
		ChartName:    req.ChartName,
		Version:      req.Version,
		Repo:         req.Repo,
		ValuesYAML:   req.ValuesYAML,
		ValuesCMName: req.ValuesCMName,
	})

	// Create Job
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

	// Set command and args
	if len(helmCmd) == 3 && helmCmd[0] == "/bin/sh" && helmCmd[1] == "-c" {
		job.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh", "-c"}
		job.Spec.Template.Spec.Containers[0].Args = []string{helmCmd[2]}
	} else {
		if len(helmCmd) > 0 {
			job.Spec.Template.Spec.Containers[0].Command = helmCmd[0:1]
		}
		if len(helmCmd) > 1 {
			job.Spec.Template.Spec.Containers[0].Args = helmCmd[1:]
		}
	}

	// Add volume for values if needed
	if req.ValuesYAML != "" && req.ValuesCMName != "" {
		job.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "values",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: req.ValuesCMName,
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

	if err := s.repo.CreateJob(ctx, dkonsoleNamespace, job); err != nil {
		return "", fmt.Errorf("failed to create helm job: %w", err)
	}

	return jobName, nil
}

// CreateHelmJobRequest represents parameters for creating a Helm Job
type CreateHelmJobRequest struct {
	Operation          string // "install" or "upgrade"
	ReleaseName        string
	Namespace          string
	ChartName          string
	Version            string
	Repo               string
	ValuesYAML         string
	ValuesCMName       string
	ServiceAccountName string
	DkonsoleNamespace  string
}









