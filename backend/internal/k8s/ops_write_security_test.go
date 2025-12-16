package k8s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestMaxBodySize_Security(t *testing.T) {
	// Setup Mocks
	mockClusterService := new(MockClusterService)
	// We need to verify if MockServiceFactory exists in this package.
	// From previous context (helm pkg), it was there. But we are in k8s pkg.
	// We should check if k8s package has a ServiceFactory mock.
	// Assuming likely not or named differently.
	// However, Service struct in k8s/k8s.go:
	// type Service struct { ... serviceFactory Factory ... } (Factory interface defined in factories.go)
	// So we need a mock that implements Factories interface.

	// Create a mock factory locally if needed or reuse if available.
	// Looking at k8s_handlers_new_test.go or k8s_handlers_test.go might reveal if there is a global mock.
	// If not, we can rely on real factories or create a mock here.

	// To be safe and simple: just construct the service with minimal dependencies.
	// CreateResourceYAML uses s.serviceFactory.CreateResourceService
	// So we DO need a mock factory or real one.
	// Since we want to test the HANDLER (body reading happens BEFORE service call potentially?),
	// Actually, body reading happens at the start of CreateResourceYAML.
	// So even if service call fails or factory is nil, we might hit the body size error FIRST?
	// The code:
	// 1. Read Body (limit here)
	// 2. Create Context
	// 3. Get Clients
	// 4. Create Service

	// So we can technically test this even if GetClient fails, provided body read happens before GetClient.
	// Let's check ops_write.go again.
	// YES. Body read is the FIRST thing.

	handlers := &models.Handlers{}
	service := NewService(handlers, mockClusterService)

	// Create a payload larger than 1MB
	largePayload := strings.Repeat("a", 1048576+10)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "CreateResource - Too Large",
			method:         http.MethodPost,
			path:           "/api/k8s/resources/create",
			body:           largePayload,
			expectedStatus: http.StatusBadRequest, // Should fail here
		},
		{
			name:           "UpdateResource - Too Large",
			method:         http.MethodPut,
			path:           "/api/k8s/resources/update?kind=Pod&name=p1&namespace=default",
			body:           largePayload,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ImportResource - Too Large",
			method:         http.MethodPost,
			path:           "/api/k8s/resources/import",
			body:           largePayload,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			// Inject User Context
			claims := &models.Claims{Username: "testuser", Role: "admin"}
			ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
			// Add valid permissions check bypass? permissions check happens BEFORE body read for UpdateResource namespaced.
			// But for CreateResource it happens later (implicit?).
			// Check ops_write.go: UpdateResourceYAML checks permissions at start if namespaced.
			// Our test case uses namespace=default. So it will check.
			// Admin role usually bypasses permissions in ValidateAction?
			// Let's ensure it passes permission check.
			// UpdateResourceYAML checks permissions.CanPerformAction.
			// If we mock permissions or just use admin, we hope it passes.
			// Actually permissions.CanPerformAction uses GetPermissionLevel.
			// Admin role usually implies full access?
			// Let's assume yes or rely on mock if needed.

			req = req.WithContext(ctx)

			switch tt.path {
			case "/api/k8s/resources/create":
				service.CreateResourceYAML(w, req)
			case "/api/k8s/resources/import":
				service.ImportResourceYAML(w, req)
			default:
				if strings.Contains(tt.path, "update") {
					service.UpdateResourceYAML(w, req)
				}
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
