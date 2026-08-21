package scanner

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// SecurityHeaderCheck mendefinisikan kriteria audit untuk setiap security header HTTP.
type SecurityHeaderCheck struct {
	HeaderName  string
	Severity    domain.FindingSeverity
	Title       string
	Description string
	Remediation string
}

var standardSecurityHeaders = []SecurityHeaderCheck{
	{
		HeaderName:  "Strict-Transport-Security",
		Severity:    domain.SeverityMedium,
		Title:       "Header HSTS (Strict-Transport-Security) Tidak Ditemukan",
		Description: "Header HTTP Strict Transport Security (HSTS) tidak aktif, membuat komunikasi browser rentan terhadap serangan downgrade SSL stripping.",
		Remediation: "Tambahkan header: add_header Strict-Transport-Security 'max-age=31536000; includeSubDomains' always;",
	},
	{
		HeaderName:  "X-Content-Type-Options",
		Severity:    domain.SeverityLow,
		Title:       "Header X-Content-Type-Options: nosniff Tidak Ditemukan",
		Description: "MIME-sniffing protection tidak diaktifkan, berpotensi memicu eksekusi berkas berbahaya yang disamarkan sebagai jenis konten lain.",
		Remediation: "Tambahkan header: add_header X-Content-Type-Options 'nosniff' always;",
	},
	{
		HeaderName:  "X-Frame-Options",
		Severity:    domain.SeverityMedium,
		Title:       "Header X-Frame-Options (Clickjacking Protection) Tidak Ditemukan",
		Description: "Situs web dapat disematkan ke dalam iframe pihak ketiga, membuka celah eksploitasi serangan Clickjacking (UI Redressing).",
		Remediation: "Tambahkan header: add_header X-Frame-Options 'SAMEORIGIN' always;",
	},
	{
		HeaderName:  "Content-Security-Policy",
		Severity:    domain.SeverityMedium,
		Title:       "Header Content-Security-Policy (CSP) Tidak Ditemukan",
		Description: "Kebijakan keamanan konten (CSP) tidak didefinisikan, mempermudah eksekusi injeksi script berbahaya (Cross-Site Scripting / XSS).",
		Remediation: "Definisikan aturan CSP yang ketat sesuai dengan domain resource aplikasi Anda.",
	},
	{
		HeaderName:  "Referrer-Policy",
		Severity:    domain.SeverityLow,
		Title:       "Header Referrer-Policy Tidak Ditemukan",
		Description: "Browser dapat membocorkan path URL dan parameter query internal saat pengguna berpindah ke tautan eksternal.",
		Remediation: "Tambahkan header: add_header Referrer-Policy 'strict-origin-when-cross-origin' always;",
	},
}

// HeadersScanner melakukan audit terhadap HTTP response headers untuk memastikan kepatuhan standar OWASP.
type HeadersScanner struct {
	client *http.Client
}

// NewHeadersScanner membuat instance baru HeadersScanner.
func NewHeadersScanner(timeout time.Duration) *HeadersScanner {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &HeadersScanner{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// Type mengembalikan tipe pemindaian domain.ScanTypeHeaders.
func (s *HeadersScanner) Type() domain.ScanType {
	return domain.ScanTypeHeaders
}

// Scan mengirimkan HTTP GET probe ke target web dan memeriksa kelengkapan header keamanan.
// Parameter ctx merupakan konteks eksekusi.
// Parameter target memuat metadata target pemindaian.
// Mengembalikan slice []domain.NormalizedFinding.
func (s *HeadersScanner) Scan(ctx context.Context, target domain.ScanTarget) ([]domain.NormalizedFinding, error) {
	ipOrHost := target.IPAddress
	if ipOrHost == "" || ipOrHost == "0.0.0.0" {
		ipOrHost = target.Hostname
	}
	if ipOrHost == "" {
		ipOrHost = "127.0.0.1"
	}

	url := fmt.Sprintf("http://%s", ipOrHost)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []domain.NormalizedFinding{}, nil
	}
	req.Header.Set("User-Agent", "Caelus-Sentinel-Auditor/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		// Target mungkin tidak menjalankan web server di port 80/443
		return []domain.NormalizedFinding{}, nil
	}
	defer resp.Body.Close()

	var findings []domain.NormalizedFinding

	for _, check := range standardSecurityHeaders {
		val := resp.Header.Get(check.HeaderName)
		if strings.TrimSpace(val) == "" {
			findings = append(findings, domain.NormalizedFinding{
				CheckID:     fmt.Sprintf("missing-header-%s", strings.ToLower(strings.ReplaceAll(check.HeaderName, "-", "_"))),
				Category:    domain.CategoryHTTPHeaders,
				Severity:    check.Severity,
				Title:       check.Title,
				Description: check.Description,
				Evidence: map[string]any{
					"target_url":     url,
					"missing_header": check.HeaderName,
					"status_code":    resp.StatusCode,
				},
				Recommendation:     check.Remediation,
				RemediationCommand: check.Remediation,
			})
		}
	}

	return findings, nil
}
