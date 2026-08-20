'use client';

import React, { useState, useRef } from 'react';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { storageService } from '@/services/storage.service';
import { X, UploadCloud, File, CheckCircle2, Loader2, AlertCircle } from 'lucide-react';

interface UploadObjectModalProps {
  isOpen: boolean;
  bucketName: string;
  currentPrefix: string;
  onClose: () => void;
  onSuccess: () => void;
}

export const UploadObjectModal: React.FC<UploadObjectModalProps> = ({
  isOpen,
  bucketName,
  currentPrefix,
  onClose,
  onSuccess,
}) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [destinationPath, setDestinationPath] = useState(currentPrefix);
  const [progress, setProgress] = useState(0);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const handleFileDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      setSelectedFile(e.dataTransfer.files[0]);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile) return;

    setUploading(true);
    setError(null);
    setProgress(0);

    let finalKey = selectedFile.name;
    if (destinationPath.trim()) {
      const sanitizedPrefix = destinationPath.trim().replace(/^\/+/, '');
      finalKey = sanitizedPrefix.endsWith('/')
        ? `${sanitizedPrefix}${selectedFile.name}`
        : `${sanitizedPrefix}/${selectedFile.name}`;
    }

    try {
      await storageService.uploadObject(
        bucketName,
        selectedFile,
        finalKey,
        (percent) => setProgress(percent)
      );
      onSuccess();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Upload failed';
      setError(msg);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className={AppContainers.modalBackdrop}>
      <div className={`${AppContainers.modalDialog} max-w-md p-6`}>
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-[#262626]">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <UploadCloud className="w-5 h-5" />
            </div>
            <div>
              <h3 className={AppText.h4}>Upload Object</h3>
              <p className={AppText.subtitle}>
                Target Bucket: <span className="text-emerald-400 font-mono">{bucketName}</span>
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            disabled={uploading}
            className="p-1 rounded-md text-[#707070] hover:text-[#ededed] hover:bg-[#262626] transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="mt-5 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-950/60 border border-rose-800/40 text-rose-300 text-xs flex items-center space-x-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Drag and Drop Zone */}
          <div
            onDragOver={(e) => {
              e.preventDefault();
              setIsDragOver(true);
            }}
            onDragLeave={() => setIsDragOver(false)}
            onDrop={handleFileDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors ${
              isDragOver
                ? 'border-emerald-500 bg-emerald-950/20'
                : selectedFile
                ? 'border-emerald-500/40 bg-[#141414]'
                : 'border-[#2e2e2e] bg-[#141414] hover:border-[#3e3e3e]'
            }`}
          >
            <input
              ref={fileInputRef}
              type="file"
              onChange={handleFileSelect}
              className="hidden"
            />
            {selectedFile ? (
              <div className="flex flex-col items-center space-y-2">
                <div className="p-3 rounded-full bg-emerald-500/10 text-emerald-400">
                  <File className="w-6 h-6" />
                </div>
                <div>
                  <p className="text-xs font-semibold text-[#ededed]">{selectedFile.name}</p>
                  <p className="text-[11px] text-[#707070] mt-0.5">
                    {(selectedFile.size / (1024 * 1024)).toFixed(2)} MB
                  </p>
                </div>
                <span className="text-[10px] text-emerald-400 hover:underline">
                  Click to choose a different file
                </span>
              </div>
            ) : (
              <div className="flex flex-col items-center space-y-2">
                <div className="p-3 rounded-full bg-[#1c1c1c] text-[#707070]">
                  <UploadCloud className="w-6 h-6" />
                </div>
                <div>
                  <p className="text-xs font-medium text-[#ededed]">
                    Drag and drop file here, or{' '}
                    <span className="text-emerald-400 underline">browse</span>
                  </p>
                  <p className="text-[11px] text-[#707070] mt-0.5">
                    Supports all file extensions up to 100MB
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Destination Path */}
          <div>
            <label className={`${AppText.label} mb-1.5`}>
              Destination Path / Virtual Prefix
            </label>
            <input
              type="text"
              placeholder="e.g. documents/reports/"
              value={destinationPath}
              onChange={(e) => setDestinationPath(e.target.value)}
              className="w-full px-3.5 py-2.5 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] placeholder-[#707070] text-sm focus:outline-none focus:border-emerald-500 font-mono text-xs transition-colors"
            />
          </div>

          {/* Progress Bar */}
          {uploading && (
            <div className="space-y-1.5 pt-1">
              <div className="flex justify-between text-xs font-medium">
                <span className="text-[#a1a1a1]">Uploading stream...</span>
                <span className="text-emerald-400 font-mono">{progress}%</span>
              </div>
              <div className="w-full h-1.5 rounded-full bg-[#262626] overflow-hidden">
                <div
                  className="h-full bg-emerald-500 transition-all duration-150"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          )}

          {/* Footer Actions */}
          <div className="pt-4 flex items-center justify-end space-x-3 border-t border-[#262626]">
            <button
              type="button"
              onClick={onClose}
              disabled={uploading}
              className="px-4 py-2 rounded-lg text-xs font-medium text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#262626] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!selectedFile || uploading}
              className={`px-4 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity disabled:opacity-50`}
            >
              {uploading ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  <span>Uploading ({progress}%)</span>
                </>
              ) : (
                <>
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>Upload File</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
