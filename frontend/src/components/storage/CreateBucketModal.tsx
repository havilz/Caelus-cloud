'use client';

import React, { useState } from 'react';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { CreateBucketInput, StorageProviderType } from '@/types/storage';
import { storageService } from '@/services/storage.service';
import { X, Database, Globe, Lock, ShieldCheck, Loader2 } from 'lucide-react';

interface CreateBucketModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const CreateBucketModal: React.FC<CreateBucketModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [formData, setFormData] = useState<CreateBucketInput>({
    name: '',
    provider_type: 'minio',
    region: 'us-east-1',
    is_public: false,
    versioning: false,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    // Validasi nama bucket S3
    const bucketNameRegex = /^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$/;
    if (!bucketNameRegex.test(formData.name)) {
      setError(
        'Bucket name must be 3-63 characters, lowercase alphanumeric, dots or hyphens only.'
      );
      setLoading(false);
      return;
    }

    try {
      await storageService.createBucket(formData);
      onSuccess();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create bucket';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={AppContainers.modalBackdrop}>
      <div className={`${AppContainers.modalDialog} max-w-lg overflow-hidden`}>
        {/* Header */}
        <div className="flex items-center justify-between p-5 border-b border-[#262626] bg-[#171717]">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <Database className="w-5 h-5" />
            </div>
            <div>
              <h3 className={AppText.h4}>Create Object Bucket</h3>
              <p className={AppText.subtitle}>
                Provision a high-throughput multi-cloud S3 storage bucket.
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-[#707070] hover:text-[#ededed] hover:bg-[#262626] transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex flex-col flex-1 overflow-hidden">
          {/* Scrollable Body */}
          <div className="p-5 overflow-y-auto max-h-[58vh] space-y-4 custom-scrollbar">
            {error && (
              <div className="p-3 rounded-lg bg-rose-950/60 border border-rose-800/40 text-rose-300 text-xs flex items-center space-x-2">
                <span>{error}</span>
              </div>
            )}

            {/* Bucket Name */}
            <div>
              <label className={`${AppText.label} mb-1.5 block`}>
                Bucket Name
              </label>
              <input
                type="text"
                required
                placeholder="e.g. production-assets"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value.toLowerCase().trim() })
                }
                className="w-full px-3.5 py-2 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] placeholder-[#707070] text-sm focus:outline-none focus:border-emerald-500 transition-colors"
              />
              <p className="mt-1 text-[11px] text-[#707070]">
                Must be globally unique, lowercase alphanumeric and hyphens.
              </p>
            </div>

            {/* Provider Selection */}
            <div>
              <label className={`${AppText.label} mb-1.5 block`}>
                Storage Provider
              </label>
              <div className="grid grid-cols-3 gap-2">
                {[
                  { id: 'minio', name: 'MinIO', desc: 'Local Cluster' },
                  { id: 's3', name: 'AWS S3', desc: 'Global Cloud' },
                  { id: 'r2', name: 'Cloudflare R2', desc: 'Zero Egress' },
                ].map((prov) => {
                  const isSelected = formData.provider_type === prov.id;
                  return (
                    <button
                      type="button"
                      key={prov.id}
                      onClick={() =>
                        setFormData({
                          ...formData,
                          provider_type: prov.id as StorageProviderType,
                          region: prov.id === 'r2' ? 'auto' : 'us-east-1',
                        })
                      }
                      className={`p-2.5 rounded-xl border text-left transition-all ${
                        isSelected
                          ? 'border-emerald-500/50 bg-emerald-950/20 text-emerald-300 ring-1 ring-emerald-500/30'
                          : 'border-[#262626] bg-[#141414] text-[#a1a1a1] hover:border-[#383838]'
                      }`}
                    >
                      <p className="text-xs font-semibold text-[#ededed]">{prov.name}</p>
                      <p className="text-[10px] text-[#707070] mt-0.5">{prov.desc}</p>
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Region */}
            <div>
              <label className={`${AppText.label} mb-1.5 block`}>
                Storage Region
              </label>
              <select
                value={formData.region}
                onChange={(e) => setFormData({ ...formData, region: e.target.value })}
                className="w-full px-3.5 py-2 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] text-sm focus:outline-none focus:border-emerald-500 transition-colors"
              >
                <option value="us-east-1">US East (N. Virginia)</option>
                <option value="ap-southeast-1">Asia Pacific (Singapore)</option>
                <option value="eu-central-1">Europe (Frankfurt)</option>
                <option value="auto">Automatic (Closest Edge)</option>
              </select>
            </div>

            {/* Toggles */}
            <div className="space-y-2.5 pt-1">
              {/* Public Access */}
              <div className="flex items-center justify-between p-3 rounded-lg bg-[#141414] border border-[#262626]">
                <div className="flex items-center space-x-2.5">
                  {formData.is_public ? (
                    <Globe className="w-4 h-4 text-amber-400" />
                  ) : (
                    <Lock className="w-4 h-4 text-emerald-400" />
                  )}
                  <div>
                    <p className="text-xs font-medium text-[#ededed]">Public Read Access</p>
                    <p className="text-[11px] text-[#707070]">
                      Allow anonymous read access to uploaded media.
                    </p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.is_public}
                  onChange={(e) => setFormData({ ...formData, is_public: e.target.checked })}
                  className="w-4 h-4 rounded text-emerald-500 bg-[#171717] border-[#2e2e2e] focus:ring-emerald-500 cursor-pointer"
                />
              </div>

              {/* Versioning */}
              <div className="flex items-center justify-between p-3 rounded-lg bg-[#141414] border border-[#262626]">
                <div className="flex items-center space-x-2.5">
                  <ShieldCheck className="w-4 h-4 text-cyan-400" />
                  <div>
                    <p className="text-xs font-medium text-[#ededed]">Object Versioning</p>
                    <p className="text-[11px] text-[#707070]">
                      Keep multiple variants of an object in the same bucket.
                    </p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.versioning}
                  onChange={(e) => setFormData({ ...formData, versioning: e.target.checked })}
                  className="w-4 h-4 rounded text-emerald-500 bg-[#171717] border-[#2e2e2e] focus:ring-emerald-500 cursor-pointer"
                />
              </div>
            </div>
          </div>

          {/* Sticky Footer Actions */}
          <div className="p-4 px-5 flex items-center justify-end space-x-3 border-t border-[#262626] bg-[#141414]">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-xs font-medium text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#262626] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className={`px-4 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity disabled:opacity-50 shadow-lg shadow-emerald-500/10`}
            >
              {loading && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
              <span>Create Bucket</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
