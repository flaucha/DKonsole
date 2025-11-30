package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockCronJobRepository is a mock implementation of CronJobRepository
type mockCronJobRepository struct {
	getCronJobFunc func(ctx context.Context, namespace, name string) (*batchv1.CronJob, error)
	createJobFunc  func(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error)
}

func (m *mockCronJobRepository) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	if m.getCronJobFunc != nil {
		return m.getCronJobFunc(ctx, namespace, name)
	}
	return nil, errors.New("get cronjob not implemented")
}

func (m *mockCronJobRepository) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	if m.createJobFunc != nil {
		return m.createJobFunc(ctx, namespace, job)
	}
	return nil, errors.New("create job not implemented")
}

func TestCronJobService_TriggerCronJob(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		cronJobName    string
		getCronJobFunc func(ctx context.Context, namespace, name string) (*batchv1.CronJob, error)
		createJobFunc  func(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error)
		wantErr        bool
		wantJobName    string
		errMsg         string
	}{
		{
			name:        "successful trigger",
			namespace:   "default",
			cronJobName: "my-cronjob",
			getCronJobFunc: func(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
				return &batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: batchv1.CronJobSpec{
						JobTemplate: batchv1.JobTemplateSpec{
							Spec: batchv1.JobSpec{
								Template: corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name:  "test",
												Image: "busybox",
											},
										},
										RestartPolicy: corev1.RestartPolicyOnFailure,
									},
								},
							},
						},
					},
				}, nil
			},
			createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
				return job, nil
			},
			wantErr:     false,
			wantJobName: "my-cronjob-manual-", // Prefix - actual timestamp will vary
		},
		{
			name:        "cronjob not found",
			namespace:   "default",
			cronJobName: "non-existent",
			getCronJobFunc: func(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
				return nil, errors.New("cronjob not found")
			},
			wantErr: true,
			errMsg:  "failed to get cronjob",
		},
		{
			name:        "create job fails",
			namespace:   "default",
			cronJobName: "my-cronjob",
			getCronJobFunc: func(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
				return &batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: batchv1.CronJobSpec{
						JobTemplate: batchv1.JobTemplateSpec{
							Spec: batchv1.JobSpec{
								Template: corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{Name: "test", Image: "busybox"},
										},
										RestartPolicy: corev1.RestartPolicyOnFailure,
									},
								},
							},
						},
					},
				}, nil
			},
			createJobFunc: func(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
				return nil, errors.New("failed to create job")
			},
			wantErr: true,
			errMsg:  "failed to create job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCronJobRepository{
				getCronJobFunc: tt.getCronJobFunc,
				createJobFunc:  tt.createJobFunc,
			}

			service := NewCronJobService(mockRepo)
			ctx := context.Background()

			jobName, err := service.TriggerCronJob(ctx, tt.namespace, tt.cronJobName)

			if (err != nil) != tt.wantErr {
				t.Errorf("TriggerCronJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("TriggerCronJob() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("TriggerCronJob() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if jobName == "" {
				t.Errorf("TriggerCronJob() jobName is empty")
				return
			}

			if tt.wantJobName != "" {
				// Check prefix if specified (since timestamp varies)
				if len(jobName) < len(tt.wantJobName) {
					t.Errorf("TriggerCronJob() jobName = %v, want prefix %v", jobName, tt.wantJobName)
				} else if jobName[:len(tt.wantJobName)] != tt.wantJobName {
					t.Errorf("TriggerCronJob() jobName = %v, want prefix %v", jobName, tt.wantJobName)
				}
			}
		})
	}
}
