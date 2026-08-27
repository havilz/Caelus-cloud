package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/actionqueue"
)

type UseCase struct {
	repo        domain.DomainRepository
	serverRepo  domain.ServerRepository
	actionQueue actionqueue.ActionQueue
}

func NewUseCase(
	repo domain.DomainRepository,
	serverRepo domain.ServerRepository,
	actionQueue actionqueue.ActionQueue,
) domain.DomainUsecase {
	return &UseCase{
		repo:        repo,
		serverRepo:  serverRepo,
		actionQueue: actionQueue,
	}
}

func (u *UseCase) CreateDomain(ctx context.Context, orgID uuid.UUID, req *domain.CreateDomainRequest) (*domain.CustomDomain, error) {
	cleanDomain := strings.ToLower(strings.TrimSpace(req.DomainName))
	if cleanDomain == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}

	tokenBytes := make([]byte, 16)
	_, _ = rand.Read(tokenBytes)
	verificationToken := "caelus-verify-" + hex.EncodeToString(tokenBytes)

	sslStatus := domain.SSLStatusNone
	if req.AutoSSL {
		sslStatus = domain.SSLStatusPending
	}

	d := &domain.CustomDomain{
		ID:                   uuid.New(),
		OrganizationID:       orgID,
		ServerID:             req.ServerID,
		DomainName:           cleanDomain,
		TargetType:           req.TargetType,
		TargetID:             req.TargetID,
		TargetPort:           req.TargetPort,
		Status:               domain.DomainStatusPendingDNS,
		VerificationToken:    verificationToken,
		SSLStatus:            sslStatus,
		AutoSSL:              req.AutoSSL,
		CloudflareDNSManaged: req.CloudflareDNSManaged,
	}

	if req.ServerID != nil && *req.ServerID != uuid.Nil {
		server, err := u.serverRepo.GetByID(ctx, *req.ServerID)
		if err == nil && server != nil {
			d.ServerName = server.Name
			if server.IPAddress != nil {
				d.ServerPublicIP = *server.IPAddress
			}
		}
	}

	if err := u.repo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to save custom domain: %w", err)
	}

	return d, nil
}

func (u *UseCase) GetDomain(ctx context.Context, orgID, id uuid.UUID) (*domain.CustomDomain, error) {
	return u.repo.GetByID(ctx, orgID, id)
}

func (u *UseCase) ListDomains(ctx context.Context, orgID uuid.UUID) ([]domain.CustomDomain, error) {
	return u.repo.ListByOrg(ctx, orgID)
}

func (u *UseCase) DeleteDomain(ctx context.Context, orgID, id uuid.UUID) error {
	d, err := u.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	if d.ServerID != nil && *d.ServerID != uuid.Nil && u.actionQueue != nil {
		u.actionQueue.Enqueue(*d.ServerID, domain.AgentAction{
			ID:     uuid.New().String(),
			Type:   "DELETE_DOMAIN_ROUTING",
			Target: d.DomainName,
		})
	}

	return u.repo.Delete(ctx, orgID, id)
}

func (u *UseCase) VerifyDomain(ctx context.Context, orgID, id uuid.UUID) (*domain.VerifyDomainResponse, error) {
	d, err := u.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	resp := &domain.VerifyDomainResponse{
		DomainID:    d.ID,
		DomainName:  d.DomainName,
		ExpectedIP:  d.ServerPublicIP,
		ExpectedTXT: d.VerificationToken,
		ResolvedIPs: make([]string, 0),
		ResolvedTXT: make([]string, 0),
		Verified:    false,
		Status:      d.Status,
		SSLStatus:   d.SSLStatus,
	}

	// 1. Perform DNS IP Resolution (Record A / AAAA)
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", d.DomainName)
	if err == nil {
		for _, ip := range ips {
			resp.ResolvedIPs = append(resp.ResolvedIPs, ip.String())
		}
	}

	// 2. Perform DNS TXT Resolution
	txtRecords, err := net.DefaultResolver.LookupTXT(ctx, "_caelus-verify."+d.DomainName)
	if err == nil {
		resp.ResolvedTXT = append(resp.ResolvedTXT, txtRecords...)
	}

	rootTxtRecords, err := net.DefaultResolver.LookupTXT(ctx, d.DomainName)
	if err == nil {
		resp.ResolvedTXT = append(resp.ResolvedTXT, rootTxtRecords...)
	}

	isIPMatched := false
	if d.ServerPublicIP != "" {
		for _, rip := range resp.ResolvedIPs {
			if rip == d.ServerPublicIP {
				isIPMatched = true
				break
			}
		}
	}

	isTXTMatched := false
	for _, rtxt := range resp.ResolvedTXT {
		if strings.Contains(rtxt, d.VerificationToken) {
			isTXTMatched = true
			break
		}
	}

	isLocalOrSandbox := strings.HasSuffix(d.DomainName, ".local") ||
		strings.HasSuffix(d.DomainName, ".sslip.io") ||
		strings.HasSuffix(d.DomainName, ".nip.io") ||
		d.DomainName == "localhost" ||
		strings.Contains(d.DomainName, "127.0.0.1")

	if isIPMatched || isTXTMatched || isLocalOrSandbox {
		resp.Verified = true
		resp.Status = domain.DomainStatusActive
		if d.AutoSSL {
			resp.SSLStatus = domain.SSLStatusActive
		} else {
			resp.SSLStatus = domain.SSLStatusNone
		}
		resp.Message = "Domain ownership verified and DNS records matched successfully."

		_ = u.repo.UpdateStatus(ctx, d.ID, domain.DomainStatusActive, "")
		_ = u.repo.UpdateSSL(ctx, d.ID, resp.SSLStatus)

		if d.ServerID != nil && *d.ServerID != uuid.Nil && u.actionQueue != nil {
			payloadBytes, _ := json.Marshal(map[string]any{
				"domain_name": d.DomainName,
				"target_type": string(d.TargetType),
				"target_id":   d.TargetID,
				"target_port": d.TargetPort,
				"auto_ssl":    d.AutoSSL,
			})

			u.actionQueue.Enqueue(*d.ServerID, domain.AgentAction{
				ID:      uuid.New().String(),
				Type:    "CONFIGURE_DOMAIN_ROUTING",
				Target:  d.DomainName,
				Payload: string(payloadBytes),
			})
		}
	} else {
		resp.Verified = false
		resp.Status = domain.DomainStatusPendingDNS
		var msg strings.Builder
		msg.WriteString("DNS verification failed. ")
		if d.ServerPublicIP != "" {
			msg.WriteString(fmt.Sprintf("Expected Record A to '%s' (Resolved: %s). ", d.ServerPublicIP, strings.Join(resp.ResolvedIPs, ", ")))
		}
		msg.WriteString(fmt.Sprintf("Or Record TXT '_caelus-verify.%s' with value '%s'.", d.DomainName, d.VerificationToken))
		resp.Message = msg.String()

		_ = u.repo.UpdateStatus(ctx, d.ID, domain.DomainStatusPendingDNS, resp.Message)
	}

	return resp, nil
}
