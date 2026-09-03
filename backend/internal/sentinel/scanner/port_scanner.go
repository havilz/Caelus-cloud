package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type CommonPortDef struct {
	Port        int
	Service     string
	IsHighRisk  bool
	Severity    domain.FindingSeverity
	Description string
	Remediation string
}

var commonPorts = []CommonPortDef{
	{Port: 21, Service: "FTP", IsHighRisk: true, Severity: domain.SeverityHigh, Description: "Layanan File Transfer Protocol (FTP) terbuka ke publik tanpa enkripsi.", Remediation: "Matikan service FTP dan gunakan SFTP (SSH File Transfer) yang terenkripsi."},
	{Port: 23, Service: "Telnet", IsHighRisk: true, Severity: domain.SeverityCritical, Description: "Layanan Telnet yang tidak terenkripsi terbuka ke publik.", Remediation: "Hentikan daemon Telnet dan gunakan SSH untuk manajemen jarak jauh."},
	{Port: 22, Service: "SSH", IsHighRisk: false, Severity: domain.SeverityInfo, Description: "Port SSH terbuka untuk akses administrasi.", Remediation: "Pastikan autentikasi berbasis SSH Key aktif dan disable password auth."},
	{Port: 80, Service: "HTTP", IsHighRisk: false, Severity: domain.SeverityLow, Description: "Port HTTP terbuka ke publik.", Remediation: "Konfigurasikan pengalihan otomatis (redirect 301) dari HTTP ke HTTPS."},
	{Port: 443, Service: "HTTPS", IsHighRisk: false, Severity: domain.SeverityInfo, Description: "Port HTTPS web aman terbuka.", Remediation: "Pertahankan pembaruan sertifikat TLS dan cipher suite modern."},
	{Port: 3306, Service: "MySQL/MariaDB", IsHighRisk: true, Severity: domain.SeverityHigh, Description: "Database MySQL/MariaDB terbuka ke jaringan publik.", Remediation: "Batasi listen address database ke localhost (127.0.0.1) atau gunakan firewall UFW."},
	{Port: 5432, Service: "PostgreSQL", IsHighRisk: true, Severity: domain.SeverityHigh, Description: "Database PostgreSQL terbuka ke jaringan publik.", Remediation: "Ubah pg_hba.conf dan bind listen_addresses ke '127.0.0.1' atau private VPN."},
	{Port: 6379, Service: "Redis", IsHighRisk: true, Severity: domain.SeverityCritical, Description: "In-memory cache Redis terbuka ke publik.", Remediation: "Aktifkan requirepass, bind ke 127.0.0.1, dan blokir port 6379 dari internet eksternal."},
	{Port: 27017, Service: "MongoDB", IsHighRisk: true, Severity: domain.SeverityCritical, Description: "Database NoSQL MongoDB terbuka ke publik.", Remediation: "Aktifkan autentikasi dan bind IP ke localhost atau jaringan privat."},
	{Port: 11211, Service: "Memcached", IsHighRisk: true, Severity: domain.SeverityCritical, Description: "Cache Memcached tanpa otentikasi terbuka ke jaringan publik.", Remediation: "Tutup port 11211 dari akses publik dan aktifkan flag '-l 127.0.0.1'."},
	{Port: 2375, Service: "Docker Unencrypted API", IsHighRisk: true, Severity: domain.SeverityCritical, Description: "Docker Engine Daemon TCP Socket terbuka tanpa enkripsi TLS (Potensi Container Escape / Remote Root Takeover).", Remediation: "Matikan daemon TCP port 2375 atau aktifkan TLS Mutual Auth (port 2376)."},
	{Port: 9200, Service: "Elasticsearch", IsHighRisk: true, Severity: domain.SeverityHigh, Description: "Endpoint HTTP REST API Elasticsearch terbuka ke publik.", Remediation: "Aktifkan xpack.security.enabled dan kunci akses port 9200 dengan UFW/IPtables."},
	{Port: 5900, Service: "VNC Remote Desktop", IsHighRisk: true, Severity: domain.SeverityHigh, Description: "Layanan VNC Remote Desktop terbuka ke jaringan publik.", Remediation: "Gunakan VNC melalui SSH Tunnel atau matikan service jika tidak digunakan."},
	{Port: 8080, Service: "HTTP-Alt", IsHighRisk: false, Severity: domain.SeverityLow, Description: "Port alternatif HTTP terbuka.", Remediation: "Pastikan aplikasi di port 8080 memiliki otentikasi yang memadai."},
}

type PortScanner struct {
	dialTimeout time.Duration
}

func NewPortScanner(dialTimeout time.Duration) *PortScanner {
	if dialTimeout <= 0 {
		dialTimeout = 800 * time.Millisecond
	}
	return &PortScanner{dialTimeout: dialTimeout}
}

func (s *PortScanner) Type() domain.ScanType {
	return domain.ScanTypePort
}

func (s *PortScanner) Scan(ctx context.Context, target domain.ScanTarget) ([]domain.NormalizedFinding, error) {
	ipOrHost := target.IPAddress
	if ipOrHost == "" || ipOrHost == "0.0.0.0" {
		ipOrHost = target.Hostname
	}
	if ipOrHost == "" {
		ipOrHost = "127.0.0.1"
	}

	var findings []domain.NormalizedFinding
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, portDef := range commonPorts {
		wg.Add(1)
		go func(p CommonPortDef) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			targetAddr := net.JoinHostPort(ipOrHost, fmt.Sprintf("%d", p.Port))
			d := net.Dialer{Timeout: s.dialTimeout}
			conn, err := d.DialContext(ctx, "tcp", targetAddr)
			if err == nil {
				_ = conn.Close()

				finding := domain.NormalizedFinding{
					CheckID:     fmt.Sprintf("port-%d-exposure", p.Port),
					Category:    domain.CategoryNetwork,
					Severity:    p.Severity,
					Title:       fmt.Sprintf("Port Terbuka: %d (%s)", p.Port, p.Service),
					Description: p.Description,
					Evidence: map[string]any{
						"port":       p.Port,
						"service":    p.Service,
						"target":     targetAddr,
						"state":      "open",
						"scanned_at": time.Now().UTC(),
					},
					Recommendation:     p.Remediation,
					RemediationCommand: fmt.Sprintf("sudo ufw deny %d/tcp", p.Port),
				}

				mu.Lock()
				findings = append(findings, finding)
				mu.Unlock()
			}
		}(portDef)
	}

	wg.Wait()
	return findings, nil
}
