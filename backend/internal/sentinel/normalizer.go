package sentinel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// FindingNormalizer menstandardisasi output temuan dari berbagai worker pemindai dan menghasilkan fingerprint deduplikasi.
type FindingNormalizer struct{}

// NewFindingNormalizer membuat instance baru FindingNormalizer.
func NewFindingNormalizer() *FindingNormalizer {
	return &FindingNormalizer{}
}

// Normalize mengubah DTO NormalizedFinding menjadi domain entity SecurityFinding yang siap disimpan ke database.
// Parameter orgID merupakan UUID organisasi pemilik resource.
// Parameter serverID merupakan pointer UUID server target (opsional).
// Parameter scanID merupakan pointer UUID pemindaian (opsional).
// Parameter item memuat data temuan dari scanner.
// Mengembalikan pointer *domain.SecurityFinding.
func (n *FindingNormalizer) Normalize(
	orgID uuid.UUID,
	serverID *uuid.UUID,
	scanID *uuid.UUID,
	item domain.NormalizedFinding,
) (*domain.SecurityFinding, error) {
	// Buat fingerprint unik untuk deduplikasi
	serverKey := "global"
	if serverID != nil {
		serverKey = serverID.String()
	}
	rawFingerprint := fmt.Sprintf("%s:%s:%s", item.Category, item.CheckID, serverKey)
	hash := sha256.Sum256([]byte(rawFingerprint))
	fingerprint := hex.EncodeToString(hash[:16])

	evidenceBytes, err := json.Marshal(item.Evidence)
	if err != nil {
		evidenceBytes = []byte("{}")
	}

	now := time.Now().UTC()
	finding := &domain.SecurityFinding{
		ID:                 uuid.New(),
		OrganizationID:     orgID,
		ServerID:           serverID,
		ScanID:             scanID,
		Fingerprint:        fingerprint,
		Category:           item.Category,
		Severity:           item.Severity,
		Title:              item.Title,
		Description:        item.Description,
		Evidence:           evidenceBytes,
		Recommendation:     item.Recommendation,
		RemediationCommand: item.RemediationCommand,
		Status:             domain.FindingStatusOpen,
		FirstDetectedAt:    now,
		LastDetectedAt:     now,
	}

	return finding, nil
}
