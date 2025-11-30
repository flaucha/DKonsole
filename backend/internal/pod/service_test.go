package pod

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockEventRepository is a mock implementation of EventRepository
type mockEventRepository struct {
	getEventsFunc func(ctx context.Context, namespace, podName string) ([]corev1.Event, error)
}

func (m *mockEventRepository) GetEvents(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
	if m.getEventsFunc != nil {
		return m.getEventsFunc(ctx, namespace, podName)
	}
	return []corev1.Event{}, nil
}

func TestPodService_GetPodEvents(t *testing.T) {
	now := metav1.Now()
	earlier := metav1.NewTime(now.Time.Add(-1 * time.Hour))

	tests := []struct {
		name            string
		namespace       string
		podName         string
		getEventsFunc   func(ctx context.Context, namespace, podName string) ([]corev1.Event, error)
		wantErr         bool
		wantCount       int
		wantFirstReason string // After sorting, most recent should be first
		errMsg          string
	}{
		{
			name:      "successful events retrieval",
			namespace: "default",
			podName:   "test-pod",
			getEventsFunc: func(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
				return []corev1.Event{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "event-1",
							Namespace: namespace,
						},
						Type:           corev1.EventTypeNormal,
						Reason:         "Started",
						Message:        "Container started",
						Count:          1,
						FirstTimestamp: earlier,
						LastTimestamp:  earlier,
						Source: corev1.EventSource{
							Component: "kubelet",
							Host:      "node1",
						},
						InvolvedObject: corev1.ObjectReference{
							Kind:      "Pod",
							Name:      podName,
							Namespace: namespace,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "event-2",
							Namespace: namespace,
						},
						Type:           corev1.EventTypeWarning,
						Reason:         "Failed",
						Message:        "Container failed",
						Count:          2,
						FirstTimestamp: earlier,
						LastTimestamp:  now, // More recent
						Source: corev1.EventSource{
							Component: "kubelet",
							Host:      "node1",
						},
						InvolvedObject: corev1.ObjectReference{
							Kind:      "Pod",
							Name:      podName,
							Namespace: namespace,
						},
					},
				}, nil
			},
			wantErr:         false,
			wantCount:       2,
			wantFirstReason: "Failed", // Should be first after sorting by LastSeen
		},
		{
			name:      "empty events list",
			namespace: "default",
			podName:   "test-pod",
			getEventsFunc: func(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
				return []corev1.Event{}, nil
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "repository error",
			namespace: "default",
			podName:   "test-pod",
			getEventsFunc: func(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
				return nil, errors.New("repository error")
			},
			wantErr: true,
			errMsg:  "failed to get events",
		},
		{
			name:      "events sorted by LastSeen (most recent first)",
			namespace: "default",
			podName:   "test-pod",
			getEventsFunc: func(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
				return []corev1.Event{
					{
						ObjectMeta:     metav1.ObjectMeta{Name: "event-1", Namespace: namespace},
						Reason:         "OldEvent",
						FirstTimestamp: earlier,
						LastTimestamp:  earlier,
						Source:         corev1.EventSource{Component: "kubelet", Host: "node1"},
						InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: podName, Namespace: namespace},
					},
					{
						ObjectMeta:     metav1.ObjectMeta{Name: "event-2", Namespace: namespace},
						Reason:         "NewEvent",
						FirstTimestamp: earlier,
						LastTimestamp:  now, // More recent
						Source:         corev1.EventSource{Component: "kubelet", Host: "node1"},
						InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: podName, Namespace: namespace},
					},
				}, nil
			},
			wantErr:         false,
			wantCount:       2,
			wantFirstReason: "NewEvent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockEventRepository{
				getEventsFunc: tt.getEventsFunc,
			}

			service := NewPodService(mockRepo)
			ctx := context.Background()

			events, err := service.GetPodEvents(ctx, tt.namespace, tt.podName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPodEvents() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetPodEvents() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if len(events) != tt.wantCount {
				t.Errorf("GetPodEvents() count = %v, want %v", len(events), tt.wantCount)
				return
			}

			if tt.wantCount > 0 && tt.wantFirstReason != "" {
				if len(events) == 0 {
					t.Errorf("GetPodEvents() expected events but got empty list")
					return
				}
				// Check that events are sorted by LastSeen (most recent first)
				if events[0].Reason != tt.wantFirstReason {
					t.Errorf("GetPodEvents() first event reason = %v, want %v", events[0].Reason, tt.wantFirstReason)
				}
				// Verify sorting: each event should be more recent than the next
				for i := 0; i < len(events)-1; i++ {
					if events[i].LastSeen.Before(events[i+1].LastSeen) {
						t.Errorf("GetPodEvents() events not sorted correctly: event %d (%v) should be after event %d (%v)",
							i, events[i].LastSeen, i+1, events[i+1].LastSeen)
					}
				}
			}

			// Verify EventInfo structure
			for i, event := range events {
				if event.Type == "" && event.Reason != "" {
					// Type can be empty, but if Reason exists, we should have valid data
				}
				if event.Reason == "" {
					t.Errorf("GetPodEvents() event %d has empty Reason", i)
				}
				if event.Source == "" {
					t.Errorf("GetPodEvents() event %d has empty Source", i)
				}
			}
		})
	}
}
