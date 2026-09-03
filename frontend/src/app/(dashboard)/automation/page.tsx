'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import {
  Zap,
  Plus,
  Play,
  Trash2,
  Activity,
  Server,
  Database,
  Clock,
  CheckCircle2,
  XCircle,
  FileText,
  RefreshCw,
} from 'lucide-react';
import { AppTheme } from '@/core/theme';
import { AutomationRule } from '@/types/automation';
import { automationService } from '@/services/automation.service';
import { CreateRuleModal } from '@/components/automation/CreateRuleModal';

export default function AutomationRulesPage() {
  const [rules, setRules] = useState<AutomationRule[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [testingRuleId, setTestingRuleId] = useState<string | null>(null);
  const [testResult, setTestResult] = useState<{ id: string; success: boolean; message: string } | null>(null);

  const fetchRules = useCallback(async () => {
    try {
      setIsLoading(true);
      const res = await automationService.listRules(1, 100);
      setRules(res.data || []);
    } catch (err) {
      console.error('Failed to load automation rules:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRules();
  }, [fetchRules]);

  const handleToggleRule = async (rule: AutomationRule) => {
    try {
      const updated = await automationService.updateRule(rule.id, {
        is_active: !rule.is_active,
      });
      setRules((prev) => prev.map((r) => (r.id === rule.id ? updated : r)));
    } catch (err) {
      console.error('Failed to toggle rule:', err);
    }
  };

  const handleDeleteRule = async (id: string) => {
    if (!confirm('Are you sure you want to delete this automation rule?')) return;
    try {
      await automationService.deleteRule(id);
      setRules((prev) => prev.filter((r) => r.id !== id));
    } catch (err) {
      console.error('Failed to delete rule:', err);
    }
  };

  const handleTestRule = async (id: string) => {
    try {
      setTestingRuleId(id);
      setTestResult(null);
      const log = await automationService.testRule(id);
      setTestResult({
        id,
        success: log.status === 'success',
        message: `Simulation finished with status: ${log.status} (${log.execution_duration_ms}ms)`,
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Test run failed';
      setTestResult({ id, success: false, message: msg });
    } finally {
      setTestingRuleId(null);
    }
  };

  const activeCount = rules.filter((r) => r.is_active).length;
  const metricCount = rules.filter((r) => r.trigger_type === 'metric_threshold').length;
  const serverCount = rules.filter((r) => r.trigger_type === 'server_status_changed').length;
  const backupCount = rules.filter((r) => r.trigger_type === 'backup_event').length;

  return (
    <div className={AppTheme.containers.pageWrapper}>
      {}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className={AppTheme.text.categoryTag}>Automation Engine</span>
          </div>
          <h1 className={AppTheme.text.h1}>Event & Automation Rules</h1>
          <p className={AppTheme.text.subtitle}>
            Build event-driven triggers, conditional logic pipelines, and self-healing cloud actions.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/automation/logs"
            className={AppTheme.controls.buttonSecondary}
          >
            <FileText className="w-4 h-4 text-emerald-400" />
            Audit Logs
          </Link>

          <button
            onClick={() => setIsModalOpen(true)}
            className={AppTheme.controls.buttonPrimary}
          >
            <Plus className="w-4 h-4" />
            New Rule
          </button>
        </div>
      </div>

      {}
      <div className={AppTheme.containers.metricsGrid}>
        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardContent} flex items-center gap-3`}>
          <div className={AppTheme.controls.iconBoxEmerald}>
            <Zap className="w-5 h-5" />
          </div>
          <div>
            <div className={AppTheme.text.caption}>Active Rules</div>
            <div className="text-xl font-bold text-[#ededed]">{activeCount} / {rules.length}</div>
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardContent} flex items-center gap-3`}>
          <div className={AppTheme.controls.iconBoxCyan}>
            <Activity className="w-5 h-5" />
          </div>
          <div>
            <div className={AppTheme.text.caption}>Metric Triggers</div>
            <div className="text-xl font-bold text-[#ededed]">{metricCount}</div>
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardContent} flex items-center gap-3`}>
          <div className={AppTheme.controls.iconBoxPurple}>
            <Server className="w-5 h-5" />
          </div>
          <div>
            <div className={AppTheme.text.caption}>Server Health Rules</div>
            <div className="text-xl font-bold text-[#ededed]">{serverCount}</div>
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardContent} flex items-center gap-3`}>
          <div className={AppTheme.controls.iconBoxAmber}>
            <Database className="w-5 h-5" />
          </div>
          <div>
            <div className={AppTheme.text.caption}>Backup Lifecycle</div>
            <div className="text-xl font-bold text-[#ededed]">{backupCount}</div>
          </div>
        </div>
      </div>

      {}
      <div className={AppTheme.containers.card}>
        <div className={`${AppTheme.containers.cardHeader} flex items-center justify-between`}>
          <div>
            <h3 className={AppTheme.text.h3}>Configured Automation Rules</h3>
            <p className={AppTheme.text.caption}>Active workflows executed automatically by the background engine.</p>
          </div>
          <button
            onClick={fetchRules}
            className={AppTheme.controls.iconButton}
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>

        <div className="p-5">
          {isLoading ? (
            <div className="py-12 flex flex-col items-center justify-center text-center">
              <RefreshCw className="w-6 h-6 animate-spin text-emerald-400 mb-2" />
              <p className={AppTheme.text.bodySm}>Loading automation rules...</p>
            </div>
          ) : rules.length === 0 ? (
            <div className="py-16 flex flex-col items-center justify-center text-center">
              <div className="p-4 rounded-full bg-[#1c1c1c] border border-[#2e2e2e] text-emerald-400 mb-3">
                <Zap className="w-8 h-8" />
              </div>
              <h3 className={AppTheme.text.h3}>No Automation Rules Yet</h3>
              <p className={`${AppTheme.text.bodySm} max-w-sm mt-1 mb-4`}>
                Create your first automated trigger to auto-scale resources, reboot failing servers, or dispatch webhook alerts.
              </p>
              <button
                onClick={() => setIsModalOpen(true)}
                className={AppTheme.controls.buttonPrimary}
              >
                Create Automation Rule
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {rules.map((rule) => (
                <div
                  key={rule.id}
                  className={AppTheme.controls.cardRowActive}
                >
                  {}
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <button
                        onClick={() => handleToggleRule(rule)}
                        className={`w-9 h-5 rounded-full p-0.5 transition-colors ${
                          rule.is_active ? 'bg-emerald-500' : 'bg-[#2a2a2a]'
                        }`}
                      >
                        <div
                          className={`w-4 h-4 rounded-full bg-zinc-950 transition-transform ${
                            rule.is_active ? 'translate-x-4' : 'translate-x-0'
                          }`}
                        />
                      </button>

                      <div>
                        <div className="flex items-center gap-2">
                          <h4 className={AppTheme.text.h4}>{rule.name}</h4>
                          <span className={rule.is_active ? AppTheme.controls.badgeActive : AppTheme.controls.badgeInactive}>
                            {rule.is_active ? 'Active' : 'Disabled'}
                          </span>
                          <span className={AppTheme.controls.badgeMono}>
                            {rule.trigger_type}
                          </span>
                        </div>
                        {rule.description && (
                          <p className={AppTheme.text.bodySm}>{rule.description}</p>
                        )}
                      </div>
                    </div>

                    {}
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleTestRule(rule.id)}
                        disabled={testingRuleId === rule.id}
                        className={AppTheme.controls.buttonGhost}
                      >
                        <Play className={`w-3.5 h-3.5 ${testingRuleId === rule.id ? 'animate-spin' : ''}`} />
                        {testingRuleId === rule.id ? 'Running Test...' : 'Test Run'}
                      </button>

                      <button
                        onClick={() => handleDeleteRule(rule.id)}
                        className={AppTheme.controls.iconButtonDanger}
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>

                  {}
                  {testResult && testResult.id === rule.id && (
                    <div
                      className={`p-3 rounded-lg text-xs flex items-center gap-2 ${
                        testResult.success
                          ? 'bg-emerald-950/60 border border-emerald-800/40 text-emerald-400'
                          : 'bg-rose-950/60 border border-rose-800/40 text-rose-400'
                      }`}
                    >
                      {testResult.success ? (
                        <CheckCircle2 className="w-4 h-4 shrink-0" />
                      ) : (
                        <XCircle className="w-4 h-4 shrink-0" />
                      )}
                      <span>{testResult.message}</span>
                    </div>
                  )}

                  {}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-1 border-t border-[#1f1f1f]">
                    {}
                    <div className="space-y-1.5">
                      <span className={AppTheme.text.caption}>Trigger Conditions (AND):</span>
                      <div className="flex flex-wrap gap-1.5">
                        {rule.conditions.map((cond, i) => (
                          <span
                            key={i}
                            className={AppTheme.controls.pillMono}
                          >
                            {cond.field} <span className="text-emerald-400">{cond.operator}</span> {String(cond.value)}
                          </span>
                        ))}
                      </div>
                    </div>

                    {}
                    <div className="space-y-1.5">
                      <span className={AppTheme.text.caption}>Dispatched Actions:</span>
                      <div className="flex flex-wrap gap-1.5">
                        {rule.actions.map((act, i) => (
                          <span
                            key={i}
                            className={AppTheme.controls.pillSubtle}
                          >
                            <Zap className="w-3 h-3 text-emerald-400" />
                            <span className="font-medium text-emerald-400">{act.type}</span>
                            {act.target && (
                              <span className="text-[#a1a1a1] font-mono text-[11px]">({act.target})</span>
                            )}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>

                  {}
                  <div className="flex items-center justify-between text-[11px] text-[#707070] pt-1">
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3.5 h-3.5" />
                      <span>Cooldown: {rule.cooldown_seconds}s</span>
                    </div>
                    <div>
                      Last Triggered: {rule.last_triggered_at ? new Date(rule.last_triggered_at).toLocaleString() : 'Never'}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {}
      <CreateRuleModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={fetchRules}
      />
    </div>
  );
}
