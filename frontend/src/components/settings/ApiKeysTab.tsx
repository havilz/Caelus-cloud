"use client";

import React, { useState, useEffect } from "react";
import { Key, Plus, Trash2, Copy, Check, AlertCircle, RefreshCw, X, Shield, Clock } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { settingsService } from "@/services/settings.service";
import { APIKey } from "@/types/settings";

export const ApiKeysTab: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [keyName, setKeyName] = useState("");
  const [expiresInDays, setExpiresInDays] = useState(30);
  const [isCreating, setIsCreating] = useState(false);
  const [createModalError, setCreateModalError] = useState<string | null>(null);

  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [hasCopied, setHasCopied] = useState(false);

  const fetchApiKeys = async () => {
    try {
      setIsLoading(true);
      const data = await settingsService.listAPIKeys();
      setApiKeys(data || []);
    } catch (err: any) {
      setErrorMsg("Failed to load Personal Access Tokens");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchApiKeys();
  }, []);

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setIsCreating(true);
      setCreateModalError(null);
      const result = await settingsService.createAPIKey({
        name: keyName,
        scopes: ["read", "write"],
        expires_in_days: expiresInDays > 0 ? expiresInDays : undefined,
      });

      setCreatedToken(result.raw_token || null);
      setIsCreateModalOpen(false);
      setKeyName("");
      fetchApiKeys();
    } catch (err: any) {
      setCreateModalError(err.response?.data?.message || "Failed to create API key");
    } finally {
      setIsCreating(false);
    }
  };

  const handleDeleteKey = async (id: string, name: string) => {
    if (!confirm(`Revoke API Key "${name}"? This key will no longer be usable.`)) return;
    try {
      await settingsService.deleteAPIKey(id);
      setSuccessMsg("API key revoked successfully");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchApiKeys();
    } catch (err: any) {
      setErrorMsg("Failed to revoke API key");
      setTimeout(() => setErrorMsg(null), 3000);
    }
  };

  const handleCopyToken = () => {
    if (!createdToken) return;
    navigator.clipboard.writeText(createdToken);
    setHasCopied(true);
    setTimeout(() => setHasCopied(false), 2500);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16 text-zinc-500">
        <RefreshCw className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Loading Personal Access Tokens...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {createdToken && (
        <div className="p-4 rounded-xl bg-amber-950/40 border border-amber-500/40 text-amber-200 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 font-semibold text-sm text-amber-400">
              <Key className="h-4 w-4" />
              <span>New Personal Access Token Generated Successfully</span>
            </div>
            <button
              type="button"
              onClick={() => setCreatedToken(null)}
              className="text-amber-400 hover:text-amber-200 text-xs cursor-pointer"
            >
              Close
            </button>
          </div>
          <p className="text-xs text-amber-300/80">
            Please copy this token now. For security reasons, this token <strong>will never be displayed again</strong> after leaving this page.
          </p>
          <div className="flex items-center gap-2">
            <input
              type="text"
              readOnly
              value={createdToken}
              className="w-full bg-[#111111] border border-amber-500/30 text-amber-300 font-mono text-xs rounded-lg px-3 py-2 select-all"
            />
            <button
              type="button"
              onClick={handleCopyToken}
              className="px-4 py-2 bg-amber-500 hover:bg-amber-400 text-zinc-950 font-bold text-xs rounded-lg transition-colors flex items-center gap-1.5 cursor-pointer shrink-0"
            >
              {hasCopied ? <Check className="h-4 w-4 text-emerald-950" /> : <Copy className="h-4 w-4" />}
              <span>{hasCopied ? "Copied!" : "Copy"}</span>
            </button>
          </div>
        </div>
      )}

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-sm font-semibold text-zinc-100">Personal Access Tokens & Developer API Keys</h3>
          <p className="text-xs text-zinc-400 mt-0.5">
            Use API tokens to authenticate CLI requests, CI/CD pipelines, and automation scripts with Caelus Cloud
          </p>
        </div>
        <button
          type="button"
          onClick={() => setIsCreateModalOpen(true)}
          className="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer shrink-0"
        >
          <Plus className="h-4 w-4" />
          <span>Generate New Token</span>
        </button>
      </div>

      {successMsg && (
        <div className="p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
          <Check className="h-4 w-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      {errorMsg && (
        <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      {apiKeys.length === 0 ? (
        <div className={`${AppTheme.containers.card} text-center py-12 space-y-3`}>
          <div className="mx-auto w-10 h-10 rounded-full bg-zinc-800 flex items-center justify-center text-zinc-500">
            <Key className="h-5 w-5" />
          </div>
          <p className="text-xs text-zinc-400">No API Keys generated yet.</p>
        </div>
      ) : (
        <div className={`${AppTheme.containers.card} overflow-hidden p-0`}>
          <div className="divide-y divide-[#222222]">
            {apiKeys.map((key) => (
              <div key={key.id} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-[#161616]/50 transition-colors">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-semibold text-zinc-200">{key.name}</span>
                    <span className="px-2 py-0.5 rounded font-mono text-[10px] bg-zinc-800 text-zinc-400 border border-zinc-700">
                      caelus_pat_{key.key_prefix}...
                    </span>
                  </div>
                  <div className="flex items-center gap-4 text-[11px] text-zinc-400">
                    <span className="flex items-center gap-1">
                      <Clock className="h-3.5 w-3.5" />
                      Created: {new Date(key.created_at).toLocaleDateString()}
                    </span>
                    {key.last_used_at ? (
                      <span>Last used: {new Date(key.last_used_at).toLocaleDateString()}</span>
                    ) : (
                      <span className="text-zinc-400">Never used</span>
                    )}
                    {key.expires_at && (
                      <span className="text-amber-400/80">Expires: {new Date(key.expires_at).toLocaleDateString()}</span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleDeleteKey(key.id, key.name)}
                    className="p-1.5 rounded-lg text-zinc-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                    title="Revoke Token"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {isCreateModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4">
          <div className="bg-[#141414] border border-[#2e2e2e] rounded-xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#262626] pb-3">
              <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <Key className="h-4 w-4 text-emerald-400" />
                Generate New Personal Access Token
              </h3>
              <button
                type="button"
                onClick={() => setIsCreateModalOpen(false)}
                className="text-zinc-400 hover:text-zinc-200 p-1 rounded-lg transition-colors cursor-pointer"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {createModalError && (
              <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{createModalError}</span>
              </div>
            )}

            <form onSubmit={handleCreateKey} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Descriptive Token Name</label>
                <input
                  type="text"
                  required
                  value={keyName}
                  onChange={(e) => setKeyName(e.target.value)}
                  placeholder="e.g. GitHub Actions Deployer, CLI Laptop"
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Expiration Period</label>
                <select
                  value={expiresInDays}
                  onChange={(e) => setExpiresInDays(Number(e.target.value))}
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 cursor-pointer"
                >
                  <option value={7}>7 Days</option>
                  <option value={30}>30 Days (Recommended)</option>
                  <option value={90}>90 Days</option>
                  <option value={365}>1 Year</option>
                  <option value={0}>Never expires</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCreateModalOpen(false)}
                  className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-lg transition-colors cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
                >
                  {isCreating && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
                  <span>Generate Token</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
