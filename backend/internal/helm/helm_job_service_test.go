package helm

import (
	"context"
	"errors"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockHelmJobRepository is a mock implementation of HelmJobRepository
type mockHelmJobRepository struct {
	createConfigMapFunc   func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error
	createJobFunc         func(ctx context.Context, namespace string, job *batchv1.Job) error
	getServiceAccountFunc func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error)
}

func (m *mockHelmJobRepository) CreateConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
	if m.createConfigMapFunc != nil {
		return m.createConfigMapFunc(ctx, namespace, cm)
	}
	return nil
}

func (m *mockHelmJobRepository) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) error {
	if m.createJobFunc != nil {
		return m.createJobFunc(ctx, namespace, job)
	}
	return nil
}

func (m *mockHelmJobRepository) GetServiceAccount(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
	if m.getServiceAccountFunc != nil {
		return m.getServiceAccountFunc(ctx, namespace, name)
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}, nil
}

func TestHelmJobService_BuildHelmRepoName(t *testing.T) {
	service := NewHelmJobService(nil)

	tests := []struct {
		name    string
		repoURL string
		want    string
	}{
		{
			name:    "prometheus-community repo",
			repoURL: "https://prometheus-community.github.io/helm-charts",
			want:    "prometheus-community",
		},
		{
			name:    "bitnami repo",
			repoURL: "https://charts.bitnami.com/bitnami",
			want:    "bitnami",
		},
		{
			name:    "ingress-nginx repo",
			repoURL: "https://kubernetes.github.io/ingress-nginx",
			want:    "ingress-nginx",
		},
		{
			name:    "stable repo",
			repoURL: "https://charts.helm.sh/stable",
			want:    "stable",
		},
		{
			name:    "generic github repo",
			repoURL: "https://example.github.io/my-charts",
			want:    "my-charts",
		},
		{
			name:    "HTTP repo URL",
			repoURL: "http://example.com/charts",
			want:    "charts",
		},
		{
			name:    "custom domain repo",
			repoURL: "https://charts.example.com",
			want:    "charts-example-com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.BuildHelmRepoName(tt.repoURL)
			if result != tt.want {
				t.Errorf("BuildHelmRepoName() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestHelmJobService_CreateValuesConfigMap(t *testing.T) {
	tests := []struct {
		name                string
		namespace           string
		nameParam           string
		valuesYAML          string
		createConfigMapFunc func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error
		wantErr             bool
		errMsg              string
		expectedCMName      string
	}{
		{
			name:       "successful create ConfigMap with valid YAML",
			namespace:  "default",
			nameParam:  "my-release",
			valuesYAML: "key: value\nnested:\n  key: nested-value",
			createConfigMapFunc: func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
				if cm.Name == "" {
					t.Errorf("CreateValuesConfigMap() ConfigMap name is empty")
				}
				if namespace != "default" {
					t.Errorf("CreateValuesConfigMap() namespace = %v, want default", namespace)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:       "empty values YAML returns empty name",
			namespace:  "default",
			nameParam:  "my-release",
			valuesYAML: "",
			createConfigMapFunc: func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:       "error creating ConfigMap",
			namespace:  "default",
			nameParam:  "my-release",
			valuesYAML: "key: value",
			createConfigMapFunc: func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
				return errors.New("create error")
			},
			wantErr: true,
			errMsg:  "failed to create values configmap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockHelmJobRepository{
				createConfigMapFunc: tt.createConfigMapFunc,
			}

			service := NewHelmJobService(mockRepo)
			ctx := context.Background()

			cmName, err := service.CreateValuesConfigMap(ctx, tt.namespace, tt.nameParam, tt.valuesYAML)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateValuesConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateValuesConfigMap() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CreateValuesConfigMap() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if tt.valuesYAML != "" && cmName == "" {
				t.Errorf("CreateValuesConfigMap() ConfigMap name is empty for non-empty YAML")
			}
			if tt.valuesYAML == "" && cmName != "" {
				t.Errorf("CreateValuesConfigMap() ConfigMap name should be empty for empty YAML")
			}
		})
	}
}

func TestHelmJobService_CreateHelmJob(t *testing.T) {
	tests := []struct {
		name                  string
		request               CreateHelmJobRequest
		getServiceAccountFunc func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error)
		createJobFunc         func(ctx context.Context, namespace string, job *batchv1.Job) error
		wantErr               bool
		errMsg                string
	}{
		{
			name: "successful create helm job for install",
			request: CreateHelmJobRequest{
				Operation:          "install",
				ReleaseName:        "my-app",
				Namespace:          "default",
				ChartName:          "bitnami/nginx",
				DkonsoleNamespace:  "dkonsole",
				ServiceAccountName: "default",
			},
			getServiceAccountFunc: func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
				return &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}, nil
			},
			createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) error {
				if job.Name == "" {
					t.Errorf("CreateHelmJob() Job name is empty")
				}
				if namespace != "dkonsole" {
					t.Errorf("CreateHelmJob() namespace = %v, want dkonsole", namespace)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "service account not found, fallback to default",
			request: CreateHelmJobRequest{
				Operation:          "install",
				ReleaseName:        "my-app",
				Namespace:          "default",
				ChartName:          "bitnami/nginx",
				DkonsoleNamespace:  "dkonsole",
				ServiceAccountName: "custom-sa",
			},
			getServiceAccountFunc: func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
				if name == "custom-sa" {
					return nil, errors.New("not found")
				}
				return &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}, nil
			},
			createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "service account not found, default also not found",
			request: CreateHelmJobRequest{
				Operation:          "install",
				ReleaseName:        "my-app",
				Namespace:          "default",
				ChartName:          "bitnami/nginx",
				DkonsoleNamespace:  "dkonsole",
				ServiceAccountName: "custom-sa",
			},
			getServiceAccountFunc: func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
				return nil, errors.New("not found")
			},
			createJobFunc: nil,
			wantErr:       true,
			errMsg:        "serviceaccount not found",
		},
		{
			name: "error creating job",
			request: CreateHelmJobRequest{
				Operation:          "install",
				ReleaseName:        "my-app",
				Namespace:          "default",
				ChartName:          "bitnami/nginx",
				DkonsoleNamespace:  "dkonsole",
				ServiceAccountName: "default",
			},
			getServiceAccountFunc: func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
				return &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}, nil
			},
			createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) error {
				return errors.New("create job error")
			},
			wantErr: true,
			errMsg:  "failed to create helm job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockHelmJobRepository{
				getServiceAccountFunc: tt.getServiceAccountFunc,
				createJobFunc:         tt.createJobFunc,
			}

			service := NewHelmJobService(mockRepo)
			ctx := context.Background()

			jobName, err := service.CreateHelmJob(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHelmJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateHelmJob() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CreateHelmJob() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if jobName == "" {
				t.Errorf("CreateHelmJob() job name is empty")
			}

			// Verify job name format
			if !strings.HasPrefix(jobName, "helm-") {
				t.Errorf("CreateHelmJob() job name should start with 'helm-', got %v", jobName)
			}
		})
	}
}

func TestHelmJobService_BuildHelmCommand(t *testing.T) {
	service := NewHelmJobService(nil)

	cmd, err := service.BuildHelmCommand(HelmCommandRequest{
		Operation:    "install",
		ReleaseName:  "demo",
		Namespace:    "default",
		ChartName:    "nginx",
		Version:      "1.0.0",
		Repo:         "https://charts.example.com",
		ValuesYAML:   "key: val",
		ValuesCMName: "values-cm",
	})
	if err != nil {
		t.Fatalf("BuildHelmCommand returned error: %v", err)
	}

	if len(cmd) < 2 || cmd[0] != "helm" {
		t.Fatalf("expected helm command, got %v", cmd)
	}
	got := strings.Join(cmd, " ")
	if !strings.Contains(got, "install demo nginx --namespace default --create-namespace") {
		t.Fatalf("helm command missing install part: %s", got)
	}
	if !strings.Contains(got, "--repo https://charts.example.com") {
		t.Fatalf("helm command missing repo flag: %s", got)
	}
	if !strings.Contains(got, "--version 1.0.0") || !strings.Contains(got, "-f /tmp/values/values.yaml") {
		t.Fatalf("helm command missing flags: %s", got)
	}

	cmdNoRepo, err := service.BuildHelmCommand(HelmCommandRequest{
		Operation:   "upgrade",
		ReleaseName: "demo",
		Namespace:   "default",
		ChartName:   "bitnami/nginx",
	})
	if err != nil {
		t.Fatalf("BuildHelmCommand returned error: %v", err)
	}
	gotNoRepo := strings.Join(cmdNoRepo, " ")
	if !strings.Contains(gotNoRepo, "upgrade demo nginx --namespace default") || !strings.Contains(gotNoRepo, "--repo https://charts.bitnami.com/bitnami") {
		t.Fatalf("expected inferred repo URL and upgrade command, got: %s", gotNoRepo)
	}
}

func TestHelmJobService_BuildHelmCommand_RejectsMetacharacters(t *testing.T) {
	service := NewHelmJobService(nil)

	cases := []HelmCommandRequest{
		{
			Operation:   "install",
			ReleaseName: "demo",
			Namespace:   "default",
			ChartName:   "nginx;rm -rf /",
			Repo:        "https://charts.example.com",
		},
		{
			Operation:   "install",
			ReleaseName: "demo",
			Namespace:   "default",
			ChartName:   "nginx",
			Repo:        "https://charts.example.com\nbad",
		},
		{
			Operation:   "install",
			ReleaseName: "demo",
			Namespace:   "default",
			ChartName:   "nginx",
			Repo:        "https://charts.example.com",
			Version:     "$(id)",
		},
		{
			Operation:   "install",
			ReleaseName: "demo",
			Namespace:   "default",
			ChartName:   "-malicious",
			Repo:        "https://charts.example.com",
		},
	}

	for _, tc := range cases {
		_, err := service.BuildHelmCommand(tc)
		if err == nil {
			t.Fatalf("expected error for %+v", tc)
		}
	}
}

func TestHelmJobService_CreateHelmJob_AddsValuesVolume(t *testing.T) {
	var capturedJob *batchv1.Job
	mockRepo := &mockHelmJobRepository{
		createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) error {
			capturedJob = job
			return nil
		},
	}
	service := NewHelmJobService(mockRepo)

	_, err := service.CreateHelmJob(context.Background(), CreateHelmJobRequest{
		Operation:         "upgrade",
		ReleaseName:       "demo",
		Namespace:         "default",
		ChartName:         "bitnami/nginx",
		DkonsoleNamespace: "dkonsole",
		ValuesYAML:        "key: val",
		ValuesCMName:      "values-cm",
	})
	if err != nil {
		t.Fatalf("CreateHelmJob returned error: %v", err)
	}
	if capturedJob == nil {
		t.Fatalf("expected job to be created")
	}
	if len(capturedJob.Spec.Template.Spec.Volumes) == 0 {
		t.Fatalf("expected values volume to be mounted")
	}
	if capturedJob.Spec.Template.Spec.Containers[0].VolumeMounts == nil {
		t.Fatalf("expected volume mount to be set")
	}
	if got := capturedJob.Spec.Template.Spec.Containers[0].Command; len(got) != 1 || got[0] != "helm" {
		t.Fatalf("expected helm command, got %v", got)
	}
	if img := capturedJob.Spec.Template.Spec.Containers[0].Image; !strings.Contains(img, "@sha256:") {
		t.Fatalf("expected pinned image digest, got %q", img)
	}
}
