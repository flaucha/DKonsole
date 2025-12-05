package logo

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubStorage satisfies LogoStorage for handler tests
type stubStorage struct {
	ensureCalled bool
	saved        bool
	removed      bool
	content      []byte
}

func (s *stubStorage) EnsureDataDir(ctx context.Context) error {
	s.ensureCalled = true
	return nil
}

func (s *stubStorage) Save(ctx context.Context, logoType string, ext string, content io.Reader) error {
	s.saved = true
	s.content, _ = io.ReadAll(content)
	return nil
}

func (s *stubStorage) Get(ctx context.Context, logoType string, ext string) (string, error) {
	return "", io.EOF
}

func (s *stubStorage) RemoveAll(ctx context.Context, logoType string) error {
	s.removed = true
	return nil
}

func (s *stubStorage) GetLogoContent(ctx context.Context, logoType string, ext string) ([]byte, error) {
	if len(s.content) > 0 {
		return s.content, nil
	}
	return nil, io.EOF
}

func TestUploadLogoHandler_Success(t *testing.T) {
	storage := &stubStorage{}
	validator := &mockLogoValidator{
		validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
			return ".png", nil
		},
	}
	logoSvc := NewLogoService(validator, storage)
	svc := &Service{logoService: logoSvc, dataDir: "./data"}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("logo", "logo.png")
	part.Write([]byte{0x1, 0x2, 0x3})
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/logo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	svc.UploadLogo(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !storage.ensureCalled || !storage.saved || !storage.removed {
		t.Fatalf("storage methods were not called as expected")
	}
}

func TestUploadLogoHandler_InvalidType(t *testing.T) {
	logoSvc := NewLogoService(&mockLogoValidator{}, &stubStorage{})
	svc := &Service{logoService: logoSvc, dataDir: "./data"}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("logo", "logo.png")
	part.Write([]byte{0x1})
	writer.WriteField("type", "weird")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/logo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	svc.UploadLogo(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetLogoHandler_ConfigMapContent(t *testing.T) {
	storage := &stubStorage{content: []byte{0x1, 0x2}}
	logoSvc := NewLogoService(&mockLogoValidator{}, storage)
	svc := &Service{logoService: logoSvc, dataDir: "./data"}

	req := httptest.NewRequest(http.MethodGet, "/logo?type=normal", nil)
	rr := httptest.NewRecorder()

	svc.GetLogo(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("unexpected content type: %s", ct)
	}
	if !strings.Contains(rr.Body.String(), string([]byte{0x1, 0x2})) {
		t.Fatalf("unexpected body")
	}
}

func TestGetLogoHandler_NotFound(t *testing.T) {
	logoSvc := NewLogoService(&mockLogoValidator{}, &stubStorage{})
	svc := &Service{logoService: logoSvc, dataDir: "./data"}

	req := httptest.NewRequest(http.MethodGet, "/logo?type=normal", nil)
	rr := httptest.NewRecorder()

	svc.GetLogo(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetLogoHandler_InvalidType(t *testing.T) {
	logoSvc := NewLogoService(&mockLogoValidator{}, &stubStorage{})
	svc := &Service{logoService: logoSvc, dataDir: "./data"}

	req := httptest.NewRequest(http.MethodGet, "/logo?type=invalid", nil)
	rr := httptest.NewRecorder()

	svc.GetLogo(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
