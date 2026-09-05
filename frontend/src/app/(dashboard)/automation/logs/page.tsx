'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  RefreshCw,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Clock,
  Zap,
  ChevronRight,
  Code,
} from 'lucide-react';
import { AppTheme } from '@/core/theme';
import { RuleExecutionLog, ExecutionStatus } from '@/types/automation';
import { automationService } from '@/services/automation.service';

export default function AutomationLogsPage() {
  const [logs, setLogs] = useState<RuleExecutionLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<ExecutionStatus | ''>('');
  const [isLoading, setIsLoading] = useState(true);
  const [expandedLogId, setExpandedLogId] = useState<string | null>(null);

  const fetchLogs = useCallback(async () => {
    try {
      setIsLoading(true);
      const res = await automationService.listLogs(
        page,
        20,
        undefined,
        statusFilter || undefined
      );
      setLogs(res.data || []);
      setTotal(res.meta?.total_items || 0);
    } catch (err) {
      console.error('Failed to fetch execution logs:', err);
    } finally {
      setIsLoading(false);
    }
  }, [page, statusFilter]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const getStatusBadge = (status: ExecutionStatus) => {
    switch (status) {
      case 'success':
        return (
          <span className="px-2.5 py-0.5 rounded-md bg-emerald-950/60 text-emerald-400 border border-emerald-800/40 text-[11px] font-semibold flex items-center gap-1">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Success
          </span>
        );
      case 'failed':
        return (
          <span className="px-2.5 py-0.5 rounded-md bg-rose-950/60 text-rose-400 border border-rose-800/40 text-[11px] font-semibold flex items-center gap-1">
            <XCircle className="w-3.5 h-3.5" />
            Failed
          </span>
        );
      case 'partially_failed':
        return (
          <span className="px-2.5 py-0.5 rounded-md bg-amber-950/60 text-amber-400 border border-amber-800/40 text-[11px] font-semibold flex items-center gap-1">
            <AlertTriangle className="w-3.5 h-3.5" />
            Partial Failure
          </span>
        );
      case 'skipped':
        return (
          <span className="px-2.5 py-0.5 rounded-md bg-zinc-900 text-zinc-400 border border-zinc-800 text-[11px] font-semibold flex items-center gap-1">
            <Clock className="w-3.5 h-3.5" />
            Skipped (Cooldown)
          </span>
        );
      default:
        return null;
    }
  };

  return (
    <div className={AppTheme.containers.pageWrapper}>
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <Link
            href="/automation"
            className="inline-flex items-center gap-1.5 text-xs text-[#a1a1a1] hover:text-[#ededed] mb-1 transition-colors"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            Back to Automation Rules
          </Link>
          <h1 className={AppTheme.text.h1}>Automation Audit Logs</h1>
          <p className={AppTheme.text.subtitle}>
            Historical trace records of automated rule evaluations, condition matches, and dispatched actions.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value as ExecutionStatus | '');
              setPage(1);
            }}
            className={AppTheme.controls.select}
          >
            <option value="">All Statuses</option>
            <option value="success">Success</option>
            <option value="failed">Failed</option>
            <option value="partially_failed">Partial Failure</option>
            <option value="skipped">Skipped (Cooldown)</option>
          </select>

          <button
            onClick={fetchLogs}
            className={AppTheme.controls.iconButton}
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      <div className={AppTheme.containers.card}>
        <div className={`${AppTheme.containers.cardHeader} flex items-center justify-between`}>
          <div>
            <h3 className={AppTheme.text.h3}>Execution History</h3>
            <p className={AppTheme.text.caption}>Total {total} execution audit records logged.</p>
          </div>
        </div>

        <div className="p-5">
          {isLoading ? (
            <div className="py-12 flex flex-col items-center justify-center text-center">
              <RefreshCw className="w-6 h-6 animate-spin text-emerald-400 mb-2" />
              <p className={AppTheme.text.bodySm}>Loading audit logs...</p>
            </div>
          ) : logs.length === 0 ? (
            <div className="py-12 flex flex-col items-center justify-center text-center text-[#707070]">
              <Clock className="w-8 h-8 mb-2 opacity-50" />
              <p className={AppTheme.text.bodySm}>No automation execution logs found for the selected filter.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {logs.map((log) => {
                const isExpanded = expandedLogId === log.id;

                return (
                  <div
                    key={log.id}
                    className={AppTheme.controls.cardRow}
                  >
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <div className={AppTheme.controls.iconBoxEmerald}>
                          <Zap className="w-4 h-4" />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-sm text-[#ededed]">
                              {log.rule_name || 'Deleted Rule'}
                            </span>
                            {getStatusBadge(log.status)}
                          </div>
                          <div className="flex items-center gap-2 mt-0.5">
                            <span className={AppTheme.text.codeMuted}>{log.trigger_event}</span>
                            <span className="text-[#404040]">&bull;</span>
                            <span className={AppTheme.text.caption}>
                              {log.execution_duration_ms}ms duration
                            </span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        <span className={AppTheme.text.caption}>
                          {new Date(log.executed_at).toLocaleString()}
                        </span>
                        <button
                          onClick={() => setExpandedLogId(isExpanded ? null : log.id)}
                          className={AppTheme.controls.iconButton}
                        >
                          <ChevronRight className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-90' : ''}`} />
                        </button>
                      </div>
                    </div>

                    {log.executed_actions && log.executed_actions.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 pt-1">
                        {log.executed_actions.map((act, i) => (
                          <span
                            key={i}
                            className={`px-2 py-0.5 rounded-md text-[11px] font-mono flex items-center gap-1.5 ${
                              act.status === 'success'
                                ? 'bg-emerald-950/40 text-emerald-300 border border-emerald-800/30'
                                : 'bg-rose-950/40 text-rose-300 border border-rose-800/30'
                            }`}
                          >
                            <span>{act.action_type}</span>
                            {act.target && <span className="opacity-70">({act.target})</span>}
                            {act.status === 'success' ? (
                              <CheckCircle2 className="w-3 h-3 text-emerald-400" />
                            ) : (
                              <XCircle className="w-3 h-3 text-rose-400" />
                            )}
                          </span>
                        ))}
                      </div>
                    )}

                    {log.error_message && (
                      <div className="p-2.5 rounded-lg bg-rose-950/40 border border-rose-800/30 text-rose-400 text-xs flex items-center gap-2">
                        <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                        <span>{log.error_message}</span>
                      </div>
                    )}

                    {isExpanded && (
                      <div className="pt-2 border-t border-[#202020] space-y-2">
                        <div className="flex items-center gap-1.5 text-xs text-[#a1a1a1]">
                          <Code className="w-3.5 h-3.5 text-emerald-400" />
                          <span>Condition Evaluation Payload:</span>
                        </div>
                        <pre className={AppTheme.controls.codeBox}>
                          {JSON.stringify(log.evaluated_conditions, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}

          {total > 20 && (
            <div className="flex items-center justify-between pt-5 mt-4 border-t border-[#222222]">
              <div className="text-xs text-[#707070]">
                Showing {logs.length} of {total} records
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1.5 rounded-lg bg-[#141414] border border-[#262626] text-xs text-[#ededed] disabled:opacity-30 hover:bg-[#222222] transition-colors"
                >
                  Previous
                </button>
                <span className="text-xs font-mono text-[#a1a1a1]">Page {page}</span>
                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={page * 20 >= total}
                  className="px-3 py-1.5 rounded-lg bg-[#141414] border border-[#262626] text-xs text-[#ededed] disabled:opacity-30 hover:bg-[#222222] transition-colors"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
