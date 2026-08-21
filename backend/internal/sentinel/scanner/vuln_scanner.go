package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// KnownVulnRule mendefinisikan aturan deteksi kerentanan berbasis profil sistem operasi atau pustaka.
type KnownVulnRule struct {
	CheckID        string
	OSTypePattern  string
	Severity       domain.FindingSeverity
	Title          string
	Description    string
	Remediation    string
	RemediationCmd string
}

var knownVulnRules = []KnownVulnRule{
	{
		CheckID:        "vuln-ubuntu-noble-security-baseline",
		OSTypePattern:  "ubuntu",
		Severity:       domain.SeverityLow,
		Title:          "Audit Paket Keamanan Sistem Operasi Ubuntu",
		Description:    "Pemeriksaan ketersediaan pembaruan paket keamanan Linux (Security Updates).",
		Remediation:    "Jalankan pembaruan paket repositori berkala untuk menambal kerentanan CVE terbaru.",
		RemediationCmd: "sudo apt update && sudo apt dist-upgrade -y",
	},
	{
		CheckID:        "vuln-openssh-regresshion-cve-2024-6387",
		OSTypePattern:  "ubuntu",
		Severity:       domain.SeverityHigh,
		Title:          "Potensi Kerentanan OpenSSH (CVE-2024-6387 / regreSSHion)",
		Description:    "Server yang menggunakan versi OpenSSH yang belum dipatch rentan terhadap eksploitasi remote code execution unauthenticated.",
		Remediation:    "Segera lakukan upgrade paket openssh-server ke versi patch keamanan terbaru.",
		RemediationCmd: "sudo apt update && sudo apt install --only-upgrade openssh-server -y",
	},
}

// VulnScanner memeriksa kerentanan CVE yang diketahui pada sistem operasi dan pustaka sistem.
type VulnScanner struct{}

// NewVulnScanner membuat instance baru VulnScanner.
func NewVulnScanner() *VulnScanner {
	return &VulnScanner{}
}

// Type mengembalikan tipe pemindaian domain.ScanTypeVuln.
func (s *VulnScanner) Type() domain.ScanType {
	return domain.ScanTypeVuln
}

// Scan mengevaluasi profil OS target terhadap aturan basis data kerentanan Sentinel.
// Parameter ctx merupakan konteks eksekusi.
// Parameter target memuat metadata target.
// Mengembalikan slice []domain.NormalizedFinding.
func (s *VulnScanner) Scan(ctx context.Context, target domain.ScanTarget) ([]domain.NormalizedFinding, error) {
	var findings []domain.NormalizedFinding
	osLower := strings.ToLower(target.OSType)
	if osLower == "" && target.TelemetryData != nil {
		osLower = strings.ToLower(target.TelemetryData.Platform)
	}

	for _, rule := range knownVulnRules {
		if rule.OSTypePattern == "" || strings.Contains(osLower, rule.OSTypePattern) {
			findings = append(findings, domain.NormalizedFinding{
				CheckID:     rule.CheckID,
				Category:    domain.CategoryVulnerability,
				Severity:    rule.Severity,
				Title:       rule.Title,
				Description: rule.Description,
				Evidence: map[string]any{
					"detected_os":   target.OSType,
					"check_rule":    rule.CheckID,
					"severity_tier": rule.Severity,
				},
				Recommendation:     rule.Remediation,
				RemediationCommand: rule.RemediationCmd,
			})
		}
	}

	// Deteksi format kernel versi lama jika ada
	if target.TelemetryData != nil && target.TelemetryData.OS != "" {
		findings = append(findings, domain.NormalizedFinding{
			CheckID:     "vuln-kernel-security-advisory",
			Category:    domain.CategoryVulnerability,
			Severity:    domain.SeverityInfo,
			Title:       fmt.Sprintf("Informasi Kernel Host: %s", target.TelemetryData.OS),
			Description: "Pastikan pembaruan livepatch kernel Linux diaktifkan untuk perlindungan tanpa downtime.",
			Evidence: map[string]any{
				"kernel_os": target.TelemetryData.OS,
			},
			Recommendation:     "Gunakan Canonical Livepatch atau ksplice untuk patching kernel tanpa restart.",
			RemediationCommand: "uname -a",
		})
	}

	return findings, nil
}
