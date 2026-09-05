'use client';

import React, { useState } from 'react';
import { X, Plus, Trash2, Zap, AlertTriangle, ShieldCheck } from 'lucide-react';
import { AppTheme } from '@/core/theme';
import {
  RuleTriggerType,
  ConditionOperator,
  ActionType,
  RuleCondition,
  RuleAction,
  CreateRulePayload,
} from '@/types/automation';
import { automationService } from '@/services/automation.service';

interface CreateRuleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const CreateRuleModal: React.FC<CreateRuleModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [triggerType, setTriggerType] = useState<RuleTriggerType>('metric_threshold');
  const [cooldownSeconds, setCooldownSeconds] = useState(300);

  const [conditions, setConditions] = useState<RuleCondition[]>([
    { field: 'cpu_usage_percent', operator: '>=', value: 85, duration_minutes: 5 },
  ]);

  const [actions, setActions] = useState<RuleAction[]>([
    { type: 'send_email', target: '' },
  ]);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleAddCondition = () => {
    setConditions([
      ...conditions,
      { field: 'memory_usage_percent', operator: '>=', value: 90, duration_minutes: 5 },
    ]);
  };

  const handleRemoveCondition = (index: number) => {
    setConditions(conditions.filter((_, i) => i !== index));
  };

  const handleConditionChange = (
    index: number,
    key: keyof RuleCondition,
    val: string | number
  ) => {
    const updated = [...conditions];
    if (key === 'value' || key === 'duration_minutes') {
      const num = Number(val);
      updated[index] = { ...updated[index], [key]: isNaN(num) ? val : num };
    } else {
      updated[index] = { ...updated[index], [key]: val };
    }
    setConditions(updated);
  };

  const handleAddAction = () => {
    setActions([...actions, { type: 'send_webhook', target: '' }]);
  };

  const handleRemoveAction = (index: number) => {
    setActions(actions.filter((_, i) => i !== index));
  };

  const handleActionChange = (index: number, key: keyof RuleAction, val: string) => {
    const updated = [...actions];
    updated[index] = { ...updated[index], [key]: val };
    setActions(updated);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Rule name is required');
      return;
    }
    if (actions.length === 0) {
      setError('At least one action is required');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const payload: CreateRulePayload = {
        name,
        description: description || undefined,
        is_active: true,
        trigger_type: triggerType,
        conditions,
        actions,
        cooldown_seconds: Number(cooldownSeconds) || 300,
      };

      await automationService.createRule(payload);
      onSuccess();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create rule';
      setError(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={AppTheme.containers.modalBackdrop}>
      <div className={`${AppTheme.containers.modalDialog} max-w-2xl`}>
        <div className={AppTheme.containers.cardHeader}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className={AppTheme.controls.iconBoxEmerald}>
                <Zap className="w-5 h-5" />
              </div>
              <div>
                <h3 className={AppTheme.text.h3}>Create Automation Rule</h3>
                <p className={AppTheme.text.subtitle}>
                  Configure automated event triggers, condition clauses, and chained actions.
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className={AppTheme.controls.iconButton}
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit} className={AppTheme.containers.modalBodyScroll}>
          <div className="p-6 space-y-6">
            {error && (
              <div className="p-3 rounded-lg bg-rose-950/60 border border-rose-800/40 flex items-center gap-2 text-rose-400 text-xs">
                <AlertTriangle className="w-4 h-4 shrink-0" />
                <span>{error}</span>
              </div>
            )}

            <div className="space-y-4">
              <div>
                <label className={AppTheme.text.label}>Rule Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Auto-Heal High Memory Server"
                  className={AppTheme.controls.input}
                  required
                />
              </div>

              <div>
                <label className={AppTheme.text.label}>Description (Optional)</label>
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="e.g. Automatically trigger emergency snapshot and notify devops"
                  className={AppTheme.controls.input}
                />
              </div>
            </div>

            <div>
              <label className={AppTheme.text.label}>Event Trigger Type</label>
              <select
                value={triggerType}
                onChange={(e) => setTriggerType(e.target.value as RuleTriggerType)}
                className={AppTheme.controls.select}
              >
                <option value="metric_threshold">Telemetry Metric Threshold (CPU, Memory, Disk, Load)</option>
                <option value="server_status_changed">Server Status Transition (Offline, Stopped, Error)</option>
                <option value="backup_event">Backup Lifecycle (Backup Failed, Expired)</option>
                <option value="scheduled_cron">Scheduled Cron Routine</option>
              </select>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className={AppTheme.text.h4}>Conditions (WHEN)</h4>
                  <p className={AppTheme.text.caption}>Evaluate clauses before triggering automated actions.</p>
                </div>
                <button
                  type="button"
                  onClick={handleAddCondition}
                  className={AppTheme.controls.buttonAction}
                >
                  <Plus className="w-3.5 h-3.5" />
                  Add Clause
                </button>
              </div>

              <div className="space-y-2.5">
                {conditions.map((cond, idx) => (
                  <div
                    key={idx}
                    className={`${AppTheme.controls.cardSubtle} flex flex-wrap sm:flex-nowrap items-center gap-2`}
                  >
                    <input
                      type="text"
                      value={cond.field}
                      onChange={(e) => handleConditionChange(idx, 'field', e.target.value)}
                      placeholder="field (e.g. cpu_usage_percent)"
                      className={AppTheme.controls.inputMono}
                      required
                    />

                    <select
                      value={cond.operator}
                      onChange={(e) =>
                        handleConditionChange(idx, 'operator', e.target.value as ConditionOperator)
                      }
                      className={AppTheme.controls.selectSm}
                    >
                      <option value=">=">&gt;=</option>
                      <option value=">">&gt;</option>
                      <option value="<=">&lt;=</option>
                      <option value="<">&lt;</option>
                      <option value="==">==</option>
                      <option value="!=">!=</option>
                      <option value="contains">contains</option>
                    </select>

                    <input
                      type="text"
                      value={String(cond.value)}
                      onChange={(e) => handleConditionChange(idx, 'value', e.target.value)}
                      placeholder="threshold (e.g. 85)"
                      className={`w-28 ${AppTheme.controls.inputMono}`}
                      required
                    />

                    <button
                      type="button"
                      onClick={() => handleRemoveCondition(idx)}
                      disabled={conditions.length <= 1}
                      className={AppTheme.controls.iconButtonDanger}
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className={AppTheme.text.h4}>Actions (THEN)</h4>
                  <p className={AppTheme.text.caption}>Dispatch actions when conditions are satisfied.</p>
                </div>
                <button
                  type="button"
                  onClick={handleAddAction}
                  className={AppTheme.controls.buttonAction}
                >
                  <Plus className="w-3.5 h-3.5" />
                  Add Action
                </button>
              </div>

              <div className="space-y-2.5">
                {actions.map((act, idx) => (
                  <div
                    key={idx}
                    className={`${AppTheme.controls.cardSubtle} flex flex-wrap sm:flex-nowrap items-center gap-2`}
                  >
                    <select
                      value={act.type}
                      onChange={(e) => handleActionChange(idx, 'type', e.target.value as ActionType)}
                      className={`w-44 ${AppTheme.controls.selectSm}`}
                    >
                      <option value="send_email">Send Email</option>
                      <option value="send_webhook">Send HTTP Webhook</option>
                      <option value="reboot_server">Reboot Server</option>
                      <option value="trigger_backup">Trigger Backup Snapshot</option>
                    </select>

                    <input
                      type="text"
                      value={act.target || ''}
                      onChange={(e) => handleActionChange(idx, 'target', e.target.value)}
                      placeholder={
                        act.type === 'send_email'
                          ? 'email@example.com'
                          : act.type === 'send_webhook'
                          ? 'https://api.domain.com/webhook'
                          : 'Target Server UUID'
                      }
                      className={AppTheme.controls.inputMono}
                      required
                    />

                    <button
                      type="button"
                      onClick={() => handleRemoveAction(idx)}
                      disabled={actions.length <= 1}
                      className={AppTheme.controls.iconButtonDanger}
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="w-4 h-4 text-emerald-400" />
                <label className={AppTheme.text.label}>Anti-Flapping Cooldown Window</label>
              </div>
              <p className={AppTheme.text.caption}>Minimum seconds to wait before re-triggering this rule.</p>
              <input
                type="number"
                min={0}
                value={cooldownSeconds}
                onChange={(e) => setCooldownSeconds(Number(e.target.value))}
                className={`mt-1.5 sm:w-48 ${AppTheme.controls.input}`}
              />
            </div>
          </div>

          <div className="p-5 border-t border-[#262626] bg-[#141414] flex items-center justify-end gap-3 rounded-b-xl">
            <button
              type="button"
              onClick={onClose}
              className={AppTheme.controls.buttonGhost}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className={AppTheme.controls.buttonPrimary}
            >
              <Zap className="w-4 h-4" />
              {isSubmitting ? 'Saving Rule...' : 'Create Rule'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
