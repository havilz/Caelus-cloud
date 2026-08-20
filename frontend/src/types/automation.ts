export type RuleTriggerType =
  | 'metric_threshold'
  | 'server_status_changed'
  | 'backup_event'
  | 'scheduled_cron';

export type ConditionOperator = '>' | '>=' | '<' | '<=' | '==' | '!=' | 'in' | 'contains';

export interface RuleCondition {
  field: string;
  operator: ConditionOperator;
  value: string | number | boolean;
  duration_minutes?: number;
}

export type ActionType =
  | 'send_email'
  | 'send_webhook'
  | 'reboot_server'
  | 'shutdown_server'
  | 'trigger_backup'
  | 'scale_server';

export interface RuleAction {
  type: ActionType;
  target?: string;
  config?: Record<string, unknown>;
}

export interface AutomationRule {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  is_active: boolean;
  trigger_type: RuleTriggerType;
  trigger_config?: Record<string, unknown>;
  conditions: RuleCondition[];
  actions: RuleAction[];
  cooldown_seconds: number;
  last_triggered_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateRulePayload {
  name: string;
  description?: string;
  is_active?: boolean;
  trigger_type: RuleTriggerType;
  trigger_config?: Record<string, unknown>;
  conditions: RuleCondition[];
  actions: RuleAction[];
  cooldown_seconds?: number;
}

export interface UpdateRulePayload {
  name?: string;
  description?: string;
  is_active?: boolean;
  trigger_type?: RuleTriggerType;
  trigger_config?: Record<string, unknown>;
  conditions?: RuleCondition[];
  actions?: RuleAction[];
  cooldown_seconds?: number;
}

export type ExecutionStatus = 'success' | 'failed' | 'partially_failed' | 'skipped';

export interface ActionResultItem {
  action_type: ActionType;
  target?: string;
  status: 'success' | 'failed';
  response?: string;
  error?: string;
}

export interface RuleExecutionLog {
  id: string;
  rule_id: string;
  organization_id: string;
  rule_name?: string;
  trigger_event: string;
  status: ExecutionStatus;
  evaluated_conditions: Record<string, unknown>;
  executed_actions: ActionResultItem[];
  error_message?: string;
  execution_duration_ms: number;
  executed_at: string;
}
