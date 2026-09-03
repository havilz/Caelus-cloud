package scanner

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type TLSScanner struct {
	timeout time.Duration
}

func NewTLSScanner(timeout time.Duration) *TLSScanner {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &TLSScanner{timeout: timeout}
}

func (s *TLSScanner) Type() domain.ScanType {
	return domain.ScanTypeTLS
}

func (s *TLSScanner) Scan(ctx context.Context, target domain.ScanTarget) ([]domain.NormalizedFinding, error) {
	ipOrHost := target.IPAddress
	if ipOrHost == "" || ipOrHost == "0.0.0.0" {
		ipOrHost = target.Hostname
	}
	if ipOrHost == "" {
		ipOrHost = "127.0.0.1"
	}

	targetAddr := net.JoinHostPort(ipOrHost, "443")
	dialer := &net.Dialer{Timeout: s.timeout}

	conn, err := tls.DialWithDialer(dialer, "tcp", targetAddr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         target.Hostname,
	})
	if err != nil {

		return []domain.NormalizedFinding{}, nil
	}
	defer conn.Close()

	var findings []domain.NormalizedFinding
	state := conn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) == 0 {
		return findings, nil
	}

	leaf := certs[0]
	now := time.Now().UTC()

	if now.After(leaf.NotAfter) {
		findings = append(findings, domain.NormalizedFinding{
			CheckID:     "tls-cert-expired",
			Category:    domain.CategoryTLS,
			Severity:    domain.SeverityCritical,
			Title:       "Sertifikat TLS/SSL Telah Kedaluwarsa",
			Description: fmt.Sprintf("Sertifikat SSL untuk subject '%s' telah kedaluwarsa pada %s.", leaf.Subject.CommonName, leaf.NotAfter.Format("2006-01-02")),
			Evidence: map[string]any{
				"subject":      leaf.Subject.CommonName,
				"issuer":       leaf.Issuer.CommonName,
				"not_after":    leaf.NotAfter,
				"days_expired": int(now.Sub(leaf.NotAfter).Hours() / 24),
			},
			Recommendation:     "Segera perbarui sertifikat SSL menggunakan Let's Encrypt (Certbot) atau sertifikat CA komersial Anda.",
			RemediationCommand: "sudo certbot renew --force-renewal",
		})
	} else if leaf.NotAfter.Sub(now) < 14*24*time.Hour {

		findings = append(findings, domain.NormalizedFinding{
			CheckID:     "tls-cert-expiring-soon",
			Category:    domain.CategoryTLS,
			Severity:    domain.SeverityHigh,
			Title:       "Sertifikat TLS/SSL Akan Segera Kedaluwarsa",
			Description: fmt.Sprintf("Sertifikat SSL untuk subject '%s' akan kedaluwarsa dalam %d hari.", leaf.Subject.CommonName, int(leaf.NotAfter.Sub(now).Hours()/24)),
			Evidence: map[string]any{
				"subject":        leaf.Subject.CommonName,
				"issuer":         leaf.Issuer.CommonName,
				"not_after":      leaf.NotAfter,
				"days_remaining": int(leaf.NotAfter.Sub(now).Hours() / 24),
			},
			Recommendation:     "Jadwalkan atau jalankan perpanjangan otomatis sertifikat SSL sebelum tanggal kedaluwarsa.",
			RemediationCommand: "sudo certbot renew",
		})
	}

	if state.Version < tls.VersionTLS12 {
		findings = append(findings, domain.NormalizedFinding{
			CheckID:     "tls-weak-version",
			Category:    domain.CategoryTLS,
			Severity:    domain.SeverityHigh,
			Title:       "Protokol TLS Lawas Terdeteksi (TLS 1.0 / 1.1)",
			Description: "Server web masih mengizinkan negosiasi koneksi dengan protokol TLS 1.0 atau 1.1 yang rentan terhadap serangan downgrade.",
			Evidence: map[string]any{
				"negotiated_version": state.Version,
				"min_recommended":    "TLSv1.2 / TLSv1.3",
			},
			Recommendation:     "Nonaktifkan dukungan TLS 1.0 dan TLS 1.1 pada konfigurasi web server (Nginx/Apache/Caddy) dan wajibkan minimal TLSv1.2.",
			RemediationCommand: "ssl_protocols TLSv1.2 TLSv1.3; # Tambahkan pada file konfigurasi Nginx",
		})
	}

	return findings, nil
}
