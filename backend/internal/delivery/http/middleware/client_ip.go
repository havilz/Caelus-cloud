package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
)

type clientIPContextKeyType string

const ClientIPContextKey clientIPContextKeyType = "client_ip"

var (
	defaultValidatorMu sync.RWMutex
	defaultValidator   = NewIPValidator([]string{"127.0.0.1", "::1"})
)

// SetTrustedProxies sets the global trusted proxies configuration.
func SetTrustedProxies(proxies []string) {
	defaultValidatorMu.Lock()
	defer defaultValidatorMu.Unlock()
	defaultValidator = NewIPValidator(proxies)
}

// GetIPValidator returns the active global IPValidator.
func GetIPValidator() *IPValidator {
	defaultValidatorMu.RLock()
	defer defaultValidatorMu.RUnlock()
	return defaultValidator
}

// IPValidator validates whether remote addresses originate from trusted proxies and safely extracts client IPs.
type IPValidator struct {
	trustedIPs   map[string]bool
	trustedCIDRs []*net.IPNet
}

// NewIPValidator parses IP addresses and CIDR networks into an IPValidator.
func NewIPValidator(trusted []string) *IPValidator {
	v := &IPValidator{
		trustedIPs:   make(map[string]bool),
		trustedCIDRs: make([]*net.IPNet, 0),
	}

	for _, entry := range trusted {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if strings.Contains(entry, "/") {
			_, ipNet, err := net.ParseCIDR(entry)
			if err == nil && ipNet != nil {
				v.trustedCIDRs = append(v.trustedCIDRs, ipNet)
			}
			continue
		}

		if parsed := net.ParseIP(entry); parsed != nil {
			v.trustedIPs[parsed.String()] = true
		}
	}

	return v
}

// IsTrustedProxy returns true if the specified IP string matches a configured trusted proxy or CIDR block.
func (v *IPValidator) IsTrustedProxy(ipStr string) bool {
	parsed := net.ParseIP(ipStr)
	if parsed == nil {
		return false
	}

	if v.trustedIPs[parsed.String()] {
		return true
	}

	for _, cidr := range v.trustedCIDRs {
		if cidr.Contains(parsed) {
			return true
		}
	}

	return false
}

// ExtractClientIP securely extracts the client IP address from the request.
func (v *IPValidator) ExtractClientIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	remoteHost = strings.TrimSpace(remoteHost)

	if !v.IsTrustedProxy(remoteHost) {
		return remoteHost
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			if candidate == "" {
				continue
			}
			parsed := net.ParseIP(candidate)
			if parsed == nil {
				continue
			}
			if !v.IsTrustedProxy(candidate) {
				return candidate
			}
		}

		for _, part := range parts {
			candidate := strings.TrimSpace(part)
			if parsed := net.ParseIP(candidate); parsed != nil {
				return candidate
			}
		}
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		candidate := strings.TrimSpace(xrip)
		if parsed := net.ParseIP(candidate); parsed != nil {
			return candidate
		}
	}

	return remoteHost
}

// ExtractClientIP extracts the validated client IP from the request using context or global validator.
func ExtractClientIP(r *http.Request) string {
	if ctxIP, ok := r.Context().Value(ClientIPContextKey).(string); ok && ctxIP != "" {
		return ctxIP
	}

	return GetIPValidator().ExtractClientIP(r)
}

// ClientIPMiddleware creates a middleware that validates client IP, injects it into context, and sanitizes RemoteAddr.
func ClientIPMiddleware(validator *IPValidator) func(http.Handler) http.Handler {
	if validator == nil {
		validator = GetIPValidator()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := validator.ExtractClientIP(r)
			ctx := context.WithValue(r.Context(), ClientIPContextKey, clientIP)

			_, port, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil && port != "" {
				r.RemoteAddr = net.JoinHostPort(clientIP, port)
			} else {
				r.RemoteAddr = clientIP
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
