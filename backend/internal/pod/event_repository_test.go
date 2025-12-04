package pod

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestK8sEventRepository_GetEvents_Success(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "evt1",
			Namespace: "ns",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod1",
			Namespace: "ns",
		},
	})
	repo := NewK8sEventRepository(client)

	events, err := repo.GetEvents(context.Background(), "ns", "pod1")
	if err != nil {
		t.Fatalf("GetEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].Name != "evt1" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestK8sEventRepository_GetEvents_Error(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	client.Fake.PrependReactor("list", "events", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("list failed")
	})
	repo := NewK8sEventRepository(client)

	if _, err := repo.GetEvents(context.Background(), "ns", "pod1"); err == nil {
		t.Fatalf("expected error from GetEvents")
	}
}
