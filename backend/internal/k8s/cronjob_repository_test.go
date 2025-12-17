package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

type stubCronJobRepo struct {
	getCalled    bool
	createCalled bool
}

func (s *stubCronJobRepo) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	s.getCalled = true
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.CronJobSpec{},
	}, nil
}

func (s *stubCronJobRepo) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	s.createCalled = true
	return job, nil
}

type stubFactory struct {
	cronService *CronJobService
}

func (f *stubFactory) CreateResourceService(dynamic.Interface, kubernetes.Interface) *ResourceService {
	return nil
}
func (f *stubFactory) CreateImportService(dynamic.Interface, kubernetes.Interface) *ImportService {
	return nil
}
func (f *stubFactory) CreateNamespaceService(kubernetes.Interface) *NamespaceService { return nil }
func (f *stubFactory) CreateClusterStatsService(kubernetes.Interface) *ClusterStatsService {
	return nil
}
func (f *stubFactory) CreateDeploymentService(kubernetes.Interface) *DeploymentService { return nil }
func (f *stubFactory) CreateCronJobService(kubernetes.Interface) *CronJobService {
	return f.cronService
}
func (f *stubFactory) CreateWatchService() *WatchService { return nil }

func TestK8sCronJobRepository(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cron",
			Namespace: "ns",
		},
	})
	repo := NewK8sCronJobRepository(client)
	ctx := context.Background()

	cron, err := repo.GetCronJob(ctx, "ns", "my-cron")
	if err != nil {
		t.Fatalf("GetCronJob error: %v", err)
	}
	if cron.Name != "my-cron" {
		t.Fatalf("unexpected cronjob: %s", cron.Name)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "ns",
		},
	}
	created, err := repo.CreateJob(ctx, "ns", job)
	if err != nil {
		t.Fatalf("CreateJob error: %v", err)
	}
	if created.Name != "job" {
		t.Fatalf("unexpected job name: %s", created.Name)
	}
}

func TestService_TriggerCronJobHandler(t *testing.T) {
	repo := &stubCronJobRepo{}
	cronService := NewCronJobService(repo)
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{
			"default": k8sfake.NewSimpleClientset(),
		},
	}
	svc := &Service{
		handlers:       handlers,
		clusterService: cluster.NewService(handlers),
		serviceFactory: &stubFactory{cronService: cronService},
	}

	body := bytes.NewBufferString(`{"namespace":"default","name":"my-cron"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cron", body)

	// Add user to context
	claims := &auth.AuthClaims{
		Claims: models.Claims{
			Username: "admin",
			Role:     "admin",
		},
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey(), claims)
	req = req.WithContext(ctx)

	svc.TriggerCronJob(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["jobName"] == "" {
		t.Fatalf("expected jobName in response")
	}
	if !repo.getCalled || !repo.createCalled {
		t.Fatalf("expected repository methods to be called")
	}
}

func TestService_TriggerCronJobHandler_Errors(t *testing.T) {
	handlers := &models.Handlers{Clients: map[string]kubernetes.Interface{}}
	svc := &Service{
		handlers:       handlers,
		clusterService: cluster.NewService(handlers),
		serviceFactory: &stubFactory{},
	}

	t.Run("missing client", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cron", bytes.NewBufferString(`{"namespace":"ns","name":"cron"}`))

		svc.TriggerCronJob(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		handlers.Clients["default"] = k8sfake.NewSimpleClientset()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cron", bytes.NewBufferString("{"))

		svc.TriggerCronJob(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
	})
}
