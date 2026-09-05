'use client';

import React, { useState, useEffect, useCallback, use } from 'react';
import Link from 'next/link';
import { AppContainers, AppColors, AppText } from '@/core/theme';
import { ObjectItem } from '@/types/storage';
import { storageService } from '@/services/storage.service';
import { UploadObjectModal } from '@/components/storage/UploadObjectModal';
import { GenerateSignedUrlModal } from '@/components/storage/GenerateSignedUrlModal';
import {
  Folder,
  File,
  ChevronRight,
  UploadCloud,
  KeyRound,
  Download,
  Trash2,
  ArrowLeft,
  Search,
  Loader2,
  Database,
  RefreshCw,
} from 'lucide-react';

interface StorageExplorerPageProps {
  params: Promise<{ bucket: string }>;
}

export default function StorageExplorerPage({ params }: StorageExplorerPageProps) {
  const resolvedParams = use(params);
  const bucketName = resolvedParams.bucket;

  const [currentPrefix, setCurrentPrefix] = useState('');
  const [folders, setFolders] = useState<string[]>([]);
  const [objects, setObjects] = useState<ObjectItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [signedUrlTarget, setSignedUrlTarget] = useState<string | null>(null);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);

  const fetchObjects = useCallback(async () => {
    try {
      setLoading(true);
      const res = await storageService.listObjects(bucketName, currentPrefix, '/', 1000);
      setFolders(res.folders || []);
      setObjects(res.objects || []);
    } catch (err) {
      console.error('Failed to list objects:', err);
    } finally {
      setLoading(false);
    }
  }, [bucketName, currentPrefix]);

  useEffect(() => {
    fetchObjects();
  }, [fetchObjects]);

  const handleDownload = async (key: string) => {
    try {
      const blob = await storageService.downloadObject(bucketName, key);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = key.split('/').pop() || key;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      console.error('Download failed:', err);
      alert('Failed to download object');
    }
  };

  const handleDelete = async (key: string) => {
    if (!confirm(`Are you sure you want to delete object "${key}"?`)) {
      return;
    }

    try {
      setDeletingKey(key);
      await storageService.deleteObject(bucketName, key);
      await fetchObjects();
    } catch (err) {
      console.error('Delete failed:', err);
      alert('Failed to delete object');
    } finally {
      setDeletingKey(null);
    }
  };

  const breadcrumbParts = currentPrefix.split('/').filter(Boolean);

  const navigateToBreadcrumb = (index: number) => {
    if (index === -1) {
      setCurrentPrefix('');
    } else {
      const newPrefix = breadcrumbParts.slice(0, index + 1).join('/') + '/';
      setCurrentPrefix(newPrefix);
    }
  };

  const filteredObjects = objects.filter((obj) =>
    obj.key.toLowerCase().includes(searchQuery.toLowerCase())
  );
  const filteredFolders = folders.filter((f) =>
    f.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className={AppContainers.pageWrapper}>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center space-x-3">
          <Link
            href="/storage"
            className="p-2 rounded-lg bg-[#1a1a1a] border border-[#2e2e2e] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#3e3e3e] transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
          </Link>
          <div>
            <div className="flex items-center space-x-2">
              <Database className="w-4 h-4 text-emerald-400" />
              <h1 className={AppText.h3}>{bucketName}</h1>
            </div>
            <p className={AppText.subtitle}>Object Browser & Virtual File Explorer</p>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={fetchObjects}
            className="p-2 rounded-lg bg-[#1a1a1a] border border-[#2e2e2e] text-[#a1a1a1] hover:text-[#ededed] transition-colors"
            title="Refresh Object List"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>

          <button
            onClick={() => setIsUploadOpen(true)}
            className={`px-3.5 py-2 rounded-lg text-xs font-semibold flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity shadow-sm`}
          >
            <UploadCloud className="w-4 h-4" />
            <span>Upload Object</span>
          </button>
        </div>
      </div>

      <div className={AppContainers.card}>
        <div className="p-4 border-b border-[#262626] flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="flex items-center space-x-1.5 overflow-x-auto w-full sm:w-auto text-xs custom-scrollbar py-1">
            <button
              onClick={() => navigateToBreadcrumb(-1)}
              className={`px-2 py-1 rounded-md font-medium transition-colors ${
                currentPrefix === ''
                  ? 'text-emerald-400 bg-emerald-950/40'
                  : 'text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#262626]'
              }`}
            >
              Root
            </button>

            {breadcrumbParts.map((part, idx) => (
              <React.Fragment key={idx}>
                <ChevronRight className="w-3.5 h-3.5 text-[#707070] flex-shrink-0" />
                <button
                  onClick={() => navigateToBreadcrumb(idx)}
                  className={`px-2 py-1 rounded-md font-medium transition-colors whitespace-nowrap ${
                    idx === breadcrumbParts.length - 1
                      ? 'text-emerald-400 bg-emerald-950/40'
                      : 'text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#262626]'
                  }`}
                >
                  {part}
                </button>
              </React.Fragment>
            ))}
          </div>

          <div className="relative w-full sm:w-72">
            <Search className="w-4 h-4 text-[#707070] absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search current folder..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-3.5 py-1.5 rounded-lg bg-[#141414] border border-[#2e2e2e] text-[#ededed] placeholder-[#707070] text-xs focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>
        </div>

        {loading ? (
          <div className="p-12 flex flex-col items-center justify-center space-y-3">
            <Loader2 className="w-6 h-6 animate-spin text-emerald-400" />
            <p className={AppText.subtitle}>Loading objects and folders...</p>
          </div>
        ) : filteredFolders.length === 0 && filteredObjects.length === 0 ? (
          <div className="p-12 text-center space-y-3">
            <div className="p-3.5 rounded-full bg-[#1f1f1f] text-[#707070] w-fit mx-auto">
              <Folder className="w-6 h-6" />
            </div>
            <div>
              <p className={AppText.body}>This folder is empty</p>
              <p className={AppText.subtitle}>
                Upload files or create objects to populate this directory.
              </p>
            </div>
            <button
              onClick={() => setIsUploadOpen(true)}
              className={`px-4 py-2 rounded-lg text-xs font-semibold inline-flex items-center space-x-2 ${AppColors.brand.primary} transition-opacity`}
            >
              <UploadCloud className="w-4 h-4" />
              <span>Upload File Here</span>
            </button>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-[#262626] bg-[#141414]/50 text-[11px] font-medium text-[#707070] uppercase tracking-wider">
                  <th className="py-3 px-4">Name</th>
                  <th className="py-3 px-4">Size</th>
                  <th className="py-3 px-4">Type</th>
                  <th className="py-3 px-4">Last Modified</th>
                  <th className="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#262626] text-xs">
                {filteredFolders.map((folderPath) => {
                  const folderDisplayName = folderPath.replace(currentPrefix, '');
                  return (
                    <tr
                      key={folderPath}
                      onClick={() => setCurrentPrefix(folderPath)}
                      className="hover:bg-[#1a1a1a]/60 cursor-pointer transition-colors"
                    >
                      <td className="py-3 px-4" colSpan={4}>
                        <div className="flex items-center space-x-2.5 font-medium text-[#ededed] hover:text-emerald-400">
                          <Folder className="w-4 h-4 text-emerald-400 fill-emerald-400/20" />
                          <span>{folderDisplayName}</span>
                        </div>
                      </td>
                      <td className="py-3 px-4 text-right text-[#707070] text-[11px]">Folder</td>
                    </tr>
                  );
                })}

                {filteredObjects.map((obj) => {
                  const fileName = obj.key.replace(currentPrefix, '');
                  const isDeleting = deletingKey === obj.key;

                  return (
                    <tr key={obj.key} className="hover:bg-[#1a1a1a]/40 transition-colors">
                      <td className="py-3 px-4">
                        <div className="flex items-center space-x-2.5">
                          <File className="w-4 h-4 text-[#707070]" />
                          <span className="font-medium text-[#ededed] truncate max-w-xs sm:max-w-md">
                            {fileName}
                          </span>
                        </div>
                      </td>

                      <td className="py-3 px-4 font-mono text-[#a1a1a1] text-[11px]">
                        {formatFileSize(obj.size)}
                      </td>

                      <td className="py-3 px-4 text-[#707070] text-[11px]">
                        {obj.content_type || 'application/octet-stream'}
                      </td>

                      <td className="py-3 px-4 text-[#707070] text-[11px]">
                        {new Date(obj.last_modified).toLocaleString(undefined, {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </td>

                      <td className="py-3 px-4 text-right">
                        <div className="flex items-center justify-end space-x-1.5">
                          <button
                            onClick={() => setSignedUrlTarget(obj.key)}
                            className="p-1.5 rounded-md text-[#707070] hover:text-cyan-400 hover:bg-cyan-950/30 transition-colors"
                            title="Generate Signed URL"
                          >
                            <KeyRound className="w-3.5 h-3.5" />
                          </button>

                          <button
                            onClick={() => handleDownload(obj.key)}
                            className="p-1.5 rounded-md text-[#707070] hover:text-emerald-400 hover:bg-emerald-950/30 transition-colors"
                            title="Download File"
                          >
                            <Download className="w-3.5 h-3.5" />
                          </button>

                          <button
                            onClick={() => handleDelete(obj.key)}
                            disabled={isDeleting}
                            className="p-1.5 rounded-md text-[#707070] hover:text-rose-400 hover:bg-rose-950/30 transition-colors disabled:opacity-50"
                            title="Delete File"
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

      <UploadObjectModal
        isOpen={isUploadOpen}
        bucketName={bucketName}
        currentPrefix={currentPrefix}
        onClose={() => setIsUploadOpen(false)}
        onSuccess={fetchObjects}
      />

      {signedUrlTarget && (
        <GenerateSignedUrlModal
          isOpen={true}
          bucketName={bucketName}
          objectKey={signedUrlTarget}
          onClose={() => setSignedUrlTarget(null)}
        />
      )}
    </div>
  );
}
