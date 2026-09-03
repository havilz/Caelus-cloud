package network

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type UseCase struct {
	repo domain.NetworkRepository
}

func NewUseCase(repo domain.NetworkRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) CreateNetwork(ctx context.Context, orgID uuid.UUID, req domain.CreateNetworkRequest) (*domain.Network, error) {
	name := strings.TrimSpace(strings.ToLower(req.Name))
	name = strings.ReplaceAll(name, " ", "-")

	gateway := req.Gateway
	if gateway == "" {
		if strings.Contains(req.CIDR, ".0.0/16") {
			gateway = strings.Replace(req.CIDR, ".0.0/16", ".0.1", 1)
		} else if strings.Contains(req.CIDR, ".0/24") {
			gateway = strings.Replace(req.CIDR, ".0/24", ".1", 1)
		} else {
			gateway = "10.20.0.1"
		}
	}

	driver := req.Driver
	if driver == "" {
		driver = "bridge"
	}

	net := &domain.Network{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		Name:            name,
		Type:            req.Type,
		CIDR:            req.CIDR,
		Gateway:         gateway,
		Region:          req.Region,
		Driver:          driver,
		Status:          domain.NetworkStatusActive,
		AttachedServers: 0,
	}

	if _, err := exec.LookPath("docker"); err == nil {
		dockerNetName := fmt.Sprintf("caelus-%s", name)
		_ = exec.CommandContext(ctx, "docker", "network", "create", "--driver", "bridge", "--subnet", req.CIDR, dockerNetName).Run()
	}

	if err := u.repo.CreateNetwork(ctx, net); err != nil {
		return nil, fmt.Errorf("gagal menyimpan network ke database: %w", err)
	}

	return net, nil
}

func (u *UseCase) ListNetworks(ctx context.Context, orgID uuid.UUID) ([]domain.Network, error) {
	return u.repo.ListNetworksByOrg(ctx, orgID)
}

func (u *UseCase) DeleteNetwork(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	net, err := u.repo.GetNetworkByID(ctx, id)
	if err == nil && net != nil {
		if _, lookErr := exec.LookPath("docker"); lookErr == nil {
			dockerNetName := fmt.Sprintf("caelus-%s", net.Name)
			_ = exec.CommandContext(ctx, "docker", "network", "rm", dockerNetName).Run()
		}
	}
	return u.repo.DeleteNetwork(ctx, id)
}

func (u *UseCase) CreateFirewallRule(ctx context.Context, orgID uuid.UUID, req domain.CreateFirewallRuleRequest) (*domain.FirewallRule, error) {
	rule := &domain.FirewallRule{
		ID:             uuid.New(),
		OrganizationID: orgID,
		NetworkID:      req.NetworkID,
		Name:           strings.TrimSpace(req.Name),
		Direction:      req.Direction,
		Protocol:       req.Protocol,
		PortRange:      strings.TrimSpace(req.PortRange),
		Source:         strings.TrimSpace(req.Source),
		Action:         req.Action,
		Status:         "active",
	}

	if err := u.repo.CreateFirewallRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("gagal menyimpan firewall rule: %w", err)
	}

	return rule, nil
}

func (u *UseCase) ListFirewallRules(ctx context.Context, orgID uuid.UUID) ([]domain.FirewallRule, error) {
	return u.repo.ListFirewallRulesByOrg(ctx, orgID)
}

func (u *UseCase) DeleteFirewallRule(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	return u.repo.DeleteFirewallRule(ctx, id)
}
