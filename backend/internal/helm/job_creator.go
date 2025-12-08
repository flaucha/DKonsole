package helm

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateHelmJobRequest represents parameters for creating a Helm Job
type CreateHelmJobRequest struct {
	Operation          string // "install" or "upgrade"
	ReleaseName        string
	Namespace          string
	ChartName          string
	Version            string
	Repo               string
	ValuesYAML         string
	ValuesCMName       string
	ServiceAccountName string
	DkonsoleNamespace  string
}

// CreateValuesConfigMap creates a ConfigMap for Helm values
func (s *HelmJobService) CreateValuesConfigMap(ctx context.Context, namespace, name, valuesYAML string) (string, error) {
	if valuesYAML == "" {
		return "", nil
	}

	cmName := fmt.Sprintf("%s-%d", name, time.Now().Unix())
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"values.yaml": valuesYAML,
		},
	}

	if err := s.repo.CreateConfigMap(ctx, namespace, cm); err != nil {
		return "", fmt.Errorf("failed to create values configmap: %w", err)
	}

	return cmName, nil
}

// CreateHelmJob creates a Kubernetes Job for running Helm commands
func (s *HelmJobService) CreateHelmJob(ctx context.Context, req CreateHelmJobRequest) (string, error) {
	// Get or validate service account
	saName := req.ServiceAccountName
	dkonsoleNamespace := req.DkonsoleNamespace

	if saName != "" {
		_, err := s.repo.GetServiceAccount(ctx, dkonsoleNamespace, saName)
		if err != nil {
			// Try default
			_, err = s.repo.GetServiceAccount(ctx, dkonsoleNamespace, "default")
			if err != nil {
				return "", fmt.Errorf("serviceaccount not found in namespace %s", dkonsoleNamespace)
			}
			saName = "default"
		}
	}

	jobName := fmt.Sprintf("helm-%s-%s-%d", req.Operation, req.ReleaseName, time.Now().Unix())

	// Build Helm command
	helmCmd := s.BuildHelmCommand(HelmCommandRequest{
		Operation:    req.Operation,
		ReleaseName:  req.ReleaseName,
		Namespace:    req.Namespace,
		ChartName:    req.ChartName,
		Version:      req.Version,
		Repo:         req.Repo,
		ValuesYAML:   req.ValuesYAML,
		ValuesCMName: req.ValuesCMName,
	})

	// Create Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: dkonsoleNamespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: func() *int32 { t := int32(300); return &t }(),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: saName,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "helm",
							Image: "alpine/helm:latest",
						},
					},
				},
			},
		},
	}

	// Set command and args
	if len(helmCmd) == 3 && helmCmd[0] == "/bin/sh" && helmCmd[1] == "-c" {
		job.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh", "-c"}
		job.Spec.Template.Spec.Containers[0].Args = []string{helmCmd[2]}
	} else {
		if len(helmCmd) > 0 {
			job.Spec.Template.Spec.Containers[0].Command = helmCmd[0:1]
		}
		if len(helmCmd) > 1 {
			job.Spec.Template.Spec.Containers[0].Args = helmCmd[1:]
		}
	}

	// Add volume for values if needed
	if req.ValuesYAML != "" && req.ValuesCMName != "" {
		job.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "values",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: req.ValuesCMName,
						},
					},
				},
			},
		}
		job.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "values",
				MountPath: "/tmp/values",
			},
		}
	}

	if err := s.repo.CreateJob(ctx, dkonsoleNamespace, job); err != nil {
		return "", fmt.Errorf("failed to create helm job: %w", err)
	}

	return jobName, nil
}
