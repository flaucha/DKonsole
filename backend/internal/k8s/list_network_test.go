package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestResourceListService_ListIngresses(t *testing.T) {
	// Setup
	client := k8sfake.NewSimpleClientset()
	svc := NewResourceListService(nil, "")

	pathTypePrefix := networkingv1.PathTypePrefix
	ingressClassName := "nginx"

	// Create complex Ingress
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: "example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/app",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "app-svc",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
								{
									Path:     "/api",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "api-svc",
											Port: networkingv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{"example.com"},
					SecretName: "tls-secret",
				},
			},
		},
		Status: networkingv1.IngressStatus{
			LoadBalancer: networkingv1.IngressLoadBalancerStatus{
				Ingress: []networkingv1.IngressLoadBalancerIngress{
					{IP: "1.2.3.4"},
					{Hostname: "lb.aws.com"},
				},
			},
		},
	}
	_, err := client.NetworkingV1().Ingresses("default").Create(context.Background(), ingress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create ingress: %v", err)
	}

	// Create Request
	req := ListResourcesRequest{
		Kind:      "Ingress",
		Namespace: "default",
		Client:    client,
	}

	// Execute
	resources, err := svc.listIngresses(context.Background(), req.Client, req.Namespace, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listIngresses failed: %v", err)
	}

	// Verify
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	res := resources[0]
	if res.Name != "test-ingress" {
		t.Errorf("expected name test-ingress, got %s", res.Name)
	}
	if res.Kind != "Ingress" {
		t.Errorf("expected kind Ingress, got %s", res.Kind)
	}

	details := res.Details.(map[string]interface{})

	// Check Class
	if details["class"] != "nginx" {
		t.Errorf("expected class nginx, got %v", details["class"])
	}

	// Check LoadBalancer parse (getExternalIPs logic is implicitly tested via Status logic? No, listIngresses parses LB separately in loop)
	lb := details["loadBalancer"].([]interface{})
	if len(lb) != 2 {
		t.Errorf("expected 2 LB entries, got %d", len(lb))
	}

	// Check Rules processing
	rules := details["rules"].([]interface{})
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
	rule0 := rules[0].(map[string]interface{})
	if rule0["host"] != "example.com" {
		t.Errorf("expected host example.com, got %v", rule0["host"])
	}

	paths := rule0["paths"].([]interface{})
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
}

func TestResourceListService_ListServices_ExternalIPs(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	svc := NewResourceListService(nil, "")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:        corev1.ServiceTypeLoadBalancer,
			ExternalIPs: []string{"10.0.0.1"},
			Ports: []corev1.ServicePort{
				{Port: 80, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP},
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "1.2.3.4"},
					{Hostname: "lb.aws.com"},
				},
			},
		},
	}
	_, _ = client.CoreV1().Services("default").Create(context.Background(), service, metav1.CreateOptions{})

	resources, err := svc.listServices(context.Background(), client, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listServices failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	details := resources[0].Details.(map[string]interface{})
	extIPs := details["externalIPs"].([]string)

	// Should have 3 IPs: 10.0.0.1 (Spec), 1.2.3.4 (Status), lb.aws.com (Status)
	if len(extIPs) != 3 {
		t.Errorf("expected 3 external IPs, got %d: %v", len(extIPs), extIPs)
	}
}

// Wait, corev1 ServicePort TargetPort is intstr.IntOrString.
// I should import "k8s.io/apimachinery/pkg/util/intstr"
