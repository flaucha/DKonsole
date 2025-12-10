package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
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

func TestWatchService_StartWatch(t *testing.T) {
	service := NewWatchService(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme)

	// Case 1: Success (Pod - namespaced)
	req := WatchRequest{Kind: "Pod", Namespace: "default"}
	w, err := service.StartWatch(ctx, client, req)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	if w != nil {
		w.Stop()
	}

	// Case 2: Success (Node - cluster scoped)
	req = WatchRequest{Kind: "Node"}
	w, err = service.StartWatch(ctx, client, req)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	if w != nil {
		w.Stop()
	}

	// Case 3: Unsupported Kind
	req = WatchRequest{Kind: "UnknownKind"}
	_, err = service.StartWatch(ctx, client, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported kind")

	// Case 4: Watch Error
	// Mock reactor to fail Watch
	client.PrependWatchReactor("*", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, nil, fmt.Errorf("simulated watch error")
	})
	req = WatchRequest{Kind: "Pod", Namespace: "default"}
	_, err = service.StartWatch(ctx, client, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated watch error")
}
