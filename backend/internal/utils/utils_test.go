package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestIsSystemNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      bool
	}{
		{"kube-system", "kube-system", true},
		{"kube-public", "kube-public", true},
		{"kube-node-lease", "kube-node-lease", true},
		{"default", "default", false},
		{"custom-namespace", "custom-namespace", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSystemNamespace(tt.namespace); got != tt.want {
				t.Errorf("IsSystemNamespace(%q) = %v, want %v", tt.namespace, got, tt.want)
			}
		})
	}
}

func TestValidateK8sName(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		paramName string
		wantErr   bool
		errMsg    string
	}{
		{"valid name", "my-resource", "name", false, ""},
		{"valid name with dot", "my.resource", "name", false, ""},
		{"valid name alphanumeric", "myresource123", "name", false, ""},
		{"empty name", "", "name", true, "name is required"},
		{"name too long", string(make([]byte, 254)), "name", true, "too long"},
		{"name with uppercase", "MyResource", "name", true, "must consist of lower case"},
		{"name starting with dash", "-myresource", "name", true, "must start and end with an alphanumeric"},
		{"name ending with dash", "myresource-", "name", true, "must start and end with an alphanumeric"},
		{"name with underscore", "my_resource", "name", true, "must consist of lower case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateK8sName(tt.inputName, tt.paramName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateK8sName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() == "" || (tt.errMsg != "" && !contains(err.Error(), tt.errMsg)) {
					t.Errorf("ValidateK8sName() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestCreateTimeoutContext(t *testing.T) {
	// Save original env
	originalTimeout := os.Getenv("K8S_OPERATION_TIMEOUT")
	defer os.Setenv("K8S_OPERATION_TIMEOUT", originalTimeout)

	// Test default timeout
	os.Unsetenv("K8S_OPERATION_TIMEOUT")
	ctx, cancel := CreateTimeoutContext()
	if ctx == nil {
		t.Error("CreateTimeoutContext() returned nil context")
	}
	if cancel == nil {
		t.Error("CreateTimeoutContext() returned nil cancel function")
	}
	cancel()

	// Test custom timeout
	os.Setenv("K8S_OPERATION_TIMEOUT", "10s")
	ctx, cancel = CreateTimeoutContext()
	if ctx == nil {
		t.Error("CreateTimeoutContext() returned nil context")
	}
	cancel()
}

func TestCreateRequestContext(t *testing.T) {
	// Save original env
	originalTimeout := os.Getenv("K8S_OPERATION_TIMEOUT")
	defer func() {
		if originalTimeout != "" {
			os.Setenv("K8S_OPERATION_TIMEOUT", originalTimeout)
		} else {
			os.Unsetenv("K8S_OPERATION_TIMEOUT")
		}
	}()

	req := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := CreateRequestContext(req)
	if ctx == nil {
		t.Error("CreateRequestContext() returned nil context")
	}
	if cancel == nil {
		t.Error("CreateRequestContext() returned nil cancel function")
	}

	// Verify context is derived from request context and not canceled initially
	select {
	case <-ctx.Done():
		t.Error("CreateRequestContext() context should not be canceled initially")
	default:
		// Context is not canceled, which is what we want
	}

	// Clean up
	cancel()
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		userMessage  string
		statusCode   int
		expectedCode int
	}{
		{"internal error", nil, "Internal server error", http.StatusInternalServerError, http.StatusInternalServerError},
		{"not found", nil, "Not found", http.StatusNotFound, http.StatusNotFound},
		{"bad request", nil, "Bad request", http.StatusBadRequest, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			HandleError(rr, tt.err, tt.userMessage, tt.statusCode)
			if rr.Code != tt.expectedCode {
				t.Errorf("HandleError() status code = %v, want %v", rr.Code, tt.expectedCode)
			}
			if rr.Body.String() != tt.userMessage+"\n" {
				t.Errorf("HandleError() body = %q, want %q", rr.Body.String(), tt.userMessage+"\n")
			}
		})
	}
}

func TestCheckQuotaLimits(t *testing.T) {
	quota := corev1.ResourceQuota{
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsCPU:    resource.MustParse("2"),
				corev1.ResourceRequestsMemory: resource.MustParse("4Gi"),
			},
		},
	}

	// Test with valid requests
	requests := map[string]interface{}{
		"cpu":    "1",
		"memory": "2Gi",
	}
	limits := map[string]interface{}{}

	err := CheckQuotaLimits(quota, requests, limits)
	if err != nil {
		t.Errorf("CheckQuotaLimits() error = %v, want nil", err)
	}

	// Test with empty requests
	err = CheckQuotaLimits(quota, map[string]interface{}{}, map[string]interface{}{})
	if err != nil {
		t.Errorf("CheckQuotaLimits() error = %v, want nil", err)
	}
}

func TestCheckStorageQuota(t *testing.T) {
	quota := corev1.ResourceQuota{
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsStorage: resource.MustParse("100Gi"),
			},
		},
	}

	// Test with valid storage
	storage := "50Gi"
	err := CheckStorageQuota(quota, storage)
	if err != nil {
		t.Errorf("CheckStorageQuota() error = %v, want nil", err)
	}

	// Test with empty storage
	err = CheckStorageQuota(quota, "")
	if err != nil {
		t.Errorf("CheckStorageQuota() error = %v, want nil", err)
	}

	// Test with nil hard limits
	emptyQuota := corev1.ResourceQuota{
		Spec: corev1.ResourceQuotaSpec{
			Hard: nil,
		},
	}
	err = CheckStorageQuota(emptyQuota, "50Gi")
	if err != nil {
		t.Errorf("CheckStorageQuota() with nil hard limits error = %v, want nil", err)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{"X-Forwarded-For", map[string]string{"X-Forwarded-For": "192.168.1.1"}, "", "192.168.1.1"},
		{"X-Real-IP", map[string]string{"X-Real-IP": "10.0.0.1"}, "", "10.0.0.1"},
		{"RemoteAddr", map[string]string{}, "192.168.1.1:8080", "192.168.1.1"},
		{"X-Forwarded-For multiple IPs", map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1"}, "", "192.168.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := GetClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("GetClientIP() IP = %v, want %v", ip, tt.expectedIP)
			}
		})
	}
}

func TestAuditLog(t *testing.T) {
	// hook logger to buffer
	var buf bytes.Buffer
	Logger.SetOutput(&buf)
	defer Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.JSONFormatter{})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	
	AuditLog(req, "test-action", "test-resource", "test-name", "default", true, nil, map[string]interface{}{"key": "value"})

	// Verify log output
	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry["action"] != "test-action" {
		t.Errorf("AuditLog() action = %v, want %v", entry["action"], "test-action")
	}
	if entry["user"] != "anonymous" {
		t.Errorf("AuditLog() user = %v, want %v", entry["user"], "anonymous")
	}
	if entry["key"] != "value" {
		t.Errorf("AuditLog() details key = %v, want %v", entry["key"], "value")
	}
}

func TestLogAuditEntry(t *testing.T) {
	var buf bytes.Buffer
	Logger.SetOutput(&buf)
	defer Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.JSONFormatter{})

	entry := AuditLogEntry{
		User:      "test-user",
		IP:        "10.0.0.1",
		Action:    "create",
		Resource:  "pods",
		Namespace: "default",
		Success:   true,
		Details:   map[string]interface{}{"foo": "bar"},
	}

	LogAuditEntry(entry)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if logEntry["user"] != "test-user" {
		t.Errorf("LogAuditEntry() user = %v, want %v", logEntry["user"], "test-user")
	}
	if logEntry["foo"] != "bar" {
		t.Errorf("LogAuditEntry() detail foo = %v, want %v", logEntry["foo"], "bar")
	}
}

func TestJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}
	
	JSONResponse(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("JSONResponse() status = %v, want %v", w.Code, http.StatusCreated)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("JSONResponse() Content-Type = %v, want application/json", w.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["foo"] != "bar" {
		t.Errorf("JSONResponse() body = %v, want %v", response["foo"], "bar")
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	ErrorResponse(w, http.StatusForbidden, "access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("ErrorResponse() status = %v, want %v", w.Code, http.StatusForbidden)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["error"] != "access denied" {
		t.Errorf("ErrorResponse() error = %v, want %v", response["error"], "access denied")
	}
}

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}
	SuccessResponse(w, "created successfully", data)

	if w.Code != http.StatusOK {
		t.Errorf("SuccessResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["status"] != "success" {
		t.Errorf("SuccessResponse() status = %v, want success", response["status"])
	}
	if response["message"] != "created successfully" {
		t.Errorf("SuccessResponse() message = %v, want created successfully", response["message"])
	}
	
	respData, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("SuccessResponse() data is not a map")
	}
	if respData["id"] != "123" {
		t.Errorf("SuccessResponse() data.id = %v, want 123", respData["id"])
	}
}

func TestParsePodParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		wantErr     bool
		wantParams  *PodParams
	}{
		{
			name: "valid params",
			queryParams: map[string]string{
				"namespace": "default",
				"pod":       "my-pod",
				"container": "my-container",
			},
			wantErr: false,
			wantParams: &PodParams{
				Namespace: "default",
				PodName:   "my-pod",
				Container: "my-container",
			},
		},
		{
			name: "missing namespace",
			queryParams: map[string]string{
				"pod": "my-pod",
			},
			wantErr: true,
		},
		{
			name: "missing pod",
			queryParams: map[string]string{
				"namespace": "default",
			},
			wantErr: true,
		},
		{
			name: "invalid namespace",
			queryParams: map[string]string{
				"namespace": "InvalidNamespace",
				"pod":       "my-pod",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()

			got, err := ParsePodParams(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePodParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Namespace != tt.wantParams.Namespace {
					t.Errorf("ParsePodParams() namespace = %v, want %v", got.Namespace, tt.wantParams.Namespace)
				}
				if got.PodName != tt.wantParams.PodName {
					t.Errorf("ParsePodParams() pod = %v, want %v", got.PodName, tt.wantParams.PodName)
				}
				if got.Container != tt.wantParams.Container {
					t.Errorf("ParsePodParams() container = %v, want %v", got.Container, tt.wantParams.Container)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
