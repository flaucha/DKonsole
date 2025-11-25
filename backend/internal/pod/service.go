package pod

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

// PodService provides business logic for pod-related operations.
type PodService struct {
	eventRepo EventRepository
}

// NewPodService creates a new PodService with the provided event repository.
func NewPodService(eventRepo EventRepository) *PodService {
	return &PodService{eventRepo: eventRepo}
}

// GetPodEvents fetches, transforms, and sorts Kubernetes events for a specific pod.
// Events are sorted by LastSeen timestamp in descending order (most recent first).
// Returns an error if the events cannot be retrieved.
func (s *PodService) GetPodEvents(ctx context.Context, namespace, podName string) ([]EventInfo, error) {
	events, err := s.eventRepo.GetEvents(ctx, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	eventList := s.transformEvents(events)
	s.sortEventsByLastSeen(eventList)

	return eventList, nil
}

// transformEvents converts Kubernetes events to EventInfo DTOs
func (s *PodService) transformEvents(events []corev1.Event) []EventInfo {
	var eventList []EventInfo
	for _, event := range events {
		eventList = append(eventList, EventInfo{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Count:     event.Count,
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
			Source:    fmt.Sprintf("%s/%s", event.Source.Component, event.Source.Host),
		})
	}
	return eventList
}

// sortEventsByLastSeen sorts events by LastSeen timestamp
func (s *PodService) sortEventsByLastSeen(eventList []EventInfo) {
	sort.Slice(eventList, func(i, j int) bool {
		return eventList[i].LastSeen.After(eventList[j].LastSeen)
	})
}
