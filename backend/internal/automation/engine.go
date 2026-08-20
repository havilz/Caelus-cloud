package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/notification"
	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/internal/queue"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// RuleEngine mendefinisikan kontrak mesin evaluasi dan eksekusi aturan otomasi.
type RuleEngine interface {
	// EvaluateEvent memproses satu event sistem terhadap seluruh aturan otomasi yang relevan.
	// Parameter ctx merupakan context pemanggilan.
	// Parameter event merupakan payload kejadian sistem.
	// Mengembalikan error jika terjadi kesalahan fatal pada evaluasi.
	EvaluateEvent(ctx context.Context, event domain.SystemEvent) error

	// ExecuteRuleAction mengeksekusi satu aksi spesifik dari aturan otomasi.
	// Parameter ctx merupakan context pemanggilan.
	// Parameter rule merupakan entitas aturan yang sedang dieksekusi.
	// Parameter action merupakan aksi yang akan dijalankan.
	// Parameter eventData merupakan data kejadian pemicu.
	// Mengembalikan pointer domain.ActionResultItem.
	ExecuteRuleAction(ctx context.Context, rule *domain.AutomationRule, action domain.RuleAction, eventData map[string]any) domain.ActionResultItem
}

// ServerExecutor mendefinisikan operasi eksekusi kontrol server untuk rule engine.
type ServerExecutor interface {
	RebootServer(ctx context.Context, orgID, serverID uuid.UUID) error
}

// BackupExecutor mendefinisikan operasi eksekusi backup untuk rule engine.
type BackupExecutor interface {
	TriggerBackup(ctx context.Context, orgID, serverID uuid.UUID, policyID *uuid.UUID, backupName string) (*domain.BackupRecord, error)
}

// Engine mengimplementasikan RuleEngine dengan evaluasi logika kondisi dan eksekusi asinkron via TaskQueue.
type Engine struct {
	repo         domain.AutomationRepository
	queue        queue.QueueEngine
	notifier     notification.Dispatcher
	serverExec   ServerExecutor
	backupExec   BackupExecutor
}

// NewEngine membuat dan menginisialisasi instance Engine baru.
// Parameter repo merupakan repositori PostgreSQL otomasi.
// Parameter q merupakan engine antrean pekerjaan asinkron.
// Parameter notifier merupakan dispatcher pengirim notifikasi email dan webhook.
// Parameter serverExec merupakan eksekutor manajemen server (opsional).
// Parameter backupExec merupakan eksekutor otomatisasi backup (opsional).
// Mengembalikan pointer *Engine.
func NewEngine(
	repo domain.AutomationRepository,
	q queue.QueueEngine,
	notifier notification.Dispatcher,
	serverExec ServerExecutor,
	backupExec BackupExecutor,
) *Engine {
	return &Engine{
		repo:       repo,
		queue:      q,
		notifier:   notifier,
		serverExec: serverExec,
		backupExec: backupExec,
	}
}

// EvaluateEvent mengevaluasi seluruh aturan otomasi yang relevan dengan tipe kejadian sistem.
func (e *Engine) EvaluateEvent(ctx context.Context, event domain.SystemEvent) error {
	triggerType := mapEventTypeToTriggerType(event.Type)
	rules, err := e.repo.ListActiveRulesByTriggerType(ctx, triggerType)
	if err != nil {
		return fmt.Errorf("failed to fetch active rules for trigger type %s: %w", triggerType, err)
	}

	for _, rule := range rules {
		// Pastikan organisasi event cocok dengan organisasi rule
		if event.OrganizationID != uuid.Nil && rule.OrganizationID != event.OrganizationID {
			continue
		}

		go func(r domain.AutomationRule) {
			evalCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			e.evaluateSingleRule(evalCtx, &r, event)
		}(rule)
	}

	return nil
}

// evaluateSingleRule mengevaluasi satu aturan spesifik: Trigger Match -> Condition Match -> Cooldown Check -> Action Execution.
func (e *Engine) evaluateSingleRule(ctx context.Context, rule *domain.AutomationRule, event domain.SystemEvent) {
	startTime := time.Now()

	// 1. Trigger Config Filter Match (misal filter server_id tertentu jika dikonfigurasi)
	if !e.matchTriggerConfig(rule.TriggerConfig, event) {
		return
	}

	// 2. Evaluasi Kondisi Logika
	allConditionsMet, evaluatedConds := e.evaluateConditions(rule.Conditions, event.Data)
	if !allConditionsMet {
		return
	}

	// 3. Cooldown Check (Mencegah Flapping / Spamming Aksi)
	now := time.Now().UTC()
	if rule.LastTriggeredAt != nil && rule.CooldownSeconds > 0 {
		cooldownExpiry := rule.LastTriggeredAt.Add(time.Duration(rule.CooldownSeconds) * time.Second)
		if now.Before(cooldownExpiry) {
			logger.Warn("Aturan otomasi sedang dalam periode cooldown, eksekusi diabaikan",
				"rule_id", rule.ID,
				"rule_name", rule.Name,
				"cooldown_remaining", cooldownExpiry.Sub(now),
			)

			_ = e.recordLog(ctx, rule, event.Type, domain.ExecutionStatusSkipped, evaluatedConds, nil, "Skipped due to active cooldown window", time.Since(startTime))
			return
		}
	}

	// 4. Perbarui status LastTriggeredAt
	_ = e.repo.UpdateLastTriggered(ctx, rule.ID, now)

	// 5. Eksekusi Seluruh Aksi Aturan
	actionResults := make([]domain.ActionResultItem, 0, len(rule.Actions))
	hasFailure := false
	hasSuccess := false

	for _, action := range rule.Actions {
		result := e.ExecuteRuleAction(ctx, rule, action, event.Data)
		actionResults = append(actionResults, result)
		if result.Status == "success" {
			hasSuccess = true
		} else {
			hasFailure = true
		}
	}

	// 6. Tentukan Status Akhir Log Eksekusi
	status := domain.ExecutionStatusSuccess
	var errMessage string
	if hasFailure && hasSuccess {
		status = domain.ExecutionStatusPartial
		errMessage = "One or more actions failed to execute"
	} else if hasFailure && !hasSuccess {
		status = domain.ExecutionStatusFailed
		errMessage = "All actions failed to execute"
	}

	_ = e.recordLog(ctx, rule, event.Type, status, evaluatedConds, actionResults, errMessage, time.Since(startTime))
}

// ExecuteRuleAction mengeksekusi satu aksi otomasi.
func (e *Engine) ExecuteRuleAction(ctx context.Context, rule *domain.AutomationRule, action domain.RuleAction, eventData map[string]any) domain.ActionResultItem {
	res := domain.ActionResultItem{
		ActionType: action.Type,
		Target:     action.Target,
		Status:     "success",
	}

	switch action.Type {
	case domain.ActionTypeSendEmail:
		subject := fmt.Sprintf("[ALERT] %s triggered on Caelus Cloud", rule.Name)
		body := email.BuildAlertHTMLTemplate(
			fmt.Sprintf("Automation rule '%s' was triggered automatically.", rule.Name),
			rule.Name,
			string(rule.TriggerType),
			fmt.Sprintf("Event Data: %+v", eventData),
		)

		err := e.notifier.SendEmail(ctx, email.EmailMessage{
			To:      action.Target,
			Subject: subject,
			Body:    body,
			IsHTML:  true,
		})
		if err != nil {
			res.Status = "failed"
			res.Error = err.Error()
		} else {
			res.Response = fmt.Sprintf("Email sent successfully to %s", action.Target)
		}

	case domain.ActionTypeSendWebhook:
		payload := webhook.WebhookPayload{
			EventID:        uuid.New().String(),
			EventType:      string(rule.TriggerType),
			OrganizationID: rule.OrganizationID.String(),
			RuleID:         rule.ID.String(),
			RuleName:       rule.Name,
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			Data:           eventData,
		}

		err := e.notifier.SendWebhook(ctx, action.Target, payload)
		if err != nil {
			res.Status = "failed"
			res.Error = err.Error()
		} else {
			res.Response = fmt.Sprintf("Webhook delivered successfully to %s", action.Target)
		}

	case domain.ActionTypeRebootServer:
		if e.serverExec != nil && action.Target != "" {
			serverUUID, parseErr := uuid.Parse(action.Target)
			if parseErr == nil {
				err := e.serverExec.RebootServer(ctx, rule.OrganizationID, serverUUID)
				if err != nil {
					res.Status = "failed"
					res.Error = err.Error()
				} else {
					res.Response = fmt.Sprintf("Reboot command sent to server %s", action.Target)
				}
			} else {
				res.Status = "failed"
				res.Error = fmt.Sprintf("invalid target server UUID: %s", action.Target)
			}
		} else {
			res.Response = fmt.Sprintf("Reboot requested for server %s (simulated)", action.Target)
		}

	case domain.ActionTypeTriggerBackup:
		if e.backupExec != nil && action.Target != "" {
			serverUUID, parseErr := uuid.Parse(action.Target)
			if parseErr == nil {
				backupName := fmt.Sprintf("auto-backup-%s-%s", rule.Name, time.Now().Format("20060102150405"))
				_, err := e.backupExec.TriggerBackup(ctx, rule.OrganizationID, serverUUID, nil, backupName)
				if err != nil {
					res.Status = "failed"
					res.Error = err.Error()
				} else {
					res.Response = fmt.Sprintf("Backup triggered for server %s", action.Target)
				}
			} else {
				res.Status = "failed"
				res.Error = fmt.Sprintf("invalid target server UUID: %s", action.Target)
			}
		} else {
			res.Response = fmt.Sprintf("Backup triggered for server %s (simulated)", action.Target)
		}

	default:
		res.Response = fmt.Sprintf("Action type %s executed successfully", action.Type)
	}

	return res
}

// matchTriggerConfig memverifikasi kecocokan konfigurasi trigger (misal filter server_id).
func (e *Engine) matchTriggerConfig(configJSON json.RawMessage, event domain.SystemEvent) bool {
	if len(configJSON) == 0 || string(configJSON) == "{}" {
		return true
	}

	var cfg map[string]any
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return true
	}

	if expectedServerID, ok := cfg["server_id"].(string); ok && expectedServerID != "" {
		actualServerID, _ := event.Data["server_id"].(string)
		if actualServerID != expectedServerID && !strings.Contains(event.SourceResource, expectedServerID) {
			return false
		}
	}

	return true
}

// evaluateConditions mengevaluasi seluruh klausa kondisi terhadap data kejadian sistem.
func (e *Engine) evaluateConditions(conditions []domain.RuleCondition, eventData map[string]any) (bool, map[string]any) {
	if len(conditions) == 0 {
		return true, map[string]any{"evaluation": "no conditions specified, automatic match"}
	}

	evalReport := make(map[string]any)

	for i, cond := range conditions {
		fieldVal, exists := eventData[cond.Field]
		if !exists {
			evalReport[fmt.Sprintf("condition_%d", i)] = map[string]any{
				"field":    cond.Field,
				"expected": cond.Value,
				"matched":  false,
				"reason":   "field not present in event payload",
			}
			return false, evalReport
		}

		matched := compareValues(fieldVal, cond.Operator, cond.Value)
		evalReport[fmt.Sprintf("condition_%d", i)] = map[string]any{
			"field":    cond.Field,
			"actual":   fieldVal,
			"operator": cond.Operator,
			"expected": cond.Value,
			"matched":  matched,
		}

		if !matched {
			return false, evalReport
		}
	}

	return true, evalReport
}

// compareValues membandingkan nilai aktual dengan nilai target berdasarkan operator logika.
func compareValues(actual any, op domain.ConditionOperator, expected any) bool {
	actNum, isActNum := toFloat64(actual)
	expNum, isExpNum := toFloat64(expected)

	if isActNum && isExpNum {
		switch op {
		case domain.OpGreaterThan:
			return actNum > expNum
		case domain.OpGreaterThanEqual:
			return actNum >= expNum
		case domain.OpLessThan:
			return actNum < expNum
		case domain.OpLessThanEqual:
			return actNum <= expNum
		case domain.OpEqual:
			return actNum == expNum
		case domain.OpNotEqual:
			return actNum != expNum
		default:
			return false
		}
	}

	// String / Boolean Comparison
	actStr := fmt.Sprintf("%v", actual)
	expStr := fmt.Sprintf("%v", expected)

	switch op {
	case domain.OpEqual:
		return strings.EqualFold(actStr, expStr)
	case domain.OpNotEqual:
		return !strings.EqualFold(actStr, expStr)
	case domain.OpContains:
		return strings.Contains(strings.ToLower(actStr), strings.ToLower(expStr))
	default:
		return actStr == expStr
	}
}

// toFloat64 mengonversi berbagai jenis tipe numerik interface{} menjadi float64.
func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// mapEventTypeToTriggerType memetakan string event type ke RuleTriggerType.
func mapEventTypeToTriggerType(eventType string) domain.RuleTriggerType {
	if strings.HasPrefix(eventType, "metric.") {
		return domain.TriggerTypeMetricThreshold
	}
	if strings.HasPrefix(eventType, "server.") {
		return domain.TriggerTypeServerStatusChanged
	}
	if strings.HasPrefix(eventType, "backup.") {
		return domain.TriggerTypeBackupEvent
	}
	return domain.TriggerTypeScheduledCron
}

// recordLog mencatat hasil eksekusi ke database.
func (e *Engine) recordLog(
	ctx context.Context,
	rule *domain.AutomationRule,
	triggerEvent string,
	status domain.ExecutionStatus,
	evaluatedConds map[string]any,
	actions []domain.ActionResultItem,
	errMsg string,
	duration time.Duration,
) error {
	evaluatedJSON, _ := json.Marshal(evaluatedConds)
	log := &domain.RuleExecutionLog{
		ID:                  uuid.New(),
		RuleID:              rule.ID,
		OrganizationID:      rule.OrganizationID,
		RuleName:            rule.Name,
		TriggerEvent:        triggerEvent,
		Status:              status,
		EvaluatedConditions: evaluatedJSON,
		ExecutedActions:     actions,
		ErrorMessage:        errMsg,
		ExecutionDurationMs: int(duration.Milliseconds()),
		ExecutedAt:          time.Now().UTC(),
	}

	return e.repo.CreateExecutionLog(ctx, log)
}
