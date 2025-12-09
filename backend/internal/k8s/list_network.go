package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listServices(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0, len(list.Items))
	for _, item := range list.Items {
		var ports []string
		for _, p := range item.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.TargetPort.IntVal, p.Protocol))
		}

		// Combine ExternalIPs and LoadBalancer IPs
		var externalIPs []string
		externalIPs = append(externalIPs, item.Spec.ExternalIPs...)
		externalIPs = append(externalIPs, getExternalIPs(item.Status.LoadBalancer)...)

		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "Service",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    string(item.Spec.Type),
			Details: map[string]interface{}{
				"type":        string(item.Spec.Type),
				"clusterIP":   item.Spec.ClusterIP,
				"externalIPs": externalIPs,
				"ports":       ports,
				"selector":    item.Spec.Selector,
			},
		})
	}
	return resources, nil
}

func (s *ResourceListService) listIngresses(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0, len(list.Items))
	for _, item := range list.Items {
		var rules []interface{}
		for _, rule := range item.Spec.Rules {
			var paths []interface{} // Changed to interface{} to support object structure
			if rule.HTTP != nil {
				for _, p := range rule.HTTP.Paths {
					// Backend struct for path info
					pathInfo := map[string]interface{}{
						"path":        p.Path,
						"pathType":    string(*p.PathType),
						"serviceName": p.Backend.Service.Name,
					}
					if p.Backend.Service.Port.Name != "" {
						pathInfo["servicePort"] = p.Backend.Service.Port.Name
					} else {
						pathInfo["servicePort"] = p.Backend.Service.Port.Number
					}
					paths = append(paths, pathInfo)
				}
			}
			rules = append(rules, map[string]interface{}{
				"host":  rule.Host,
				"paths": paths,
			})
		}

		var tls []interface{}
		for _, t := range item.Spec.TLS {
			tls = append(tls, map[string]interface{}{
				"hosts":      t.Hosts,
				"secretName": t.SecretName,
			})
		}

		var lb []interface{}
		for _, ingress := range item.Status.LoadBalancer.Ingress {
			lbObj := map[string]string{}
			if ingress.IP != "" {
				lbObj["ip"] = ingress.IP
			}
			if ingress.Hostname != "" {
				lbObj["hostname"] = ingress.Hostname
			}
			lb = append(lb, lbObj)
		}

		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "Ingress",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    "Active",
			Details: map[string]interface{}{
				"class": func() string {
					if item.Spec.IngressClassName != nil {
						return *item.Spec.IngressClassName
					}
					return ""
				}(),
				"rules":        rules,
				"tls":          tls,
				"loadBalancer": lb,
				"annotations":  item.Annotations,
			},
		})
	}
	return resources, nil
}

func (s *ResourceListService) listNetworkPolicies(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0, len(list.Items))
	for _, item := range list.Items {
		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "NetworkPolicy",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    "Active",
			Details: map[string]interface{}{
				"podSelector": item.Spec.PodSelector, // Contains MatchLabels
				"policyTypes": item.Spec.PolicyTypes,
			},
		})
	}
	return resources, nil
}

func getExternalIPs(lb corev1.LoadBalancerStatus) []string {
	var ips []string
	for _, ingress := range lb.Ingress {
		if ingress.IP != "" {
			ips = append(ips, ingress.IP)
		}
		if ingress.Hostname != "" {
			ips = append(ips, ingress.Hostname)
		}
	}
	return ips
}
