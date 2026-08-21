package sentinel

import (
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// RiskEngine bertanggung jawab untuk menghitung skor risiko keamanan (Security Score 0 - 100) dan menentukan predikat postur keamanan.
type RiskEngine struct {
	criticalPenalty int
	highPenalty     int
	mediumPenalty   int
	lowPenalty      int
}

// NewRiskEngine membuat instance baru RiskEngine dengan bobot penalti terstandarisasi.
func NewRiskEngine() *RiskEngine {
	return &RiskEngine{
		criticalPenalty: 20,
		highPenalty:     10,
		mediumPenalty:   5,
		lowPenalty:      2,
	}
}

// CalculateScore menghitung skor keamanan dari daftar temuan aktif (0 - 100).
// Parameter findings memuat slice temuan keamanan.
// Mengembalikan nilai integer skor dan jumlah temuan per tingkat keparahan.
func (e *RiskEngine) CalculateScore(findings []domain.SecurityFinding) (score int, critical, high, medium, low int) {
	score = 100

	for _, f := range findings {
		if f.Status == domain.FindingStatusResolved || f.Status == domain.FindingStatusFalsePositive {
			continue
		}

		switch f.Severity {
		case domain.SeverityCritical:
			critical++
			score -= e.criticalPenalty
		case domain.SeverityHigh:
			high++
			score -= e.highPenalty
		case domain.SeverityMedium:
			medium++
			score -= e.mediumPenalty
		case domain.SeverityLow:
			low++
			score -= e.lowPenalty
		}
	}

	if score < 0 {
		score = 0
	}
	return score, critical, high, medium, low
}

// CalculateGrade menentukan predikat huruf (Grade A/B/C/D/F) berdasarkan skor 0 - 100.
// Parameter score merupakan nilai skor keamanan.
// Mengembalikan string grade huruf ("A", "B", "C", "D", "F").
func (e *RiskEngine) CalculateGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 50:
		return "D"
	default:
		return "F"
	}
}
