package pod

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// EventRepository defines the interface for fetching pod events from Kubernetes
type EventRepository interface {
	GetEvents(ctx context.Context, namespace, podName string) ([]EventInfo, error)
}

// K8sEventRepository implements EventRepository using Kubernetes client-go
type K8sEventRepository struct {
	client kubernetes.Interface
}

// NewK8sEventRepository creates a new K8sEventRepository
func NewK8sEventRepository(client kubernetes.Interface) *K8sEventRepository {
	return &K8sEventRepository{client: client}
}

// GetEvents fetches events for a specific pod
func (r *K8sEventRepository) GetEvents(ctx context.Context, namespace, podName string) ([]EventInfo, error) {
	coreEvents, coreErr := r.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if coreErr != nil {
		utils.LogWarn("Failed to list corev1 pod events", map[string]interface{}{
			"namespace": namespace,
			"pod":       podName,
			"error":     coreErr.Error(),
		})
	}

	v1Events, v1Err := r.client.EventsV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("regarding.name=%s,regarding.kind=Pod", podName),
	})
	if v1Err != nil {
		utils.LogWarn("Failed to list events.k8s.io pod events", map[string]interface{}{
			"namespace": namespace,
			"pod":       podName,
			"error":     v1Err.Error(),
		})
	}

	if coreErr != nil && v1Err != nil {
		return nil, fmt.Errorf("failed to list pod events: core=%v, events.k8s.io=%v", coreErr, v1Err)
	}

	var eventInfos []EventInfo
	if coreErr == nil {
		for _, event := range coreEvents.Items {
			eventInfos = append(eventInfos, eventInfoFromCore(event))
		}
	}
	if v1Err == nil {
		for _, event := range v1Events.Items {
			eventInfos = append(eventInfos, eventInfoFromEventsV1(event))
		}
	}

	return dedupeEventInfos(eventInfos), nil
}

func eventInfoFromCore(event corev1.Event) EventInfo {
	count := event.Count
	if event.Series != nil && event.Series.Count > 0 {
		count = event.Series.Count
	}
	if count == 0 {
		count = 1
	}

	firstSeen := event.FirstTimestamp.Time
	lastSeen := event.LastTimestamp.Time
	if event.Series != nil && !event.Series.LastObservedTime.IsZero() {
		lastSeen = event.Series.LastObservedTime.Time
	}
	if !event.EventTime.IsZero() {
		lastSeen = event.EventTime.Time
	}

	firstSeen, lastSeen = normalizeEventTimes(firstSeen, lastSeen, event.CreationTimestamp.Time)

	return EventInfo{
		Type:      event.Type,
		Reason:    event.Reason,
		Message:   event.Message,
		Count:     count,
		FirstSeen: firstSeen,
		LastSeen:  lastSeen,
		Source:    formatEventSource(event.Source.Component, event.Source.Host),
	}
}

func eventInfoFromEventsV1(event eventsv1.Event) EventInfo {
	count := int32(1)
	if event.Series != nil && event.Series.Count > 0 {
		count = event.Series.Count
	} else if event.DeprecatedCount > 0 {
		count = event.DeprecatedCount
	}

	firstSeen := event.DeprecatedFirstTimestamp.Time
	lastSeen := event.DeprecatedLastTimestamp.Time
	if event.Series != nil && !event.Series.LastObservedTime.IsZero() {
		lastSeen = event.Series.LastObservedTime.Time
	}
	if !event.EventTime.IsZero() {
		lastSeen = event.EventTime.Time
	}

	firstSeen, lastSeen = normalizeEventTimes(firstSeen, lastSeen, event.CreationTimestamp.Time)

	source := formatEventSource(event.ReportingController, event.ReportingInstance)
	if source == "" {
		source = formatEventSource(event.DeprecatedSource.Component, event.DeprecatedSource.Host)
	}

	return EventInfo{
		Type:      event.Type,
		Reason:    event.Reason,
		Message:   event.Note,
		Count:     count,
		FirstSeen: firstSeen,
		LastSeen:  lastSeen,
		Source:    source,
	}
}

func formatEventSource(component, host string) string {
	if component == "" && host == "" {
		return ""
	}
	if component == "" {
		return host
	}
	if host == "" {
		return component
	}
	return fmt.Sprintf("%s/%s", component, host)
}

func normalizeEventTimes(firstSeen, lastSeen, fallback time.Time) (time.Time, time.Time) {
	if lastSeen.IsZero() && !fallback.IsZero() {
		lastSeen = fallback
	}
	if firstSeen.IsZero() {
		if !lastSeen.IsZero() {
			firstSeen = lastSeen
		} else if !fallback.IsZero() {
			firstSeen = fallback
		}
	}
	if lastSeen.IsZero() {
		lastSeen = firstSeen
	}
	return firstSeen, lastSeen
}

func dedupeEventInfos(events []EventInfo) []EventInfo {
	if len(events) == 0 {
		return events
	}

	merged := make(map[string]EventInfo, len(events))
	for _, event := range events {
		key := fmt.Sprintf("%s|%s|%s", event.Type, event.Reason, event.Message)
		existing, ok := merged[key]
		if !ok {
			merged[key] = event
			continue
		}

		if event.Count > existing.Count {
			existing.Count = event.Count
		}
		if existing.FirstSeen.IsZero() || (!event.FirstSeen.IsZero() && event.FirstSeen.Before(existing.FirstSeen)) {
			existing.FirstSeen = event.FirstSeen
		}
		if event.LastSeen.After(existing.LastSeen) {
			existing.LastSeen = event.LastSeen
		}
		if existing.Source == "" {
			existing.Source = event.Source
		}
		merged[key] = existing
	}

	result := make([]EventInfo, 0, len(merged))
	for _, event := range merged {
		result = append(result, event)
	}
	return result
}
