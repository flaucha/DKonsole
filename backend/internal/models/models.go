package models

import (
	"sync"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Credentials representa las credenciales de autenticación
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IDP      string `json:"idp,omitempty"` // "core" or "ldap" - optional, defaults to trying both
}

// Claims representa los claims del JWT
type Claims struct {
	Username         string            `json:"username"`
	Role             string            `json:"role"`
	IDP              string            `json:"idp,omitempty"`         // Identity Provider: "core" or "ldap"
	Permissions      map[string]string `json:"permissions,omitempty"` // namespace -> permission (view/edit)
	RegisteredClaims interface{}       `json:"-"`                     // Se manejará con jwt.RegisteredClaims en el paquete auth
}

// PaginatedResources representa una respuesta paginada de recursos
type PaginatedResources struct {
	Resources []Resource `json:"resources"`
	Continue  string     `json:"continue,omitempty"`  // Token para la siguiente página
	Remaining int        `json:"remaining,omitempty"` // Cantidad estimada de recursos restantes
}

// APIResourceInfo contiene información sobre un recurso de API descubierto
type APIResourceInfo struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Resource   string `json:"resource"`
	Kind       string `json:"kind"`
	Namespaced bool   `json:"namespaced"`
}

// APIResourceObject representa un objeto de un recurso de API
type APIResourceObject struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status,omitempty"`
	Created   string      `json:"created,omitempty"`
	Raw       interface{} `json:"raw,omitempty"`
}

// Handlers es la estructura principal que contiene los clients de Kubernetes
type Handlers struct {
	Clients       map[string]kubernetes.Interface
	Dynamics      map[string]dynamic.Interface
	Metrics       map[string]*metricsv.Clientset
	RESTConfigs   map[string]*rest.Config
	PrometheusURL string
	mu            sync.RWMutex
}

// Lock locks the handlers mutex for writing
func (h *Handlers) Lock() {
	h.mu.Lock()
}

// Unlock unlocks the handlers mutex
func (h *Handlers) Unlock() {
	h.mu.Unlock()
}

// RLock locks the handlers mutex for reading
func (h *Handlers) RLock() {
	h.mu.RLock()
}

// RUnlock unlocks the handlers mutex for reading
func (h *Handlers) RUnlock() {
	h.mu.RUnlock()
}

