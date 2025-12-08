package helm

import (
	"fmt"
	"strings"
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
