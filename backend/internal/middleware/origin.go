package middleware

import (
	"net"
	"net/http"
	"net/netip"
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

// TrustedProxyCIDRsFromEnv parses TRUSTED_PROXY_CIDRS as a comma-separated list of CIDRs.
// When empty, forwarded headers (e.g. X-Forwarded-Host) are treated as untrusted.
func TrustedProxyCIDRsFromEnv() []netip.Prefix {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var prefixes []netip.Prefix
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(trimmed)
		if err != nil {
			continue
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func isRequestFromTrustedProxy(r *http.Request) bool {
	ip, ok := remoteIP(r)
	if !ok {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	for _, p := range TrustedProxyCIDRsFromEnv() {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteIP(r *http.Request) (netip.Addr, bool) {
	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return netip.Addr{}, false
	}
	if h, _, err := net.SplitHostPort(host); err == nil && h != "" {
		host = h
	}
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return ip, true
}

// EffectiveRequestHost returns the host to compare against for same-origin checks.
// X-Forwarded-Host is honored only when the request comes from a trusted proxy
// (see TRUSTED_PROXY_CIDRS), to prevent client-controlled header spoofing.
func EffectiveRequestHost(r *http.Request) string {
	host := r.Host
	if isRequestFromTrustedProxy(r) {
		if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
			// X-Forwarded-Host may contain multiple values, take the first.
			if first, _, ok := strings.Cut(forwardedHost, ","); ok {
				forwardedHost = strings.TrimSpace(first)
			}
			host = forwardedHost
		}
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
	allowed := AllowedOriginsFromEnv()
	if os.Getenv("GO_ENV") == "production" && len(allowed) == 0 {
		// Fail closed in production if the allowlist is not configured. Relying on Host matching
		// is error-prone when the service is reachable directly or through multiple hostnames.
		return false
	}
	return IsOriginAllowed(origin, allowed, EffectiveRequestHost(r))
}

// IsRequestOriginOrRefererAllowed validates Origin (preferred) or Referer (fallback) against ALLOWED_ORIGINS or same-origin.
func IsRequestOriginOrRefererAllowed(r *http.Request) bool {
	allowed := AllowedOriginsFromEnv()
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" {
		if os.Getenv("GO_ENV") == "production" && len(allowed) == 0 {
			return false
		}
		return IsOriginAllowed(origin, allowed, EffectiveRequestHost(r))
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
	if os.Getenv("GO_ENV") == "production" && len(allowed) == 0 {
		return false
	}
	return IsOriginAllowed(refOrigin, allowed, EffectiveRequestHost(r))
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
