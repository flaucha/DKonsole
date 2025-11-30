package k8s

import (
	"context"
	"errors"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockDeploymentRepository is a mock implementation of DeploymentRepository for testing
type mockDeploymentRepository struct {
	getScaleFunc         func(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error)
	updateScaleFunc      func(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error)
	getDeploymentFunc    func(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	updateDeploymentFunc func(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error)
}

func (m *mockDeploymentRepository) GetScale(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error) {
	if m.getScaleFunc != nil {
		return m.getScaleFunc(ctx, namespace, name)
	}
	return &autoscalingv1.Scale{Spec: autoscalingv1.ScaleSpec{Replicas: 3}}, nil
}

func (m *mockDeploymentRepository) UpdateScale(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error) {
	if m.updateScaleFunc != nil {
		return m.updateScaleFunc(ctx, namespace, name, scale)
	}
	return scale, nil
}

func (m *mockDeploymentRepository) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	if m.getDeploymentFunc != nil {
		return m.getDeploymentFunc(ctx, namespace, name)
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: make(map[string]string),
				},
			},
		},
	}, nil
}

func (m *mockDeploymentRepository) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	if m.updateDeploymentFunc != nil {
		return m.updateDeploymentFunc(ctx, namespace, deployment)
	}
	return deployment, nil
}

func TestDeploymentService_ScaleDeployment(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		deploymentName  string
		delta           int
		currentReplicas int32
		getScaleErr     error
		updateScaleErr  error
		wantReplicas    int32
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "scale up successfully",
			namespace:       "default",
			deploymentName:  "my-deployment",
			delta:           2,
			currentReplicas: 3,
			wantReplicas:    5,
			wantErr:         false,
		},
		{
			name:            "scale down successfully",
			namespace:       "default",
			deploymentName:  "my-deployment",
			delta:           -2,
			currentReplicas: 5,
			wantReplicas:    3,
			wantErr:         false,
		},
		{
			name:            "scale to zero when delta is negative enough",
			namespace:       "default",
			deploymentName:  "my-deployment",
			delta:           -5,
			currentReplicas: 3,
			wantReplicas:    0, // Minimum is 0
			wantErr:         false,
		},
		{
			name:            "scale from zero",
			namespace:       "default",
			deploymentName:  "my-deployment",
			delta:           3,
			currentReplicas: 0,
			wantReplicas:    3,
			wantErr:         false,
		},
		{
			name:            "get scale error",
			namespace:       "default",
			deploymentName:  "my-deployment",
			delta:           2,
			currentReplicas: 3,
			getScaleErr:     errors.New("deployment not found"),
			wantErr:         true,
			errMsg:          "failed to get current scale",
		},
		{
			name:            "update scale error",
			namespace:       "default",
			deploymentName:  "my-deployment",
			delta:           2,
			currentReplicas: 3,
			updateScaleErr:  errors.New("update failed"),
			wantErr:         true,
			errMsg:          "failed to update scale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDeploymentRepository{
				getScaleFunc: func(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error) {
					if tt.getScaleErr != nil {
						return nil, tt.getScaleErr
					}
					return &autoscalingv1.Scale{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
						},
						Spec: autoscalingv1.ScaleSpec{
							Replicas: tt.currentReplicas,
						},
					}, nil
				},
				updateScaleFunc: func(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error) {
					if tt.updateScaleErr != nil {
						return nil, tt.updateScaleErr
					}
					return scale, nil
				},
			}

			service := NewDeploymentService(mockRepo)
			ctx := context.Background()

			gotReplicas, err := service.ScaleDeployment(ctx, tt.namespace, tt.deploymentName, tt.delta)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScaleDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errMsg != "" {
					if err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("ScaleDeployment() error = %v, want error starting with %v", err, tt.errMsg)
					}
				}
				return
			}

			if gotReplicas != tt.wantReplicas {
				t.Errorf("ScaleDeployment() replicas = %v, want %v", gotReplicas, tt.wantReplicas)
			}
		})
	}
}

func TestDeploymentService_RolloutDeployment(t *testing.T) {
	tests := []struct {
		name                string
		namespace           string
		deploymentName      string
		getDeploymentErr    error
		updateDeploymentErr error
		wantErr             bool
		errMsg              string
		checkAnnotation     bool
	}{
		{
			name:            "rollout successful",
			namespace:       "default",
			deploymentName:  "my-deployment",
			wantErr:         false,
			checkAnnotation: true,
		},
		{
			name:             "deployment not found",
			namespace:        "default",
			deploymentName:   "not-found",
			getDeploymentErr: errors.New("deployment not found"),
			wantErr:          true,
			errMsg:           "failed to get deployment",
		},
		{
			name:                "update deployment error",
			namespace:           "default",
			deploymentName:      "my-deployment",
			updateDeploymentErr: errors.New("update failed"),
			wantErr:             true,
			errMsg:              "failed to update deployment",
		},
		{
			name:            "rollout with existing annotations",
			namespace:       "default",
			deploymentName:  "my-deployment",
			wantErr:         false,
			checkAnnotation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.deploymentName,
					Namespace: tt.namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"existing": "annotation",
							},
						},
					},
				},
			}

			// Clear annotations if needed for test
			if tt.name == "rollout with existing annotations" {
				deployment.Spec.Template.Annotations = make(map[string]string)
			}

			mockRepo := &mockDeploymentRepository{
				getDeploymentFunc: func(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
					if tt.getDeploymentErr != nil {
						return nil, tt.getDeploymentErr
					}
					return deployment, nil
				},
				updateDeploymentFunc: func(ctx context.Context, namespace string, dep *appsv1.Deployment) (*appsv1.Deployment, error) {
					if tt.updateDeploymentErr != nil {
						return nil, tt.updateDeploymentErr
					}
					if tt.checkAnnotation {
						// Verify restart annotation was added
						if _, exists := dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartAt"]; !exists {
							t.Errorf("RolloutDeployment() restart annotation not found")
						} else {
							// Verify timestamp format
							restartAt := dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartAt"]
							_, err := time.Parse(time.RFC3339, restartAt)
							if err != nil {
								t.Errorf("RolloutDeployment() restart annotation timestamp format invalid: %v", err)
							}
						}
					}
					return dep, nil
				},
			}

			service := NewDeploymentService(mockRepo)
			ctx := context.Background()

			err := service.RolloutDeployment(ctx, tt.namespace, tt.deploymentName)

			if (err != nil) != tt.wantErr {
				t.Errorf("RolloutDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errMsg != "" {
					if err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("RolloutDeployment() error = %v, want error starting with %v", err, tt.errMsg)
					}
				}
			}
		})
	}
}
