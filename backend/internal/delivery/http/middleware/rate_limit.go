package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
)

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

type AuthRateLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	entries     map[string]*rateLimitEntry
	lastCleanup time.Time
}

func NewAuthRateLimiter(limit int, window time.Duration) *AuthRateLimiter {
	limiter := &AuthRateLimiter{
		limit:       limit,
		window:      window,
		entries:     make(map[string]*rateLimitEntry),
		lastCleanup: time.Now(),
	}
	return limiter
}

func (l *AuthRateLimiter) Limit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			email := extractEmailFromPayload(r)

			key := ip
			if email != "" {
				key = ip + ":" + strings.ToLower(email)
			}

			if !l.allow(key) {
				response.Error(w, http.StatusTooManyRequests, "Terlalu banyak percobaan autentikasi", "Batas percobaan terlampaui. Silakan coba lagi dalam 1 menit.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *AuthRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	if now.Sub(l.lastCleanup) > 5*time.Minute {
		for k, v := range l.entries {
			if now.After(v.resetTime) {
				delete(l.entries, k)
			}
		}
		l.lastCleanup = now
	}

	entry, exists := l.entries[key]
	if !exists || now.After(entry.resetTime) {
		l.entries[key] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(l.window),
		}
		return true
	}

	if entry.count >= l.limit {
		return false
	}

	entry.count++
	return true
}

func extractEmailFromPayload(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err == nil && payload.Email != "" {
		return strings.TrimSpace(payload.Email)
	}

	return ""
}
