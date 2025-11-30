package k8s

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func TestWatchService_TransformEvent(t *testing.T) {
	service := NewWatchService(nil)

	tests := []struct {
		name     string
		event    watch.Event
		wantErr  bool
		wantType string
		wantName string
		wantNS   string
	}{
		{
			name: "valid ADD event",
			event: watch.Event{
				Type: watch.Added,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":      "test-pod",
							"namespace": "default",
						},
					},
				},
			},
			wantErr:  false,
			wantType: "ADDED",
			wantName: "test-pod",
			wantNS:   "default",
		},
		{
			name: "valid MODIFIED event",
			event: watch.Event{
				Type: watch.Modified,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":      "test-deployment",
							"namespace": "production",
						},
					},
				},
			},
			wantErr:  false,
			wantType: "MODIFIED",
			wantName: "test-deployment",
			wantNS:   "production",
		},
		{
			name: "valid DELETED event",
			event: watch.Event{
				Type: watch.Deleted,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":      "test-service",
							"namespace": "default",
						},
					},
				},
			},
			wantErr:  false,
			wantType: "DELETED",
			wantName: "test-service",
			wantNS:   "default",
		},
		{
			name: "cluster-scoped resource (no namespace)",
			event: watch.Event{
				Type: watch.Added,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": "test-namespace",
						},
					},
				},
			},
			wantErr:  false,
			wantType: "ADDED",
			wantName: "test-namespace",
			wantNS:   "",
		},
		{
			name: "invalid object type",
			event: watch.Event{
				Type:   watch.Added,
				Object: &runtime.Unknown{}, // Not unstructured
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.TransformEvent(tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("TransformEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if result == nil {
				t.Errorf("TransformEvent() result is nil")
				return
			}

			if result.Type != tt.wantType {
				t.Errorf("TransformEvent() Type = %v, want %v", result.Type, tt.wantType)
			}

			if result.Name != tt.wantName {
				t.Errorf("TransformEvent() Name = %v, want %v", result.Name, tt.wantName)
			}

			if result.Namespace != tt.wantNS {
				t.Errorf("TransformEvent() Namespace = %v, want %v", result.Namespace, tt.wantNS)
			}
		})
	}
}
