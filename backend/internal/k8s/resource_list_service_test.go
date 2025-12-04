package k8s

import (
	"context"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func ctxWithPerms(perms map[string]string, role string) context.Context {
	return context.WithValue(context.Background(), auth.UserContextKey(), &models.Claims{
		Role:        role,
		Permissions: perms,
	})
}

func TestResourceListService_ListResources_PermissionDenied(t *testing.T) {
	svc := NewResourceListService(nil, "")
	client := fake.NewSimpleClientset()

	ctx := ctxWithPerms(map[string]string{"default": "view"}, "")
	_, err := svc.ListResources(ctx, ListResourcesRequest{
		Kind:      "ConfigMap",
		Namespace: "other",
		Client:    client,
	})

	if err == nil || !strings.Contains(err.Error(), "access denied to namespace: other") {
		t.Fatalf("expected permission error for other namespace, got %v", err)
	}
}

func TestResourceListService_ListResources_FiltersByPermissions(t *testing.T) {
	cmDefault := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-default",
			Namespace: "default",
		},
		Data: map[string]string{"k": "v"},
	}
	cmSystem := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-system",
			Namespace: "kube-system",
		},
		Data: map[string]string{"k": "v"},
	}

	client := fake.NewSimpleClientset(cmDefault, cmSystem)
	svc := NewResourceListService(nil, "")

	ctx := ctxWithPerms(map[string]string{"default": "view"}, "")
	resources, err := svc.ListResources(ctx, ListResourcesRequest{
		Kind:          "ConfigMap",
		Namespace:     "all",
		AllNamespaces: true,
		Client:        client,
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource after filtering, got %d", len(resources))
	}
	if resources[0].Namespace != "default" || resources[0].Name != "cm-default" {
		t.Fatalf("unexpected resource returned: %+v", resources[0])
	}
}

func TestResourceListService_ListResources_NoUser(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewResourceListService(nil, "")

	_, err := svc.ListResources(context.Background(), ListResourcesRequest{
		Kind:      "ConfigMap",
		Namespace: "default",
		Client:    client,
	})
	if err == nil || !strings.Contains(err.Error(), "user not found") {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestResourceListService_ListResources_ManyKinds(t *testing.T) {
	adminCtx := context.WithValue(context.Background(), auth.UserContextKey(), &auth.AuthClaims{
		Claims: models.Claims{Role: "admin"},
	})
	cases := []struct {
		name      string
		kind      string
		namespace string
		objs      []runtime.Object
	}{
		{
			name:      "secret",
			kind:      "Secret",
			namespace: "default",
			objs: []runtime.Object{
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "s1", Namespace: "default"}},
			},
		},
		{
			name:      "job",
			kind:      "Job",
			namespace: "default",
			objs: []runtime.Object{
				&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "job1", Namespace: "default"}},
			},
		},
		{
			name:      "cronjob",
			kind:      "CronJob",
			namespace: "default",
			objs: []runtime.Object{
				&batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{Name: "cj", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						Schedule: "*/5 * * * *",
						JobTemplate: batchv1.JobTemplateSpec{
							Spec: batchv1.JobSpec{
								Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{RestartPolicy: corev1.RestartPolicyOnFailure}},
							},
						},
					},
				},
			},
		},
		{
			name:      "statefulset",
			kind:      "StatefulSet",
			namespace: "default",
			objs: []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "sts", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo"}},
					},
				},
			},
		},
		{
			name:      "daemonset",
			kind:      "DaemonSet",
			namespace: "default",
			objs: []runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "default"},
					Spec: appsv1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo"}},
						Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "i"}}}},
					},
				},
			},
		},
		{
			name:      "hpa",
			kind:      "HorizontalPodAutoscaler",
			namespace: "default",
			objs: []runtime.Object{
				&autoscalingv2.HorizontalPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{Name: "hpa1", Namespace: "default"},
				},
			},
		},
		{
			name:      "ingress",
			kind:      "Ingress",
			namespace: "default",
			objs: []runtime.Object{
				&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "default"}},
			},
		},
		{
			name:      "serviceaccount",
			kind:      "ServiceAccount",
			namespace: "default",
			objs: []runtime.Object{
				&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "sa", Namespace: "default"}},
			},
		},
		{
			name:      "role",
			kind:      "Role",
			namespace: "default",
			objs: []runtime.Object{
				&rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "role", Namespace: "default"}},
			},
		},
		{
			name:      "clusterrole",
			kind:      "ClusterRole",
			namespace: "default",
			objs: []runtime.Object{
				&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "cr"}},
			},
		},
		{
			name:      "rolebinding",
			kind:      "RoleBinding",
			namespace: "default",
			objs: []runtime.Object{
				&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "rb", Namespace: "default"},
					RoleRef:    rbacv1.RoleRef{Name: "role"},
				},
			},
		},
		{
			name:      "clusterrolebinding",
			kind:      "ClusterRoleBinding",
			namespace: "default",
			objs: []runtime.Object{
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "crb"},
					RoleRef:    rbacv1.RoleRef{Name: "cr"},
				},
			},
		},
		{
			name:      "networkpolicy",
			kind:      "NetworkPolicy",
			namespace: "default",
			objs: []runtime.Object{
				&networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "np", Namespace: "default"},
					Spec:       networkingv1.NetworkPolicySpec{PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo"}}},
				},
			},
		},
		{
			name:      "pvc",
			kind:      "PersistentVolumeClaim",
			namespace: "default",
			objs: []runtime.Object{
				&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "default"}},
			},
		},
		{
			name:      "pv",
			kind:      "PersistentVolume",
			namespace: "default",
			objs: []runtime.Object{
				&corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv"}},
			},
		},
		{
			name:      "storageclass",
			kind:      "StorageClass",
			namespace: "default",
			objs: []runtime.Object{
				&storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "sc"}},
			},
		},
		{
			name:      "resourcequota",
			kind:      "ResourceQuota",
			namespace: "default",
			objs: []runtime.Object{
				&corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "rq", Namespace: "default"}},
			},
		},
		{
			name:      "limitrange",
			kind:      "LimitRange",
			namespace: "default",
			objs: []runtime.Object{
				&corev1.LimitRange{ObjectMeta: metav1.ObjectMeta{Name: "lr", Namespace: "default"}},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tt.objs...)
			svc := NewResourceListService(nil, "")

			resources, err := svc.ListResources(adminCtx, ListResourcesRequest{
				Kind:      tt.kind,
				Namespace: tt.namespace,
				Client:    client,
			})
			if err != nil {
				t.Fatalf("ListResources(%s) returned error: %v", tt.kind, err)
			}
			if len(resources) != 1 {
				t.Fatalf("expected 1 resource for %s, got %d", tt.kind, len(resources))
			}
			if resources[0].Name == "" {
				t.Fatalf("expected resource name to be set for %s", tt.kind)
			}
		})
	}
}
