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

type FindingNormalizer struct{}

func NewFindingNormalizer() *FindingNormalizer {
	return &FindingNormalizer{}
}

func (n *FindingNormalizer) Normalize(
	orgID uuid.UUID,
	serverID *uuid.UUID,
	scanID *uuid.UUID,
	item domain.NormalizedFinding,
) (*domain.SecurityFinding, error) {

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
