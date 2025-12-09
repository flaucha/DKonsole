package helm

import (
	"context"
	"errors"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// Using mockHelmJobRepository from helm_job_service_test.go (same package)

func TestInstallHelmRelease(t *testing.T) {
	mockRepo := &mockHelmJobRepository{
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

	// Create HelmJobService with mock repo
	jobService := NewHelmJobService(mockRepo)
	// Create HelmInstallService with job service
	installService := NewHelmInstallService(jobService)

	tests := []struct {
		name      string
		req       InstallHelmReleaseRequest
		wantErr   bool
		mockSetup func()
	}{
		{
			name: "Success",
			req: InstallHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx",
				DkonsoleNS: "dkonsole",
			},
			wantErr: false,
		},
		{
			name: "Missing Chart",
			req: InstallHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
		},
		{
			name: "Job Creation Failure",
			req: InstallHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
			mockSetup: func() {
				mockRepo.createJobFunc = func(ctx context.Context, namespace string, job *batchv1.Job) error {
					return errors.New("job creation failed")
				}
			},
		},
		{
			name: "ConfigMap Creation Failure",
			req: InstallHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx",
				ValuesYAML: "foo: bar",
				DkonsoleNS: "dkonsole",
			},
			wantErr: true,
			mockSetup: func() {
				mockRepo.createConfigMapFunc = func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
					return errors.New("cm creation failed")
				}
			},
		},
		{
			name: "Success with Values",
			req: InstallHelmReleaseRequest{
				Name:       "test-release",
				Namespace:  "default",
				Chart:      "nginx",
				ValuesYAML: "foo: bar",
				DkonsoleNS: "dkonsole",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.createConfigMapFunc = func(ctx context.Context, namespace string, cm *corev1.ConfigMap) error { return nil }
			mockRepo.createJobFunc = func(ctx context.Context, namespace string, job *batchv1.Job) error { return nil }

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			_, err := installService.InstallHelmRelease(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("InstallHelmRelease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
