package helm

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

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

// BuildHelmCommand builds a Helm command for install/upgrade without invoking a shell.
func (s *HelmJobService) BuildHelmCommand(req HelmCommandRequest) ([]string, error) {
	if req.Operation != "install" && req.Operation != "upgrade" {
		return nil, fmt.Errorf("invalid operation: %s", req.Operation)
	}
	if err := utils.ValidateK8sName(req.ReleaseName, "releaseName"); err != nil {
		return nil, err
	}
	if err := utils.ValidateK8sName(req.Namespace, "namespace"); err != nil {
		return nil, err
	}
	if req.ChartName == "" {
		return nil, fmt.Errorf("chartName is required")
	}
	if containsForbiddenHelmChars(req.ChartName) || strings.HasPrefix(req.ChartName, "-") {
		return nil, fmt.Errorf("invalid chartName")
	}
	if req.Repo != "" {
		if containsForbiddenHelmChars(req.Repo) || strings.HasPrefix(req.Repo, "-") {
			return nil, fmt.Errorf("invalid repo")
		}
		u, err := url.Parse(req.Repo)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, fmt.Errorf("invalid repo URL")
		}
	}
	if req.Version != "" {
		if containsForbiddenHelmChars(req.Version) || strings.HasPrefix(req.Version, "-") {
			return nil, fmt.Errorf("invalid version")
		}
	}

	chartArg, repoURL, err := resolveChartAndRepo(req.ChartName, req.Repo)
	if err != nil {
		return nil, err
	}

	args := []string{req.Operation, req.ReleaseName, chartArg, "--namespace", req.Namespace}
	if req.Operation == "install" {
		args = append(args, "--create-namespace")
	}
	if req.Version != "" {
		args = append(args, "--version", req.Version)
	}
	if req.ValuesYAML != "" && req.ValuesCMName != "" {
		args = append(args, "-f", "/tmp/values/values.yaml")
	}
	if repoURL != "" && !isDirectChartRef(chartArg) {
		args = append(args, "--repo", repoURL)
	}

	return append([]string{"helm"}, args...), nil
}

func isDirectChartRef(chart string) bool {
	return strings.HasPrefix(chart, "oci://") || strings.HasPrefix(chart, "http://") || strings.HasPrefix(chart, "https://")
}

func resolveChartAndRepo(chartName, repo string) (chartArg string, repoURL string, err error) {
	if isDirectChartRef(chartName) {
		return chartName, "", nil
	}

	// When repo is provided, accept either "chart" or "repo/chart" and strip the prefix if present.
	if repo != "" {
		if parts := strings.Split(chartName, "/"); len(parts) == 2 {
			return parts[1], repo, nil
		}
		return chartName, repo, nil
	}

	// Without explicit repo URL, require a known repo prefix (e.g. "bitnami/nginx").
	parts := strings.Split(chartName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("repo is required unless chartName is a direct URL/OCI ref or uses a known repo prefix (e.g. bitnami/<chart>)")
	}
	repoName := parts[0]
	chart := parts[1]
	if chart == "" {
		return "", "", fmt.Errorf("invalid chartName")
	}

	for _, r := range commonRepos {
		if strings.EqualFold(r.Name, repoName) {
			return chart, r.URL, nil
		}
	}

	return "", "", fmt.Errorf("unknown chart repo prefix: %s (set repo URL explicitly)", repoName)
}

func containsForbiddenHelmChars(s string) bool {
	for _, r := range s {
		if r == '`' || r == ';' || r == '|' || r == '&' || r == '$' || r == '(' || r == ')' || r == '<' || r == '>' ||
			r == '\n' || r == '\r' || r == '\t' || unicode.IsSpace(r) || unicode.IsControl(r) {
			return true
		}
	}
	return false
}
