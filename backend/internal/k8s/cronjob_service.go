package k8s

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// CronJobRepository defines the interface for CronJob operations
type CronJobRepository interface {
	GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error)
	CreateJob(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error)
}

// K8sCronJobRepository implements CronJobRepository
type K8sCronJobRepository struct {
	client kubernetes.Interface
}

// NewK8sCronJobRepository creates a new K8sCronJobRepository
func NewK8sCronJobRepository(client kubernetes.Interface) *K8sCronJobRepository {
	return &K8sCronJobRepository{client: client}
}

// GetCronJob gets a CronJob
func (r *K8sCronJobRepository) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	cronJob, err := r.client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}
	return cronJob, nil
}

// CreateJob creates a Job
func (r *K8sCronJobRepository) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	createdJob, err := r.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	return createdJob, nil
}

// CronJobService provides business logic for CronJob operations
type CronJobService struct {
	repo CronJobRepository
}

// NewCronJobService creates a new CronJobService
func NewCronJobService(repo CronJobRepository) *CronJobService {
	return &CronJobService{repo: repo}
}

// TriggerCronJob manually triggers a CronJob by creating an immediate Job
func (s *CronJobService) TriggerCronJob(ctx context.Context, namespace, name string) (string, error) {
	cronJob, err := s.repo.GetCronJob(ctx, namespace, name)
	if err != nil {
		return "", fmt.Errorf("failed to get cronjob: %w", err)
	}

	jobName := fmt.Sprintf("%s-manual-%d", name, time.Now().Unix())
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Annotations: map[string]string{
				"cronjob.kubernetes.io/instantiate": "manual",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cronJob, schema.GroupVersionKind{
					Group:   "batch",
					Version: "v1",
					Kind:    "CronJob",
				}),
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}

	_, err = s.repo.CreateJob(ctx, namespace, job)
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	return jobName, nil
}
