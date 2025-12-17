package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestIsRequestOriginAllowed_RequiresAllowlistInProduction(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("TRUSTED_PROXY_CIDRS", "")

	r := httptest.NewRequest("GET", "http://example.com/", nil)
	r.Host = "dkonsole.example.com"
	r.RemoteAddr = "203.0.113.10:12345"
	r.Header.Set("Origin", "https://dkonsole.example.com")

	if IsRequestOriginAllowed(r) {
		t.Fatalf("expected IsRequestOriginAllowed() to be false when ALLOWED_ORIGINS is empty in production")
	}
}

func TestIsRequestOriginOrRefererAllowed_RequiresAllowlistInProduction(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("TRUSTED_PROXY_CIDRS", "")

	r := httptest.NewRequest("POST", "http://example.com/", nil)
	r.Host = "dkonsole.example.com"
	r.RemoteAddr = "203.0.113.10:12345"
	r.Header.Set("Referer", "https://dkonsole.example.com/some/path")

	if IsRequestOriginOrRefererAllowed(r) {
		t.Fatalf("expected IsRequestOriginOrRefererAllowed() to be false when ALLOWED_ORIGINS is empty in production")
	}
}

func TestEffectiveRequestHost_DoesNotTrustXForwardedHostByDefault(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "")

	r := httptest.NewRequest("GET", "http://example.com/", nil)
	r.Host = "api.internal.local:8080"
	r.RemoteAddr = "203.0.113.10:12345"
	r.Header.Set("X-Forwarded-Host", "evil.example")

	got := EffectiveRequestHost(r)
	if got != "api.internal.local" {
		t.Fatalf("EffectiveRequestHost() = %q, want %q", got, "api.internal.local")
	}
}

func TestEffectiveRequestHost_TrustsXForwardedHostFromTrustedProxy(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")

	r := httptest.NewRequest("GET", "http://example.com/", nil)
	r.Host = "api.internal.local:8080"
	r.RemoteAddr = "10.1.2.3:54321"
	r.Header.Set("X-Forwarded-Host", "dkonsole.example.com")

	got := EffectiveRequestHost(r)
	if got != "dkonsole.example.com" {
		t.Fatalf("EffectiveRequestHost() = %q, want %q", got, "dkonsole.example.com")
	}
}

func TestEffectiveRequestHost_TrustsOnlyFirstXForwardedHostValue(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")

	r := httptest.NewRequest("GET", "http://example.com/", nil)
	r.Host = "api.internal.local:8080"
	r.RemoteAddr = "10.1.2.3:54321"
	r.Header.Set("X-Forwarded-Host", "dkonsole.example.com, evil.example")

	got := EffectiveRequestHost(r)
	if got != "dkonsole.example.com" {
		t.Fatalf("EffectiveRequestHost() = %q, want %q", got, "dkonsole.example.com")
	}
}
