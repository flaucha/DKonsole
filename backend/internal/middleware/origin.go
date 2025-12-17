package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

// AllowedOriginsFromEnv parses ALLOWED_ORIGINS as a comma-separated list.
// Entries can be full origins (e.g. "https://example.com") or hostnames (e.g. "example.com").
func AllowedOriginsFromEnv() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var allowed []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			allowed = append(allowed, trimmed)
		}
	}
	return allowed
}

// EffectiveRequestHost returns the host to compare against for same-origin checks.
// X-Forwarded-Host is honored when present.
func EffectiveRequestHost(r *http.Request) string {
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); strings.TrimSpace(forwardedHost) != "" {
		host = forwardedHost
	}
	return normalizeHost(host)
}

// IsRequestOriginAllowed validates the request Origin header against ALLOWED_ORIGINS or same-origin.
// Empty Origin is allowed to support non-browser clients.
func IsRequestOriginAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	return IsOriginAllowed(origin, AllowedOriginsFromEnv(), EffectiveRequestHost(r))
}

// IsRequestOriginOrRefererAllowed validates Origin (preferred) or Referer (fallback) against ALLOWED_ORIGINS or same-origin.
func IsRequestOriginOrRefererAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" {
		return IsOriginAllowed(origin, AllowedOriginsFromEnv(), EffectiveRequestHost(r))
	}
	referer := strings.TrimSpace(r.Header.Get("Referer"))
	if referer == "" {
		return false
	}
	refURL, err := url.Parse(referer)
	if err != nil || refURL.Host == "" {
		return false
	}
	refOrigin := refURL.Scheme + "://" + refURL.Host
	return IsOriginAllowed(refOrigin, AllowedOriginsFromEnv(), EffectiveRequestHost(r))
}

// IsOriginAllowed checks Origin against an allowlist (full origin or hostname) or enforces same-origin when allowlist is empty.
func IsOriginAllowed(origin string, allowedList []string, requestHost string) bool {
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}
	originHost := normalizeHost(originURL.Host)

	// If allowlist is set, match against it (full origin match or host match).
	if len(allowedList) > 0 {
		for _, allowed := range allowedList {
			if strings.Contains(allowed, "*") {
				// Wildcards are intentionally unsupported for credentialed requests.
				continue
			}
			if strings.EqualFold(allowed, origin) {
				return true
			}
			if normalizeHost(allowed) == originHost {
				return true
			}
		}
		return false
	}

	// Without allowlist, require same host.
	return originHost == requestHost
}

func normalizeHost(hostport string) string {
	host := strings.TrimSpace(hostport)
	if strings.Contains(host, ":") {
		host, _, _ = strings.Cut(host, ":")
	}
	return strings.ToLower(host)
}
