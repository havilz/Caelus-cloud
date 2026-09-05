'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { Bucket } from '@/types/storage';
import { storageService } from '@/services/storage.service';
import { CreateBucketModal } from '@/components/storage/CreateBucketModal';
import {
  Database,
  FolderTree,
  Plus,
  Search,
  ExternalLink,
  Trash2,
  Globe,
  Lock,
  HardDrive,
  Shield,
  Loader2,
  Archive,
  RefreshCw,
} from 'lucide-react';

export default function StorageDashboardPage() {
  const [buckets, setBuckets] = useState<Bucket[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [deletingBucket, setDeletingBucket] = useState<string | null>(null);

  const fetchBuckets = useCallback(async () => {
    try {
      setLoading(true);
      const res = await storageService.listBuckets(1, 100);
      setBuckets(res.data || []);
    } catch (err) {
      console.error('Failed to fetch buckets:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleSyncBuckets = async () => {
    try {
      setSyncing(true);
      const synced = await storageService.syncBuckets();
      setBuckets(synced);
    } catch (err) {
      console.error('Failed to sync cloud buckets:', err);
    } finally {
      setSyncing(false);
    }
  };

  useEffect(() => {
    fetchBuckets();
  }, [fetchBuckets]);

  const handleDeleteBucket = async (bucketName: string) => {
    if (
      !confirm(
        `Are you sure you want to delete bucket "${bucketName}"? The bucket must be empty.`
      )
    ) {
      return;
    }

    try {
      setDeletingBucket(bucketName);
      await storageService.deleteBucket(bucketName);
      await fetchBuckets();
    } catch (err: any) {
      const serverMsg = err?.response?.data?.message || err?.response?.data?.errors || err?.message || '';
      if (serverMsg.includes('NotEmpty') || serverMsg.includes('not empty') || err?.response?.status === 400) {
        alert(`Bucket "${bucketName}" cannot be deleted because it still contains files or objects.\n\nFor your data safety, please navigate into the bucket and delete all objects before deleting the bucket.`);
      } else {
        alert(serverMsg || 'Failed to delete bucket');
      }
    } finally {
      setDeletingBucket(null);
    }
  };

  const filteredBuckets = buckets.filter(
    (b) =>
      b.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      b.provider_type.toLowerCase().includes(searchQuery.toLowerCase()) ||
      b.region.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const publicBucketsCount = buckets.filter((b) => b.is_public).length;
  const uniqueProvidersCount = new Set(buckets.map((b) => b.provider_type)).size;

  return (
    <div className={AppContainers.pageWrapper}>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className={AppText.h2}>Object Storage & Buckets</h1>
          <p className={AppText.subtitle}>
            Multi-cloud S3-compatible object storage infrastructure across MinIO, AWS S3, and Cloudflare R2.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={handleSyncBuckets}
            disabled={syncing}
            className="px-3.5 py-2 rounded-lg bg-[#1a1a1a] border border-[#2e2e2e] text-xs font-medium text-[#ededed] hover:border-cyan-500/40 flex items-center space-x-2 transition-all shadow-sm disabled:opacity-50"
            title="Sync two-way buckets directly with Cloudflare R2, AWS S3, and MinIO"
          >
            <RefreshCw className={`w-3.5 h-3.5 text-cyan-400 ${syncing ? 'animate-spin' : ''}`} />
            <span>{syncing ? 'Syncing...' : 'Sync Cloud Storage'}</span>
          </button>

          <Link
            href="/storage/backups"
            className="px-3.5 py-2 rounded-lg bg-[#1a1a1a] border border-[#2e2e2e] text-xs font-medium text-[#ededed] hover:border-emerald-500/40 flex items-center space-x-2 transition-all shadow-sm"
          >
            <Archive className="w-4 h-4 text-emerald-400" />
            <span>Automated Backups</span>
          </Link>

          <button
            onClick={() => setIsCreateModalOpen(true)}
            className={`px-3.5 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity shadow-sm`}
          >
            <Plus className="w-4 h-4" />
            <span>Create Bucket</span>
          </button>
        </div>
      </div>

      <div className={AppContainers.metricsGrid}>
        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Total Active Buckets</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Database className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>{buckets.length}</p>
          <p className="text-[11px] text-[#707070] mt-1">Multi-tenant object stores</p>
        </div>

        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Connected Providers</span>
            <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
              <HardDrive className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>{uniqueProvidersCount || 1}</p>
          <p className="text-[11px] text-[#707070] mt-1">MinIO, AWS S3, Cloudflare R2</p>
        </div>

        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Public Read Buckets</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400">
              <Globe className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>{publicBucketsCount}</p>
          <p className="text-[11px] text-[#707070] mt-1">Anonymous read endpoints</p>
        </div>

        <div className={`${AppContainers.card} ${AppContainers.cardContent}`}>
          <div className="flex items-center justify-between">
            <span className={AppText.caption}>Disaster Recovery</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Shield className="w-4 h-4" />
            </div>
          </div>
          <p className={`${AppText.h2} mt-2`}>Active</p>
          <p className="text-[11px] text-[#707070] mt-1">Automated backup scheduler</p>
        </div>
      </div>

      <div className={AppContainers.card}>
        <div className="p-4 border-b border-[#262626] flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="relative w-full sm:w-80">
            <Search className="w-4 h-4 text-[#707070] absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search buckets by name, provider..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-3.5 py-1.5 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] placeholder-[#707070] text-xs focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <div className="flex items-center space-x-2 text-xs text-[#707070]">
            <span>Showing {filteredBuckets.length} buckets</span>
          </div>
        </div>

        {loading ? (
          <div className="p-12 flex flex-col items-center justify-center space-y-3">
            <Loader2 className="w-6 h-6 animate-spin text-emerald-400" />
            <p className={AppText.subtitle}>Loading buckets from storage cluster...</p>
          </div>
        ) : filteredBuckets.length === 0 ? (
          <div className="p-12 text-center space-y-3">
            <div className="p-3.5 rounded-full bg-[#1f1f1f] text-[#707070] w-fit mx-auto">
              <Database className="w-6 h-6" />
            </div>
            <div>
              <p className={AppText.body}>No storage buckets found</p>
              <p className={AppText.subtitle}>
                {searchQuery
                  ? 'Try searching with a different keyword.'
                  : 'Create your first S3 bucket to start storing and serving objects.'}
              </p>
            </div>
            {!searchQuery && (
              <button
                onClick={() => setIsCreateModalOpen(true)}
                className={`px-4 py-2 rounded-lg text-xs font-semibold inline-flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity`}
              >
                <Plus className="w-4 h-4" />
                <span>Create First Bucket</span>
              </button>
            )}
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-[#262626] bg-[#141414]/50 text-[11px] font-medium text-[#707070] uppercase tracking-wider">
                  <th className="py-3 px-4">Bucket Name</th>
                  <th className="py-3 px-4">Provider</th>
                  <th className="py-3 px-4">Region</th>
                  <th className="py-3 px-4">Access Policy</th>
                  <th className="py-3 px-4">Created At</th>
                  <th className="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#262626] text-xs">
                {filteredBuckets.map((bucket) => {
                  const isDeleting = deletingBucket === bucket.name;
                  return (
                    <tr key={bucket.id} className="hover:bg-[#1a1a1a]/50 transition-colors">
                      <td className="py-3.5 px-4">
                        <Link
                          href={`/storage/${bucket.name}`}
                          className="flex items-center space-x-2.5 font-medium text-[#ededed] hover:text-emerald-400 transition-colors group"
                        >
                          <div className="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 group-hover:bg-emerald-500/20 transition-colors">
                            <FolderTree className="w-4 h-4" />
                          </div>
                          <span>{bucket.name}</span>
                        </Link>
                      </td>

                      <td className="py-3.5 px-4">
                        <span
                          className={`px-2 py-0.5 rounded-md text-[11px] font-mono uppercase font-semibold border ${
                            bucket.provider_type === 'minio'
                              ? 'bg-rose-950/40 text-rose-300 border-rose-800/30'
                              : bucket.provider_type === 'r2'
                              ? 'bg-amber-950/40 text-amber-300 border-amber-800/30'
                              : 'bg-emerald-950/40 text-emerald-300 border-emerald-800/30'
                          }`}
                        >
                          {bucket.provider_type}
                        </span>
                      </td>

                      <td className="py-3.5 px-4 font-mono text-[#a1a1a1] text-[11px]">
                        {bucket.region}
                      </td>

                      <td className="py-3.5 px-4">
                        {bucket.is_public ? (
                          <span className="inline-flex items-center space-x-1 text-amber-400 text-[11px]">
                            <Globe className="w-3 h-3" />
                            <span>Public Read</span>
                          </span>
                        ) : (
                          <span className="inline-flex items-center space-x-1 text-[#707070] text-[11px]">
                            <Lock className="w-3 h-3" />
                            <span>Private (IAM)</span>
                          </span>
                        )}
                      </td>

                      <td className="py-3.5 px-4 text-[#707070] text-[11px]">
                        {new Date(bucket.created_at).toLocaleDateString(undefined, {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric',
                        })}
                      </td>

                      <td className="py-3.5 px-4 text-right">
                        <div className="flex items-center justify-end space-x-2">
                          <Link
                            href={`/storage/${bucket.name}`}
                            className="px-2.5 py-1 rounded-md bg-[#1f1f1f] border border-[#2e2e2e] text-[11px] font-medium text-[#ededed] hover:border-emerald-500/40 hover:text-emerald-300 flex items-center space-x-1 transition-colors"
                          >
                            <span>Browse</span>
                            <ExternalLink className="w-3 h-3" />
                          </Link>

                          <button
                            onClick={() => handleDeleteBucket(bucket.name)}
                            disabled={isDeleting}
                            className="p-1 rounded-md text-[#707070] hover:text-rose-400 hover:bg-rose-950/30 transition-colors disabled:opacity-50"
                            title="Delete Bucket"
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

      <CreateBucketModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSuccess={fetchBuckets}
      />
    </div>
  );
}
