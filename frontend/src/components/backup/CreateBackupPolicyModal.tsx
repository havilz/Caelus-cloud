'use client';

import React, { useState, useEffect } from 'react';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { CreateBackupPolicyInput } from '@/types/backup';
import { Server } from '@/types/server';
import { Bucket } from '@/types/storage';
import { backupService } from '@/services/backup.service';
import { serverService } from '@/services/server.service';
import { storageService } from '@/services/storage.service';
import { X, ShieldAlert, Clock, HardDrive, Database, Loader2 } from 'lucide-react';

interface CreateBackupPolicyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const CreateBackupPolicyModal: React.FC<CreateBackupPolicyModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [servers, setServers] = useState<Server[]>([]);
  const [buckets, setBuckets] = useState<Bucket[]>([]);
  const [loadingResources, setLoadingResources] = useState(false);

  const [formData, setFormData] = useState<CreateBackupPolicyInput>({
    server_id: '',
    bucket_id: undefined,
    name: '',
    cron_expression: '0 2 * * *',
    retention_days: 7,
    include_disks: true,
  });

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setLoadingResources(true);
      Promise.all([
        serverService.listServers(1, 100).catch(() => ({ data: [] })),
        storageService.listBuckets(1, 100).catch(() => ({ data: [] })),
      ])
        .then(([srvRes, bktRes]) => {
          const srvList = srvRes.data || [];
          setServers(srvList);
          if (srvList.length > 0) {
            setFormData((prev) => ({
              ...prev,
              server_id: srvList[0].id,
              name: `Daily Backup - ${srvList[0].name}`,
            }));
          }
          const bktList = bktRes.data || [];
          setBuckets(bktList);
          if (bktList.length > 0) {
            setFormData((prev) => ({ ...prev, bucket_id: bktList[0].id }));
          }
        })
        .finally(() => setLoadingResources(false));
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.server_id) {
      setError('Please select a target server.');
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      await backupService.createPolicy(formData);
      onSuccess();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create backup policy';
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className={AppContainers.modalBackdrop}>
      <div className={`${AppContainers.modalDialog} max-w-lg p-6`}>
        <div className="flex items-center justify-between pb-4 border-b border-[#262626]">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <Clock className="w-5 h-5" />
            </div>
            <div>
              <h3 className={AppText.h4}>Configure Backup Policy</h3>
              <p className={AppText.subtitle}>
                Automate periodic snapshot schedules and multi-cloud retention rules.
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-md text-[#707070] hover:text-[#ededed] hover:bg-[#262626] transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-5 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-950/60 border border-rose-800/40 text-rose-300 text-xs">
              {error}
            </div>
          )}

          {loadingResources ? (
            <div className="py-8 flex items-center justify-center space-x-2 text-xs text-[#a1a1a1]">
              <Loader2 className="w-4 h-4 animate-spin text-emerald-400" />
              <span>Loading servers and buckets...</span>
            </div>
          ) : (
            <>
              <div>
                <label className={`${AppText.label} mb-1.5`}>
                  Policy Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Daily Production DB Snapshot"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] placeholder-[#707070] text-sm focus:outline-none focus:border-emerald-500 transition-colors"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className={`${AppText.label} mb-1.5`}>
                    Target Server
                  </label>
                  <select
                    value={formData.server_id}
                    onChange={(e) => setFormData({ ...formData, server_id: e.target.value })}
                    className="w-full px-3 py-2 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] text-xs focus:outline-none focus:border-emerald-500 transition-colors"
                  >
                    {servers.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name} ({s.region})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className={`${AppText.label} mb-1.5`}>
                    Destination Bucket
                  </label>
                  <select
                    value={formData.bucket_id || ''}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        bucket_id: e.target.value ? e.target.value : undefined,
                      })
                    }
                    className="w-full px-3 py-2 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] text-xs focus:outline-none focus:border-emerald-500 transition-colors"
                  >
                    <option value="">Default Cluster Storage</option>
                    {buckets.map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.name} ({b.provider_type})
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <label className={`${AppText.label} mb-1.5`}>
                  Backup Frequency Schedule
                </label>
                <div className="grid grid-cols-3 gap-2">
                  {[
                    { cron: '0 2 * * *', label: 'Daily (2:00 AM)' },
                    { cron: '0 */6 * * *', label: 'Every 6 Hours' },
                    { cron: '0 3 * * 0', label: 'Weekly (Sunday)' },
                  ].map((preset) => {
                    const isSelected = formData.cron_expression === preset.cron;
                    return (
                      <button
                        type="button"
                        key={preset.cron}
                        onClick={() =>
                          setFormData({ ...formData, cron_expression: preset.cron })
                        }
                        className={`p-2.5 rounded-lg border text-center text-xs transition-all ${
                          isSelected
                            ? 'border-emerald-500/50 bg-emerald-950/20 text-emerald-300 font-semibold'
                            : 'border-[#262626] bg-[#141414] text-[#a1a1a1] hover:border-[#383838]'
                        }`}
                      >
                        {preset.label}
                      </button>
                    );
                  })}
                </div>
              </div>

              <div>
                <label className={`${AppText.label} mb-1.5`}>
                  Retention Period (Lifecycle Policy)
                </label>
                <select
                  value={formData.retention_days}
                  onChange={(e) =>
                    setFormData({ ...formData, retention_days: Number(e.target.value) })
                  }
                  className="w-full px-3.5 py-2.5 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] text-sm focus:outline-none focus:border-emerald-500 transition-colors"
                >
                  <option value={7}>Keep for 7 Days</option>
                  <option value={14}>Keep for 14 Days (2 Weeks)</option>
                  <option value={30}>Keep for 30 Days (1 Month)</option>
                  <option value={90}>Keep for 90 Days (Quarterly)</option>
                  <option value={365}>Keep for 365 Days (1 Year)</option>
                </select>
                <p className="mt-1 text-[11px] text-[#707070]">
                  Expired snapshots will be purged automatically by background retention cleaner.
                </p>
              </div>

              <div className="flex items-center justify-between p-3 rounded-lg bg-[#141414] border border-[#262626]">
                <div className="flex items-center space-x-2.5">
                  <HardDrive className="w-4 h-4 text-emerald-400" />
                  <div>
                    <p className="text-xs font-medium text-[#ededed]">
                      Include Attached Persistent Disks
                    </p>
                    <p className="text-[11px] text-[#707070]">
                      Take consistent volume block snapshots of all attached volumes.
                    </p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.include_disks}
                  onChange={(e) =>
                    setFormData({ ...formData, include_disks: e.target.checked })
                  }
                  className="w-4 h-4 rounded text-emerald-500 bg-[#171717] border-[#2e2e2e] focus:ring-emerald-500"
                />
              </div>
            </>
          )}

          <div className="pt-4 flex items-center justify-end space-x-3 border-t border-[#262626]">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 rounded-lg text-xs font-medium text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#262626] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting || loadingResources || servers.length === 0}
              className={`px-4 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity disabled:opacity-50`}
            >
              {submitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
              <span>Save Policy</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
