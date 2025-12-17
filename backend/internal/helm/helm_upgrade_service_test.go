package helm

import (
	"context"
	"errors"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// Using mockHelmJobRepository from helm_job_service_test.go (same package)
// Using MockHelmReleaseService from helm_test.go (same package)

func TestUpgradeHelmRelease(t *testing.T) {
	mockJobRepo := &mockHelmJobRepository{
		createConfigMapFunc: func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
			return nil
		},
		createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) error {
			return nil
		},
		getServiceAccountFunc: func(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
			return &corev1.ServiceAccount{}, nil
		},
	}

	mockReleaseService := &MockHelmReleaseService{
		GetChartInfoFunc: func(ctx context.Context, namespace, releaseName string) (*ChartInfo, error) {
			return &ChartInfo{}, nil
		},
	}

	jobService := NewHelmJobService(mockJobRepo)
	upgradeService := NewHelmUpgradeService(mockReleaseService, jobService)

	tests := []struct {
		name      string
		req       UpgradeHelmReleaseRequest
		wantErr   bool
		mockSetup func()
	}{
		{
			name: "Success with explicit chart",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx", // Explicit
				Repo:       "https://charts.bitnami.com/bitnami",
				DkonsoleNS: "dkonsole",
			},
			wantErr: false,
		},
		{
			name: "Success inferring chart from release",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "", // Missing, should infer
				DkonsoleNS: "dkonsole",
			},
			wantErr: false,
			mockSetup: func() {
				mockReleaseService.GetChartInfoFunc = func(ctx context.Context, ns, name string) (*ChartInfo, error) {
					return &ChartInfo{ChartName: "inferred-nginx", Repo: "https://charts.example.com"}, nil
				}
			},
		},
		{
			name: "Partial Inference (Chart explicit, Repo inferred)",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx", // Explicit
				Repo:       "",      // Missing
				DkonsoleNS: "dkonsole",
			},
			wantErr: false,
			mockSetup: func() {
				mockReleaseService.GetChartInfoFunc = func(ctx context.Context, ns, name string) (*ChartInfo, error) {
					return &ChartInfo{ChartName: "nginx", Repo: "https://charts.example.com"}, nil
				}
			},
		},
		{
			name: "Failure inferring chart (error getting info)",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
			mockSetup: func() {
				mockReleaseService.GetChartInfoFunc = func(ctx context.Context, ns, name string) (*ChartInfo, error) {
					return nil, errors.New("inference error")
				}
			},
		},
		{
			name: "Failure inferring chart (no chart name found)",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
			mockSetup: func() {
				mockReleaseService.GetChartInfoFunc = func(ctx context.Context, ns, name string) (*ChartInfo, error) {
					return &ChartInfo{}, nil // Empty
				}
			},
		},
		{
			name: "ConfigMap Creation Failure",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx",
				Repo:       "https://charts.bitnami.com/bitnami",
				ValuesYAML: "foo: bar",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
			mockSetup: func() {
				mockJobRepo.createConfigMapFunc = func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
					return errors.New("cm creation failed")
				}
			},
		},
		{
			name: "Job Creation Failure",
			req: UpgradeHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx",
				Repo:       "https://charts.bitnami.com/bitnami",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
			mockSetup: func() {
				mockJobRepo.createJobFunc = func(ctx context.Context, namespace string, job *batchv1.Job) error {
					return errors.New("job creation failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockReleaseService.GetChartInfoFunc = func(ctx context.Context, ns, name string) (*ChartInfo, error) {
				return &ChartInfo{}, nil
			}
			mockJobRepo.createConfigMapFunc = func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error { return nil }
			mockJobRepo.createJobFunc = func(ctx context.Context, namespace string, job *batchv1.Job) error { return nil }

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			_, err := upgradeService.UpgradeHelmRelease(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeHelmRelease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
