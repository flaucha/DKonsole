package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/example/k8s-view/internal/utils"
)

// TriggerCronJob triggers a CronJob manually
func (s *Service) TriggerCronJob(w http.ResponseWriter, r *http.Request) {
	// Note: authenticateRequest is handled by middleware
	// This will be updated when we fully migrate auth

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()
	cronJob, err := client.BatchV1().CronJobs(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobName := fmt.Sprintf("%s-manual-%d", req.Name, time.Now().Unix())
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: req.Namespace,
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

	ctx2, cancel2 := utils.CreateTimeoutContext()
	defer cancel2()
	_, err = client.BatchV1().Jobs(req.Namespace).Create(ctx2, job, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"jobName": jobName})
}

