package pod

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecService provides business logic for pod exec operations.
// It creates executors for executing commands in pod containers via the Kubernetes API.
type ExecService struct{}

// NewExecService creates a new ExecService instance.
func NewExecService() *ExecService {
	return &ExecService{}
}

// ExecRequest represents parameters for executing a command in a pod.
type ExecRequest struct {
	Namespace string // Kubernetes namespace
	PodName   string // Pod name
	Container string // Optional container name (for multi-container pods)
}

// CreateExecutor creates a remote command executor for a pod.
// The command executed is hardcoded to /bin/sh (or /bin/bash if available) for security.
// Returns the executor, the exec URL, and an error if creation fails.
func (s *ExecService) CreateExecutor(client kubernetes.Interface, config *rest.Config, req ExecRequest) (remotecommand.Executor, string, error) {
	// Build exec request
	execReq := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(req.PodName).
		Namespace(req.Namespace).
		SubResource("exec")

	// Build command for interactive terminal
	command := []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh"}

	execReq.VersionedParams(&corev1.PodExecOptions{
		Container: req.Container,
		Command:   command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, runtime.NewParameterCodec(clientgoscheme.Scheme))

	// Get URL for executor
	url := execReq.URL()

	// Create executor
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create executor: %w", err)
	}

	return executor, url.String(), nil
}
