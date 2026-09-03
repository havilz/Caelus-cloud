package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type AlertEvaluator struct {
	alertRepo   domain.AlertRepository
	broadcaster domain.TelemetryBroadcaster
}

func NewAlertEvaluator(alertRepo domain.AlertRepository, broadcaster domain.TelemetryBroadcaster) *AlertEvaluator {
	return &AlertEvaluator{
		alertRepo:   alertRepo,
		broadcaster: broadcaster,
	}
}

func (e *AlertEvaluator) EvaluateMetrics(ctx context.Context, server *domain.Server, metric *domain.ServerMetric) error {
	rules, err := e.alertRepo.ListRulesForServer(ctx, server.OrganizationID, server.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch alert rules for evaluation: %w", err)
	}

	activeAlerts, err := e.alertRepo.ListActiveAlertsByServer(ctx, server.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch active alerts for server: %w", err)
	}

	activeRuleMap := make(map[uuid.UUID]bool)
	for _, a := range activeAlerts {
		if a.RuleID != nil {
			activeRuleMap[*a.RuleID] = true
		}
	}

	for _, rule := range rules {
		val, matched := e.checkRuleCondition(&rule, metric)
		if matched {
			if !activeRuleMap[rule.ID] {
				_ = e.triggerAlert(ctx, server, &rule, val)
			}
		}
	}

	return nil
}

func (e *AlertEvaluator) checkRuleCondition(rule *domain.AlertRule, metric *domain.ServerMetric) (float64, bool) {
	var val float64
	switch rule.MetricName {
	case "cpu_usage":
		val = metric.CPUUsagePct
	case "memory_usage":
		val = metric.MemoryUsagePct
	case "disk_usage":
		val = metric.DiskUsagePct
	default:
		return 0, false
	}

	switch rule.Operator {
	case ">":
		return val, val > rule.Threshold
	case ">=":
		return val, val >= rule.Threshold
	case "<":
		return val, val < rule.Threshold
	case "<=":
		return val, val <= rule.Threshold
	case "==":
		return val, val == rule.Threshold
	default:
		return val, val > rule.Threshold
	}
}

func (e *AlertEvaluator) triggerAlert(ctx context.Context, server *domain.Server, rule *domain.AlertRule, currentVal float64) error {
	alert := &domain.Alert{
		OrganizationID: server.OrganizationID,
		ServerID:       server.ID,
		RuleID:         &rule.ID,
		AlertType:      rule.MetricName + "_threshold",
		Severity:       rule.Severity,
		Title:          fmt.Sprintf("Alert: %s exceeded on %s", rule.Name, server.Name),
		Message:        fmt.Sprintf("Metric %s reached %.2f%% (Threshold: %.2f%%)", rule.MetricName, currentVal, rule.Threshold),
		Status:         domain.AlertStatusActive,
		CurrentValue:   &currentVal,
		ThresholdValue: &rule.Threshold,
		TriggeredAt:    time.Now().UTC(),
	}

	if err := e.alertRepo.CreateAlert(ctx, alert); err != nil {
		return err
	}

	if e.broadcaster != nil {
		e.broadcaster.BroadcastToOrg(server.OrganizationID, "alert.created", alert)
		e.broadcaster.BroadcastToServer(server.ID, "alert.created", alert)
	}

	return nil
}
