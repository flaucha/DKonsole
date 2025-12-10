package pod

import (
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// execCreator abstracts exec creation to allow testing with mocks.
type execCreator interface {
	CreateExecutor(client kubernetes.Interface, config *rest.Config, req ExecRequest) (remotecommand.Executor, string, error)
}

// ClusterService defines the interface for cluster operations needed by the pod service.
type ClusterService interface {
	GetClient(r *http.Request) (kubernetes.Interface, error)
	GetRESTConfig(r *http.Request) (*rest.Config, error)
}

// Service provides HTTP handlers for pod-specific operations including log streaming and exec.
type Service struct {
	handlers       *models.Handlers
	clusterService ClusterService
	logRepoFactory func(kubernetes.Interface) LogRepository
	execFactory    func() execCreator
}

// NewService creates a new pod service with the provided handlers and cluster service.
func NewService(h *models.Handlers, cs ClusterService) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
		logRepoFactory: func(client kubernetes.Interface) LogRepository {
			return NewK8sLogRepository(client)
		},
		execFactory: func() execCreator {
			return NewExecService()
		},
		// PodService will be created on-demand in handlers that need it. Factories are overridable in tests.
	}
}
