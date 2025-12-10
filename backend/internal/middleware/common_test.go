package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// MockFlusher implements http.Flusher for testing
type MockFlusher struct {
	http.ResponseWriter
	flushed bool
}

func (m *MockFlusher) Flush() {
	m.flushed = true
}

func TestStatusRecorder_Flush(t *testing.T) {
	mockWriter := &MockFlusher{
		ResponseWriter: httptest.NewRecorder(),
	}
	recorder := &StatusRecorder{
		ResponseWriter: mockWriter,
	}

	recorder.Flush()

	if !mockWriter.flushed {
		t.Error("StatusRecorder.Flush() did not call underlying Flush()")
	}
}

func TestChain(t *testing.T) {
	// Create a recorder to verify order
	var order []string

	mw1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1 start")
			next(w, r)
			order = append(order, "mw1 end")
		}
	}

	mw2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2 start")
			next(w, r)
			order = append(order, "mw2 end")
		}
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}

	chained := Chain(handler, mw1, mw2)
	chained(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	expected := []string{
		"mw2 start",
		"mw1 start",
		"handler",
		"mw1 end",
		"mw2 end",
	}

	if !reflect.DeepEqual(order, expected) {
		t.Errorf("Chain() execution order = %v, want %v", order, expected)
	}
}
