package k8s

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestK8sClusterStatsRepository_Counts(t *testing.T) {
	client := k8sfake.NewSimpleClientset(
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "cp", Labels: map[string]string{"node-role.kubernetes.io/control-plane": "true"}}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "master", Labels: map[string]string{"node-role.kubernetes.io/master": "true"}}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "cp-taint"}, Spec: corev1.NodeSpec{Taints: []corev1.Taint{{Key: "node-role.kubernetes.io/control-plane"}}}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "master-taint"}, Spec: corev1.NodeSpec{Taints: []corev1.Taint{{Key: "node-role.kubernetes.io/master"}}}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default"}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deploy", Namespace: "default"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"}},
		&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "default"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "default"}},
		&corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv"}},
	)
	repo := NewK8sClusterStatsRepository(client)
	ctx := context.Background()

	if count, err := repo.GetNodeCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetNodeCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetNamespaceCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetNamespaceCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetPodCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetPodCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetDeploymentCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetDeploymentCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetServiceCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetServiceCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetIngressCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetIngressCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetPVCCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetPVCCount = %d, err=%v", count, err)
	}
	if count, err := repo.GetPVCount(ctx); err != nil || count != 1 {
		t.Fatalf("GetPVCount = %d, err=%v", count, err)
	}
}

func TestK8sDeploymentRepository(t *testing.T) {
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy",
			Namespace: "ns",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
	}
	client := k8sfake.NewSimpleClientset(dep)
	scaleObj := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy",
			Namespace: "ns",
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replicas,
		},
	}
	client.Fake.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if sub, ok := action.(k8stesting.GetAction); ok && sub.GetSubresource() == "scale" {
			return true, scaleObj, nil
		}
		return false, nil, nil
	})
	client.Fake.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if sub, ok := action.(k8stesting.UpdateAction); ok && sub.GetSubresource() == "scale" {
			obj := sub.GetObject().(*autoscalingv1.Scale)
			scaleObj = obj
			return true, obj, nil
		}
		return false, nil, nil
	})
	repo := NewK8sDeploymentRepository(client)
	ctx := context.Background()

	scale, err := repo.GetScale(ctx, "ns", "deploy")
	if err != nil {
		t.Fatalf("GetScale error: %v", err)
	}
	if scale.Spec.Replicas != replicas {
		t.Fatalf("expected replicas %d, got %d", replicas, scale.Spec.Replicas)
	}

	scale.Spec.Replicas = 3
	updatedScale, err := repo.UpdateScale(ctx, "ns", "deploy", scale)
	if err != nil {
		t.Fatalf("UpdateScale error: %v", err)
	}
	if updatedScale.Spec.Replicas != 3 {
		t.Fatalf("scale not updated, got %d", updatedScale.Spec.Replicas)
	}

	deployment, err := repo.GetDeployment(ctx, "ns", "deploy")
	if err != nil {
		t.Fatalf("GetDeployment error: %v", err)
	}
	deployment.Spec.Replicas = &replicas

	updatedDep, err := repo.UpdateDeployment(ctx, "ns", deployment)
	if err != nil {
		t.Fatalf("UpdateDeployment error: %v", err)
	}
	if updatedDep.Spec.Replicas == nil || *updatedDep.Spec.Replicas != replicas {
		t.Fatalf("deployment replicas not preserved")
	}
}
