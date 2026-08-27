"use client";

import React, { useState, useEffect } from "react";
import { X, Globe, Server, Shield, Cloud, ArrowRight, Loader2, Info } from "lucide-react";
import { CreateDomainRequest, IngressTargetType, domainService } from "@/services/domain.service";
import { serverService } from "@/services/server.service";
import { Server as ServerType } from "@/types/server";

interface AddDomainModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const AddDomainModal: React.FC<AddDomainModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [domainName, setDomainName] = useState("");
  const [serverId, setServerId] = useState<string>("");
  const [targetType, setTargetType] = useState<IngressTargetType>("container");
  const [targetId, setTargetId] = useState("");
  const [targetPort, setTargetPort] = useState<number>(3000);
  const [autoSSL, setAutoSSL] = useState(true);
  const [cloudflareManaged, setCloudflareManaged] = useState(false);

  const [servers, setServers] = useState<ServerType[]>([]);
  const [loadingServers, setLoadingServers] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      loadServers();
      setError(null);
    }
  }, [isOpen]);

  const loadServers = async () => {
    try {
      setLoadingServers(true);
      const res = await serverService.listServers(1, 50);
      const serverList = res.data || [];
      setServers(serverList);
      if (serverList.length > 0 && !serverId) {
        setServerId(serverList[0].id);
      }
    } catch (err: any) {
      console.error("Failed to load servers:", err);
    } finally {
      setLoadingServers(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!domainName.trim()) {
      setError("Domain name is required.");
      return;
    }
    if (!targetId.trim()) {
      setError("Target container/service identifier is required.");
      return;
    }

    try {
      setSubmitting(true);
      setError(null);

      const payload: CreateDomainRequest = {
        server_id: serverId ? serverId : undefined,
        domain_name: domainName.trim().toLowerCase(),
        target_type: targetType,
        target_id: targetId.trim(),
        target_port: Number(targetPort) || 80,
        auto_ssl: autoSSL,
        cloudflare_dns_managed: cloudflareManaged,
      };

      await domainService.createDomain(payload);
      onSuccess();
      onClose();
    } catch (err: any) {
      setError(err?.response?.data?.message || err?.message || "Failed to register custom domain");
    } finally {
      setSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 overflow-y-auto">
      <div className="w-full max-w-lg rounded-xl border border-[#262626] bg-[#121212] p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between border-b border-[#262626] pb-4 mb-5">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <Globe className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-base font-semibold text-[#ededed]">Add Custom Domain</h2>
              <p className="text-xs text-zinc-400">Map a public domain to your container or host service</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-zinc-400 hover:bg-[#1a1a1a] hover:text-zinc-200 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-400">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">
              Domain Name / FQDN <span className="text-emerald-400">*</span>
            </label>
            <div className="relative">
              <input
                type="text"
                placeholder="app.mycompany.com or mysite.id"
                value={domainName}
                onChange={(e) => setDomainName(e.target.value)}
                className="w-full rounded-lg border border-[#262626] bg-[#181818] px-3.5 py-2.5 text-sm text-[#ededed] placeholder-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                required
              />
            </div>
            <p className="mt-1 text-[11px] text-zinc-500">
              Enter your root domain or subdomain. Subdomains like app.127.0.0.1.sslip.io are supported for testing.
            </p>
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">
              Target Server Node <span className="text-emerald-400">*</span>
            </label>
            <select
              value={serverId}
              onChange={(e) => setServerId(e.target.value)}
              className="w-full rounded-lg border border-[#262626] bg-[#181818] px-3.5 py-2.5 text-sm text-[#ededed] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            >
              {servers.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.ip_address || "No Public IP"}) - {s.region}
                </option>
              ))}
              {servers.length === 0 && (
                <option value="">Control Plane / Local Host</option>
              )}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                Target Type
              </label>
              <select
                value={targetType}
                onChange={(e) => setTargetType(e.target.value as IngressTargetType)}
                className="w-full rounded-lg border border-[#262626] bg-[#181818] px-3.5 py-2.5 text-sm text-[#ededed] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              >
                <option value="container">Container (Docker)</option>
                <option value="port">Host Port</option>
                <option value="service">Internal Service</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                Target Port <span className="text-emerald-400">*</span>
              </label>
              <input
                type="number"
                min="1"
                max="65535"
                placeholder="3000"
                value={targetPort}
                onChange={(e) => setTargetPort(Number(e.target.value))}
                className="w-full rounded-lg border border-[#262626] bg-[#181818] px-3.5 py-2.5 text-sm text-[#ededed] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                required
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">
              Container / Service Name <span className="text-emerald-400">*</span>
            </label>
            <input
              type="text"
              placeholder="e.g. frontend, web-app, backend-api"
              value={targetId}
              onChange={(e) => setTargetId(e.target.value)}
              className="w-full rounded-lg border border-[#262626] bg-[#181818] px-3.5 py-2.5 text-sm text-[#ededed] placeholder-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              required
            />
          </div>

          <div className="rounded-lg border border-[#262626] bg-[#161616] p-3.5 space-y-3">
            <label className="flex items-center justify-between cursor-pointer">
              <div className="flex items-center gap-2">
                <Shield className="h-4 w-4 text-emerald-400" />
                <div>
                  <span className="text-xs font-medium text-zinc-200">Automatic HTTPS / SSL</span>
                  <p className="text-[11px] text-zinc-500">Free TLS certificate with auto-renewal (Let's Encrypt)</p>
                </div>
              </div>
              <input
                type="checkbox"
                checked={autoSSL}
                onChange={(e) => setAutoSSL(e.target.checked)}
                className="h-4 w-4 rounded border-zinc-700 bg-zinc-900 text-emerald-500 focus:ring-emerald-500/20"
              />
            </label>

            <label className="flex items-center justify-between cursor-pointer pt-2 border-t border-[#262626]">
              <div className="flex items-center gap-2">
                <Cloud className="h-4 w-4 text-sky-400" />
                <div>
                  <span className="text-xs font-medium text-zinc-200">Cloudflare Auto-DNS</span>
                  <p className="text-[11px] text-zinc-500">Automatically create Record A in linked Cloudflare account</p>
                </div>
              </div>
              <input
                type="checkbox"
                checked={cloudflareManaged}
                onChange={(e) => setCloudflareManaged(e.target.checked)}
                className="h-4 w-4 rounded border-zinc-700 bg-zinc-900 text-sky-500 focus:ring-sky-500/20"
              />
            </label>
          </div>

          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-[#262626] bg-[#181818] px-4 py-2 text-xs font-medium text-zinc-300 hover:bg-[#222222] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex items-center gap-2 rounded-lg bg-emerald-500 px-4 py-2 text-xs font-medium text-zinc-950 hover:bg-emerald-400 disabled:opacity-50 transition-colors shadow-sm"
            >
              {submitting ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Registering...</span>
                </>
              ) : (
                <>
                  <span>Add Domain</span>
                  <ArrowRight className="h-3.5 w-3.5" />
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
