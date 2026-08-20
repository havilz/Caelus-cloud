'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { BackupPolicy, BackupRecord } from '@/types/backup';
import { Server } from '@/types/server';
import { backupService } from '@/services/backup.service';
import { serverService } from '@/services/server.service';
import { CreateBackupPolicyModal } from '@/components/backup/CreateBackupPolicyModal';
import {
  Clock,
  Archive,
  Play,
  Trash2,
  CheckCircle2,
  AlertTriangle,
  ArrowLeft,
  Calendar,
  Layers,
  Database,
  Loader2,
  HardDrive,
  RefreshCw,
  Plus,
} from 'lucide-react';

export default function StorageBackupsPage() {
  const [activeTab, setActiveTab] = useState<'policies' | 'records'>('policies');
  const [policies, setPolicies] = useState<BackupPolicy[]>([]);
  const [records, setRecords] = useState<BackupRecord[]>([]);
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);

  // Modals & Action states
  const [isCreatePolicyOpen, setIsCreatePolicyOpen] = useState(false);
  const [triggeringServerId, setTriggeringServerId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [policiesRes, recordsRes, serversRes] = await Promise.all([
        backupService.listPolicies().catch(() => []),
        backupService.listRecords(1, 100).catch(() => ({ data: [] })),
        serverService.listServers(1, 100).catch(() => ({ data: [] })),
      ]);
      setPolicies(policiesRes);
      setRecords(recordsRes.data || []);
      setServers(serversRes.data || []);
    } catch (err) {
      console.error('Failed to fetch backup data:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleTriggerBackup = async (serverId: string, policyId?: string) => {
    try {
      setTriggeringServerId(serverId);
      await backupService.triggerBackup(serverId, { policy_id: policyId });
      await fetchData();
      setActiveTab('records');
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to trigger backup';
      alert(msg);
    } finally {
      setTriggeringServerId(null);
    }
  };

  const handleDeletePolicy = async (id: string) => {
    if (!confirm('Are you sure you want to delete this backup schedule policy?')) return;
    try {
      setDeletingId(id);
      await backupService.deletePolicy(id);
      await fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete policy';
      alert(msg);
    } finally {
      setDeletingId(null);
    }
  };

  const handleDeleteRecord = async (id: string) => {
    if (
      !confirm(
        'Are you sure you want to delete this backup archive and its object storage snapshot?'
      )
    ) {
      return;
    }
    try {
      setDeletingId(id);
      await backupService.deleteRecord(id);
      await fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete record';
      alert(msg);
    } finally {
      setDeletingId(null);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const totalBytesStored = records.reduce((acc, r) => acc + (r.size_bytes || 0), 0);
  const completedBackups = records.filter((r) => r.status === 'completed').length;

  return (
    <div className={AppContainers.pageWrapper}>
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center space-x-3">
          <Link
            href="/storage"
            className="p-2 rounded-lg bg-[#1a1a1a] border border-[#2e2e2e] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#3e3e3e] transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
          </Link>
          <div>
            <h1 className={AppText.h2}>Automated Backup Pipeline</h1>
            <p className={AppText.subtitle}>
              Automated server snapshots, disaster recovery lifecycle policies, and object storage archives.
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={fetchData}
            className="p-2 rounded-lg bg-[#1a1a1a] border border-[#2e2e2e] text-[#a1a1a1] hover:text-[#ededed] transition-colors"
            title="Refresh"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>

          <button
            onClick={() => setIsCreatePolicyOpen(true)}
            className={`px-3.5 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity shadow-sm`}
          >
            <Plus className="w-4 h-4" />
            <span>Create Policy</span>
          </button>
        </div>
      </div>

      {/* Metric Cards Grid */}
      <div className={AppContainers.metricsGrid}>
        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Active Policies</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Clock className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>{policies.length}</p>
          <p className="text-[11px] text-[#707070] mt-1">Scheduled cron routines</p>
        </div>

        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Backup Archives</span>
            <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
              <Archive className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>{completedBackups}</p>
          <p className="text-[11px] text-[#707070] mt-1">Stored snapshots</p>
        </div>

        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Total Storage Used</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400">
              <HardDrive className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>{formatFileSize(totalBytesStored)}</p>
          <p className="text-[11px] text-[#707070] mt-1">Compressed volume blocks</p>
        </div>

        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Auto Retention</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <CheckCircle2 className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>Active</p>
          <p className="text-[11px] text-[#707070] mt-1">Background cleaner worker</p>
        </div>
      </div>

      {/* Tabs & Content */}
      <div className={AppContainers.card}>
        {/* Tab Selector */}
        <div className="p-2 border-b border-[#262626] flex items-center space-x-2">
          <button
            onClick={() => setActiveTab('policies')}
            className={`px-4 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 transition-colors ${
              activeTab === 'policies'
                ? 'bg-[#262626] text-[#ededed]'
                : 'text-[#707070] hover:text-[#ededed]'
            }`}
          >
            <Clock className="w-3.5 h-3.5" />
            <span>Scheduled Policies ({policies.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('records')}
            className={`px-4 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 transition-colors ${
              activeTab === 'records'
                ? 'bg-[#262626] text-[#ededed]'
                : 'text-[#707070] hover:text-[#ededed]'
            }`}
          >
            <Layers className="w-3.5 h-3.5" />
            <span>Backup Records ({records.length})</span>
          </button>
        </div>

        {/* Tab 1: Policies */}
        {activeTab === 'policies' && (
          <div>
            {loading ? (
              <div className="p-12 flex flex-col items-center justify-center space-y-3">
                <Loader2 className="w-6 h-6 animate-spin text-emerald-400" />
                <p className={AppText.subtitle}>Loading backup policies...</p>
              </div>
            ) : policies.length === 0 ? (
              <div className="p-12 text-center space-y-3">
                <div className="p-3.5 rounded-full bg-[#1f1f1f] text-[#707070] w-fit mx-auto">
                  <Clock className="w-6 h-6" />
                </div>
                <div>
                  <p className={AppText.body}>No backup policies configured</p>
                  <p className={AppText.subtitle}>
                    Create automated backup routines for your cloud VPS instances.
                  </p>
                </div>
                <button
                  onClick={() => setIsCreatePolicyOpen(true)}
                  className={`px-4 py-2 rounded-lg text-xs font-semibold inline-flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity`}
                >
                  <Plus className="w-4 h-4" />
                  <span>Configure First Policy</span>
                </button>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-[#262626] bg-[#141414]/50 text-[11px] font-medium text-[#707070] uppercase tracking-wider">
                      <th className="py-3 px-4">Policy Name</th>
                      <th className="py-3 px-4">Server</th>
                      <th className="py-3 px-4">Schedule</th>
                      <th className="py-3 px-4">Retention</th>
                      <th className="py-3 px-4">Last Run</th>
                      <th className="py-3 px-4 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#262626] text-xs">
                    {policies.map((p) => {
                      const isTriggering = triggeringServerId === p.server_id;
                      const isDeleting = deletingId === p.id;
                      return (
                        <tr key={p.id} className="hover:bg-[#1a1a1a]/50 transition-colors">
                          <td className="py-3.5 px-4 font-medium text-[#ededed]">
                            <div className="flex items-center space-x-2">
                              <span className="w-2 h-2 rounded-full bg-emerald-400" />
                              <span>{p.name}</span>
                            </div>
                          </td>

                          <td className="py-3.5 px-4">
                            <span className="text-[#a1a1a1]">
                              {p.server_name || p.server_id.slice(0, 8)}
                            </span>
                          </td>

                          <td className="py-3.5 px-4 font-mono text-[11px] text-emerald-400">
                            {p.cron_expression}
                          </td>

                          <td className="py-3.5 px-4 text-[#a1a1a1] text-[11px]">
                            {p.retention_days} Days
                          </td>

                          <td className="py-3.5 px-4 text-[#707070] text-[11px]">
                            {p.last_run_at
                              ? new Date(p.last_run_at).toLocaleString(undefined, {
                                  month: 'short',
                                  day: 'numeric',
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })
                              : 'Never'}
                          </td>

                          <td className="py-3.5 px-4 text-right">
                            <div className="flex items-center justify-end space-x-2">
                              <button
                                onClick={() => handleTriggerBackup(p.server_id, p.id)}
                                disabled={isTriggering}
                                className="px-2.5 py-1 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 hover:bg-emerald-500/20 text-[11px] font-semibold flex items-center space-x-1.5 transition-colors disabled:opacity-50"
                              >
                                {isTriggering ? (
                                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                ) : (
                                  <Play className="w-3.5 h-3.5 fill-emerald-400" />
                                )}
                                <span>Run Now</span>
                              </button>

                              <button
                                onClick={() => handleDeletePolicy(p.id)}
                                disabled={isDeleting}
                                className="p-1 rounded-md text-[#707070] hover:text-rose-400 hover:bg-rose-950/30 transition-colors disabled:opacity-50"
                                title="Delete Policy"
                              >
                                {isDeleting ? (
                                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                ) : (
                                  <Trash2 className="w-3.5 h-3.5" />
                                )}
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* Tab 2: Records */}
        {activeTab === 'records' && (
          <div>
            {loading ? (
              <div className="p-12 flex flex-col items-center justify-center space-y-3">
                <Loader2 className="w-6 h-6 animate-spin text-emerald-400" />
                <p className={AppText.subtitle}>Loading backup archives...</p>
              </div>
            ) : records.length === 0 ? (
              <div className="p-12 text-center space-y-3">
                <div className="p-3.5 rounded-full bg-[#1f1f1f] text-[#707070] w-fit mx-auto">
                  <Archive className="w-6 h-6" />
                </div>
                <div>
                  <p className={AppText.body}>No backup records found</p>
                  <p className={AppText.subtitle}>
                    Execute a scheduled policy or trigger a manual snapshot above.
                  </p>
                </div>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-[#262626] bg-[#141414]/50 text-[11px] font-medium text-[#707070] uppercase tracking-wider">
                      <th className="py-3 px-4">Backup Name</th>
                      <th className="py-3 px-4">Server</th>
                      <th className="py-3 px-4">Storage Key</th>
                      <th className="py-3 px-4">Size</th>
                      <th className="py-3 px-4">Status</th>
                      <th className="py-3 px-4">Created At</th>
                      <th className="py-3 px-4">Expires At</th>
                      <th className="py-3 px-4 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#262626] text-xs">
                    {records.map((r) => {
                      const isDeleting = deletingId === r.id;
                      return (
                        <tr key={r.id} className="hover:bg-[#1a1a1a]/50 transition-colors">
                          <td className="py-3.5 px-4 font-medium text-[#ededed]">
                            <div className="flex items-center space-x-2">
                              <Archive className="w-3.5 h-3.5 text-emerald-400" />
                              <span>{r.backup_name}</span>
                            </div>
                          </td>

                          <td className="py-3.5 px-4 text-[#a1a1a1]">
                            {r.server_name || r.server_id.slice(0, 8)}
                          </td>

                          <td className="py-3.5 px-4 font-mono text-[11px] text-[#707070] truncate max-w-xs">
                            {r.storage_key}
                          </td>

                          <td className="py-3.5 px-4 font-mono text-[#ededed] text-[11px]">
                            {formatFileSize(r.size_bytes)}
                          </td>

                          <td className="py-3.5 px-4">
                            <span
                              className={`px-2 py-0.5 rounded-md text-[11px] font-medium border ${
                                r.status === 'completed'
                                  ? 'bg-emerald-950/60 text-emerald-300 border-emerald-800/40'
                                  : r.status === 'in_progress'
                                  ? 'bg-cyan-950/60 text-cyan-300 border-cyan-800/40'
                                  : 'bg-rose-950/60 text-rose-300 border-rose-800/40'
                              }`}
                            >
                              {r.status}
                            </span>
                          </td>

                          <td className="py-3.5 px-4 text-[#707070] text-[11px]">
                            {new Date(r.created_at).toLocaleString(undefined, {
                              month: 'short',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </td>

                          <td className="py-3.5 px-4 text-[#707070] text-[11px]">
                            {r.expires_at
                              ? new Date(r.expires_at).toLocaleDateString(undefined, {
                                  year: 'numeric',
                                  month: 'short',
                                  day: 'numeric',
                                })
                              : 'Never'}
                          </td>

                          <td className="py-3.5 px-4 text-right">
                            <button
                              onClick={() => handleDeleteRecord(r.id)}
                              disabled={isDeleting}
                              className="p-1 rounded-md text-[#707070] hover:text-rose-400 hover:bg-rose-950/30 transition-colors disabled:opacity-50"
                              title="Delete Record"
                            >
                              {isDeleting ? (
                                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                              ) : (
                                <Trash2 className="w-3.5 h-3.5" />
                              )}
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Create Backup Policy Modal */}
      <CreateBackupPolicyModal
        isOpen={isCreatePolicyOpen}
        onClose={() => setIsCreatePolicyOpen(false)}
        onSuccess={fetchData}
      />
    </div>
  );
}
