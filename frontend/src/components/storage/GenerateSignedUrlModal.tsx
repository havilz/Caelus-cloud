'use client';

import React, { useState } from 'react';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { SignedURLOperation, SignedURLResponse } from '@/types/storage';
import { storageService } from '@/services/storage.service';
import { X, KeyRound, Copy, Check, ExternalLink, Loader2 } from 'lucide-react';

interface GenerateSignedUrlModalProps {
  isOpen: boolean;
  bucketName: string;
  objectKey: string;
  onClose: () => void;
}

export const GenerateSignedUrlModal: React.FC<GenerateSignedUrlModalProps> = ({
  isOpen,
  bucketName,
  objectKey,
  onClose,
}) => {
  const [operation, setOperation] = useState<SignedURLOperation>('download');
  const [expiryMinutes, setExpiryMinutes] = useState(60);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SignedURLResponse | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleGenerate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await storageService.generateSignedURL(bucketName, {
        key: objectKey,
        operation,
        expiry_minutes: expiryMinutes,
      });
      setResult(res);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to generate signed URL';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = () => {
    if (result) {
      navigator.clipboard.writeText(result.url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    }
  };

  return (
    <div className={AppContainers.modalBackdrop}>
      <div className={`${AppContainers.modalDialog} max-w-lg p-6`}>
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-[#262626]">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <KeyRound className="w-5 h-5" />
            </div>
            <div>
              <h3 className={AppText.h4}>Generate Signed URL</h3>
              <p className={AppText.subtitle}>
                Create a time-bounded presigned URL with cryptographic HMAC signature.
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

        {/* Form Body */}
        <form onSubmit={handleGenerate} className="mt-5 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-950/60 border border-rose-800/40 text-rose-300 text-xs">
              {error}
            </div>
          )}

          {/* Target Key Info */}
          <div>
            <label className={`${AppText.caption} block mb-1`}>
              Target Object
            </label>
            <div className="p-2.5 rounded-lg bg-[#141414] border border-[#262626] text-xs font-mono text-emerald-400 truncate">
              {bucketName} / {objectKey}
            </div>
          </div>

          {/* Operation Selector */}
          <div>
            <label className={`${AppText.label} mb-1.5`}>
              Permission / Operation Type
            </label>
            <div className="grid grid-cols-2 gap-3">
              {[
                { id: 'download', name: 'Download (GET)', desc: 'Grant temporary read access' },
                { id: 'upload', name: 'Upload (PUT)', desc: 'Grant temporary upload access' },
              ].map((op) => {
                const isSelected = operation === op.id;
                return (
                  <button
                    type="button"
                    key={op.id}
                    onClick={() => {
                      setOperation(op.id as SignedURLOperation);
                      setResult(null);
                    }}
                    className={`p-3 rounded-xl border text-left transition-all ${
                      isSelected
                        ? 'border-emerald-500/50 bg-emerald-950/20 text-emerald-300'
                        : 'border-[#262626] bg-[#141414] text-[#a1a1a1] hover:border-[#383838]'
                    }`}
                  >
                    <p className="text-xs font-semibold text-[#ededed]">{op.name}</p>
                    <p className="text-[10px] text-[#707070] mt-0.5">{op.desc}</p>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Expiration Duration */}
          <div>
            <label className={`${AppText.label} mb-1.5`}>
              Link Expiration
            </label>
            <select
              value={expiryMinutes}
              onChange={(e) => {
                setExpiryMinutes(Number(e.target.value));
                setResult(null);
              }}
              className="w-full px-3.5 py-2.5 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] text-sm focus:outline-none focus:border-emerald-500 transition-colors"
            >
              <option value={15}>15 Minutes</option>
              <option value={60}>1 Hour</option>
              <option value={360}>6 Hours</option>
              <option value={1440}>24 Hours</option>
              <option value={10080}>7 Days</option>
            </select>
          </div>

          {/* Generate Button */}
          {!result && (
            <div className="pt-2">
              <button
                type="submit"
                disabled={loading}
                className={`w-full py-2.5 rounded-lg text-xs font-semibold flex items-center justify-center space-x-2 ${AppColors.brand.primary} transition-opacity disabled:opacity-50`}
              >
                {loading && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                <span>Generate Signed URL</span>
              </button>
            </div>
          )}

          {/* Result Section */}
          {result && (
            <div className="p-4 rounded-xl bg-[#141414] border border-emerald-500/30 space-y-3 animate-in fade-in">
              <div className="flex items-center justify-between">
                <span className="text-[11px] text-emerald-400 font-semibold uppercase tracking-wider">
                  Generated Presigned URL
                </span>
                <span className="text-[11px] text-[#707070]">
                  Expires in {Math.round(result.expires_in_sec / 60)} min
                </span>
              </div>
              <div className="p-2.5 rounded-lg bg-[#0d0d0d] border border-[#262626] text-xs font-mono text-[#ededed] break-all max-h-24 overflow-y-auto custom-scrollbar">
                {result.url}
              </div>
              <div className="flex items-center justify-end space-x-2 pt-1">
                <a
                  href={result.url}
                  target="_blank"
                  rel="noreferrer"
                  className="px-3 py-1.5 rounded-lg bg-[#1f1f1f] border border-[#2e2e2e] text-xs text-[#ededed] hover:border-[#3e3e3e] flex items-center space-x-1.5 transition-colors"
                >
                  <ExternalLink className="w-3.5 h-3.5 text-[#a1a1a1]" />
                  <span>Open URL</span>
                </a>
                <button
                  type="button"
                  onClick={handleCopy}
                  className={`px-3 py-1.5 rounded-lg text-xs font-semibold flex items-center space-x-1.5 ${
                    copied
                      ? 'bg-emerald-500 text-zinc-950'
                      : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 hover:bg-emerald-500/20'
                  } transition-all`}
                >
                  {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copied ? 'Copied to Clipboard' : 'Copy URL'}</span>
                </button>
              </div>
            </div>
          )}
        </form>
      </div>
    </div>
  );
};
